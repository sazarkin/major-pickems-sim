package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"
	"time"
)

type Team struct {
	Name   string
	Seed   int
	Rating []int
	Index  int
}

func (t *Team) String() string { return t.Name }

type Record struct {
	Wins   int
	Losses int
}

func (r *Record) Diff() int { return r.Wins - r.Losses }

type SwissSystem struct {
	Sigma     []int
	Teams     []*Team   // original team list in index order
	Records   []*Record // indexed by Team.Index
	Faced     [][]int   // indexed by Team.Index, holds opponent indices
	Remaining []bool    // indexed by Team.Index
	Finished  []bool    // indexed by Team.Index
	rng       *rand.Rand
}

func NewSwissSystem(teams []*Team, sigma []int, rng *rand.Rand) *SwissSystem {
	n := len(teams)
	records := make([]*Record, n)
	faced := make([][]int, n)
	remaining := make([]bool, n)
	finished := make([]bool, n)
	for i := 0; i < n; i++ {
		records[i] = &Record{}
		faced[i] = make([]int, 0, 3)
		remaining[i] = true
	}
	return &SwissSystem{
		Sigma:     sigma,
		Teams:     teams,
		Records:   records,
		Faced:     faced,
		Remaining: remaining,
		Finished:  finished,
		rng:       rng,
	}
}

func (ss *SwissSystem) Reset() {
	for i := range ss.Records {
		ss.Records[i].Wins = 0
		ss.Records[i].Losses = 0
		ss.Faced[i] = ss.Faced[i][:0]
		ss.Remaining[i] = true
		ss.Finished[i] = false
	}
}

func winProb(a, b *Team, sigma []int) float64 {
	probs := make([]float64, len(sigma))
	for i, s := range sigma {
		diff := float64(b.Rating[i] - a.Rating[i])
		divisor := float64(2 * s)
		probs[i] = 1.0 / (1.0 + math.Pow(10.0, diff/divisor))
	}
	sort.Float64s(probs)
	mid := len(probs) / 2
	if len(probs)%2 == 0 {
		return (probs[mid-1] + probs[mid]) / 2
	}
	return probs[mid]
}

func (ss *SwissSystem) seeding(t *Team) (int, int, int) {
	idx := t.Index
	diff := -ss.Records[idx].Diff()
	buch := 0
	for _, oppIdx := range ss.Faced[idx] {
		buch += ss.Records[oppIdx].Diff()
	}
	buch = -buch
	return diff, buch, t.Seed
}

func (ss *SwissSystem) SimulateMatch(a, b *Team) {
	idxA := a.Index
	idxB := b.Index
	recA := ss.Records[idxA]
	recB := ss.Records[idxB]
	isBO3 := recA.Wins == 2 || recA.Losses == 2

	p := winProb(a, b, ss.Sigma)

	var teamAWins bool
	if isBO3 {
		aWins, bWins := 0, 0
		for aWins < 2 && bWins < 2 {
			if ss.rng.Float64() < p {
				aWins++
			} else {
				bWins++
			}
		}
		teamAWins = aWins > bWins
	} else {
		teamAWins = ss.rng.Float64() < p
	}

	if teamAWins {
		recA.Wins++
		recB.Losses++
	} else {
		recA.Losses++
		recB.Wins++
	}

	ss.Faced[idxA] = append(ss.Faced[idxA], idxB)
	ss.Faced[idxB] = append(ss.Faced[idxB], idxA)

	if isBO3 {
		for _, t := range []*Team{a, b} {
			idx := t.Index
			r := ss.Records[idx]
			if r.Wins == 3 || r.Losses == 3 {
				ss.Remaining[idx] = false
				ss.Finished[idx] = true
			}
		}
	}
}

func (ss *SwissSystem) SimulateRound() {
	pos := []*Team{}
	even := []*Team{}
	neg := []*Team{}
	for i, t := range ss.Teams {
		if !ss.Remaining[i] {
			continue
		}
		diff := ss.Records[i].Diff()
		if diff > 0 {
			pos = append(pos, t)
		} else if diff < 0 {
			neg = append(neg, t)
		} else {
			even = append(even, t)
		}
	}

	sortGroup := func(group []*Team) {
		sort.Slice(group, func(i, j int) bool {
			di, bi, si := ss.seeding(group[i])
			dj, bj, sj := ss.seeding(group[j])
			if di != dj {
				return di < dj
			}
			if bi != bj {
				return bi < bj
			}
			return si < sj
		})
	}
	sortGroup(pos)
	sortGroup(even)
	sortGroup(neg)

	// special first round handling (seed 1-9,2-10,...)
	if len(even) == len(ss.Teams) {
		half := len(even) / 2
		secondHalf := even[half:]
		for i := 0; i < len(secondHalf)/2; i++ {
			j := len(secondHalf) - 1 - i
			secondHalf[i], secondHalf[j] = secondHalf[j], secondHalf[i]
		}
	}

	for _, group := range [][]*Team{pos, even, neg} {
		half := len(group) / 2
		for i := range half {
			a := group[i]
			b := group[len(group)-1-i]
			ss.SimulateMatch(a, b)
		}
	}
}

func (ss *SwissSystem) SimulateTournament() {
	for {
		remaining := 0
		for i := range ss.Teams {
			if ss.Remaining[i] {
				remaining++
			}
		}
		if remaining == 0 {
			break
		}
		ss.SimulateRound()
	}
}

type Simulation struct {
	Sigma   []int
	Teams   []*Team
	TeamMap map[string]*Team
}

func NewSimulation(filepath string) (*Simulation, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data map[string]any
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	sigmaMap := data["sigma"].(map[string]any)
	sigmaKeys := make([]string, 0, len(sigmaMap))
	for k := range sigmaMap {
		sigmaKeys = append(sigmaKeys, k)
	}
	sort.Strings(sigmaKeys)
	sigma := make([]int, len(sigmaKeys))
	for i, k := range sigmaKeys {
		sigma[i] = int(sigmaMap[k].(float64))
	}

	teamsData := data["teams"].(map[string]any)

	teams := make([]*Team, 0, len(teamsData))
	teamMap := make(map[string]*Team)

	for name, t := range teamsData {
		tmap := t.(map[string]any)
		seed := int(tmap["seed"].(float64))
		ratings := make([]int, len(sigmaKeys))
		for i, sysKey := range sigmaKeys {
			ratingVal := tmap[sysKey].(float64)
			ratings[i] = int(ratingVal)
		}
		idx := len(teams)
		team := &Team{Name: name, Seed: seed, Rating: ratings, Index: idx}
		teams = append(teams, team)
		teamMap[name] = team
	}

	return &Simulation{Sigma: sigma, Teams: teams, TeamMap: teamMap}, nil
}

type BatchResult struct {
	Results map[*Team]map[string]int
	Success []int
}

func (sim *Simulation) Batch(n int, predictions []map[string][]string) (*BatchResult, error) {
	teams := sim.Teams

	// map team name to index (0..len(teams)-1)
	name2idx := make(map[string]int, len(teams))
	for idx, t := range teams {
		name2idx[t.Name] = idx
	}

	type predMask struct {
		perfectMask uint32
		advanceMask uint32
		zeroMask    uint32
	}
	predMasks := make([]predMask, len(predictions))
	for i, p := range predictions {
		var perfectMask, advanceMask, zeroMask uint32
		for _, tn := range p["3-0"] {
			idx := name2idx[tn]
			perfectMask |= 1 << uint(idx)
		}
		for _, tn := range p["3-1 or 3-2"] {
			idx := name2idx[tn]
			advanceMask |= 1 << uint(idx)
		}
		for _, tn := range p["0-3"] {
			idx := name2idx[tn]
			zeroMask |= 1 << uint(idx)
		}
		predMasks[i] = predMask{
			perfectMask: perfectMask,
			advanceMask: advanceMask,
			zeroMask:    zeroMask,
		}
	}

	results := make(map[*Team]map[string]int)
	for _, t := range teams {
		results[t] = map[string]int{"3-0": 0, "3-1 or 3-2": 0, "0-3": 0}
	}
	success := make([]int, len(predictions))

	// Create a local random source for this batch
	localRand := rand.New(rand.NewSource(time.Now().UnixNano() + int64(n)))

	for range n {
		// Create a new random source for each iteration using the local random source to seed it
		seed := localRand.Int63()
		rng := rand.New(rand.NewSource(seed))
		ss := NewSwissSystem(teams, sim.Sigma, rng)
		ss.SimulateTournament()

		var masks [3]uint32 // 0:3-0, 1:3-1 or 3-2, 2:0-3
		for idx := range teams {
			rec := ss.Records[idx]
			if rec.Wins == 3 {
				if rec.Losses == 0 {
					masks[0] |= 1 << uint(idx)
				} else {
					masks[1] |= 1 << uint(idx)
				}
			} else if rec.Losses == 3 {
				masks[2] |= 1 << uint(idx)
			}
		}

		for idx, pm := range predMasks {
			score := bits.OnesCount32(masks[0]&pm.perfectMask) +
				bits.OnesCount32(masks[1]&pm.advanceMask) +
				bits.OnesCount32(masks[2]&pm.zeroMask)
			if score >= 6 {
				success[idx]++
			}
		}
	}
	return &BatchResult{Results: results, Success: success}, nil
}

func (sim *Simulation) Run(n, k int, predictions []map[string][]string) (map[*Team]map[string]int, []float64) {
	batchSize := n / k
	remainder := n % k

	var wg sync.WaitGroup
	mu := sync.Mutex{}

	combinedResults := make(map[*Team]map[string]int)
	for _, t := range sim.Teams {
		combinedResults[t] = map[string]int{"3-0": 0, "3-1 or 3-2": 0, "0-3": 0}
	}
	combinedSuccess := make([]int, len(predictions))

	for i := 0; i < k; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			size := batchSize
			if idx < remainder {
				size++
			}
			batchRes, err := sim.Batch(size, predictions)
			if err != nil {
				return
			}
			mu.Lock()
			for t, resMap := range batchRes.Results {
				for key, val := range resMap {
					combinedResults[t][key] += val
				}
			}
			for j, val := range batchRes.Success {
				combinedSuccess[j] += val
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	percentages := make([]float64, len(predictions))
	for i, succ := range combinedSuccess {
		percentages[i] = float64(succ) / float64(n) * 100.0
	}
	return combinedResults, percentages
}

func hashPrediction(pred map[string][]string) string {
	g1 := append([]string{}, pred["3-0"]...)
	sort.Strings(g1)
	g2 := append([]string{}, pred["3-1 or 3-2"]...)
	sort.Strings(g2)
	g3 := append([]string{}, pred["0-3"]...)
	sort.Strings(g3)
	var concat []byte
	for _, s := range g1 {
		concat = append(concat, s...)
	}
	for _, s := range g2 {
		concat = append(concat, s...)
	}
	for _, s := range g3 {
		concat = append(concat, s...)
	}
	hash := md5.Sum(concat)
	return hex.EncodeToString(hash[:])
}

func main() {
	var file string
	var n, k, p, s int
	var profilePath string
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -f <data.json> [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&file, "f", "", "path to input data (.json)")
	flag.IntVar(&n, "n", 1_000_000, "number of iterations to run")
	flag.IntVar(&k, "k", runtime.NumCPU(), "number of cores to use")
	flag.IntVar(&p, "p", 1_000, "number of predictions to run")
	flag.IntVar(&s, "s", 0, "random seed")
	flag.StringVar(&profilePath, "profile", "", "write cpu profile to file")
	flag.Parse()

	if file == "" {
		flag.Usage()
		os.Exit(1)
	}

	// CPU profiling
	if profilePath != "" {
		f, err := os.Create(profilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	// Set up a master random source for generating seeds
	var masterRand *rand.Rand
	if s != 0 {
		masterRand = rand.New(rand.NewSource(int64(s)))
	} else {
		masterRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	sim, err := NewSimulation(file)
	if err != nil {
		panic(err)
	}

	teamNames := make([]string, 0, len(sim.Teams))
	for _, t := range sim.Teams {
		teamNames = append(teamNames, t.Name)
	}

	predictions := make([]map[string][]string, 0, p)
	hashes := make(map[string]bool)

	for i := 0; i < p; i++ {
		for {
			shuffled := make([]string, len(teamNames))
			copy(shuffled, teamNames)
			// Use masterRand to shuffle
			masterRand.Shuffle(len(shuffled), func(i, j int) {
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			})
			pred := map[string][]string{
				"3-0":        shuffled[:2],
				"3-1 or 3-2": shuffled[2:8],
				"0-3":        shuffled[8:10],
			}
			h := hashPrediction(pred)
			if !hashes[h] {
				hashes[h] = true
				predictions = append(predictions, pred)
				break
			}
		}
	}

	start := time.Now()
	results, scores := sim.Run(n, k, predictions)
	_ = results // keep for potential future use
	elapsed := time.Since(start).Seconds()
	_ = elapsed // keep for potential future use

	type predScore struct {
		score float64
		pred  map[string][]string
	}
	psList := make([]predScore, len(predictions))
	for i, pred := range predictions {
		psList[i] = predScore{score: scores[i], pred: pred}
	}
	sort.Slice(psList, func(i, j int) bool {
		return psList[i].score > psList[j].score
	})

	for idx := 0; idx < 5 && idx < len(psList); idx++ {
		ps := psList[idx]
		fmt.Printf("Percent of success: %.2f%%\n", ps.score)
		h := hashPrediction(ps.pred)
		if len(h) >= 5 {
			fmt.Printf("%s\n", h[len(h)-5:])
		} else {
			fmt.Printf("%s\n", h)
		}
		for key, val := range ps.pred {
			fmt.Printf("'%s': %s\n", key, strings.Join(val, ", "))
		}
		fmt.Println()
	}
}
