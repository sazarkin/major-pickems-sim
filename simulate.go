package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type lockedRand struct {
	sync.Mutex
	r *rand.Rand
}

func (lr *lockedRand) Float64() float64 {
	lr.Lock()
	defer lr.Unlock()
	return lr.r.Float64()
}

func (lr *lockedRand) Intn(n int) int {
	lr.Lock()
	defer lr.Unlock()
	return lr.r.Intn(n)
}

func (lr *lockedRand) Shuffle(n int, swap func(i, j int)) {
	lr.Lock()
	defer lr.Unlock()
	lr.r.Shuffle(n, swap)
}

var globalRand = &lockedRand{r: rand.New(rand.NewSource(1))}

type Team struct {
	Name   string
	Seed   int
	Rating []int
}

func (t *Team) String() string { return t.Name }

type Record struct {
	Wins   int
	Losses int
}

func (r *Record) Diff() int { return r.Wins - r.Losses }

type SwissSystem struct {
	Sigma     []int
	Records   map[*Team]*Record
	Faced     map[*Team]map[*Team]bool
	Remaining map[*Team]bool
	Finished  map[*Team]bool
}

func NewSwissSystem(teams []*Team, sigma []int) *SwissSystem {
	records := make(map[*Team]*Record)
	faced := make(map[*Team]map[*Team]bool)
	remaining := make(map[*Team]bool)
	for _, t := range teams {
		records[t] = &Record{}
		faced[t] = make(map[*Team]bool)
		remaining[t] = true
	}
	return &SwissSystem{
		Sigma:     sigma,
		Records:   records,
		Faced:     faced,
		Remaining: remaining,
		Finished:  make(map[*Team]bool),
	}
}

func (ss *SwissSystem) Reset() {
	for _, rec := range ss.Records {
		rec.Wins = 0
		rec.Losses = 0
	}
	for _, m := range ss.Faced {
		for k := range m {
			delete(m, k)
		}
	}
	ss.Remaining = make(map[*Team]bool)
	for t := range ss.Records {
		ss.Remaining[t] = true
	}
	ss.Finished = make(map[*Team]bool)
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
	diff := -ss.Records[t].Diff()
	buch := 0
	for opp := range ss.Faced[t] {
		buch += ss.Records[opp].Diff()
	}
	buch = -buch
	return diff, buch, t.Seed
}

func (ss *SwissSystem) SimulateMatch(a, b *Team) {
	recA := ss.Records[a]
	recB := ss.Records[b]
	isBO3 := recA.Wins == 2 || recA.Losses == 2

	p := winProb(a, b, ss.Sigma)

	var teamAWins bool
	if isBO3 {
		aWins, bWins := 0, 0
		for aWins < 2 && bWins < 2 {
			if globalRand.Float64() < p {
				aWins++
			} else {
				bWins++
			}
		}
		teamAWins = aWins > bWins
	} else {
		teamAWins = globalRand.Float64() < p
	}

	if teamAWins {
		recA.Wins++
		recB.Losses++
	} else {
		recA.Losses++
		recB.Wins++
	}

	ss.Faced[a][b] = true
	ss.Faced[b][a] = true

	if isBO3 {
		for _, t := range []*Team{a, b} {
			r := ss.Records[t]
			if r.Wins == 3 || r.Losses == 3 {
				delete(ss.Remaining, t)
				ss.Finished[t] = true
			}
		}
	}
}

func (ss *SwissSystem) SimulateRound() {
	pos := []*Team{}
	even := []*Team{}
	neg := []*Team{}
	for t := range ss.Remaining {
		diff := ss.Records[t].Diff()
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
	if len(even) == len(ss.Records) {
		half := len(even) / 2
		secondHalf := even[half:]
		for i := 0; i < len(secondHalf)/2; i++ {
			j := len(secondHalf) - 1 - i
			secondHalf[i], secondHalf[j] = secondHalf[j], secondHalf[i]
		}
	}

	for _, group := range [][]*Team{pos, even, neg} {
		half := len(group) / 2
		for i := 0; i < half; i++ {
			a := group[i]
			b := group[len(group)-1-i]
			ss.SimulateMatch(a, b)
		}
	}
}

func (ss *SwissSystem) SimulateTournament() {
	for len(ss.Remaining) > 0 {
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

	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	sigmaMap := data["sigma"].(map[string]interface{})
	sigmaKeys := make([]string, 0, len(sigmaMap))
	for k := range sigmaMap {
		sigmaKeys = append(sigmaKeys, k)
	}
	sort.Strings(sigmaKeys)
	sigma := make([]int, len(sigmaKeys))
	for i, k := range sigmaKeys {
		sigma[i] = int(sigmaMap[k].(float64))
	}

	teamsData := data["teams"].(map[string]interface{})

	teams := make([]*Team, 0, len(teamsData))
	teamMap := make(map[string]*Team)

	for name, t := range teamsData {
		tmap := t.(map[string]interface{})
		seed := int(tmap["seed"].(float64))
		ratings := make([]int, len(sigmaKeys))
		for i, sysKey := range sigmaKeys {
			ratingVal := tmap[sysKey].(float64)
			ratings[i] = int(ratingVal)
		}
		team := &Team{Name: name, Seed: seed, Rating: ratings}
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
	type predSet struct {
		perfect map[string]bool
		advance map[string]bool
		zero    map[string]bool
	}
	predSets := make([]predSet, len(predictions))
	for i, p := range predictions {
		ps := predSet{
			perfect: make(map[string]bool),
			advance: make(map[string]bool),
			zero:    make(map[string]bool),
		}
		for _, t := range p["3-0"] {
			ps.perfect[t] = true
		}
		for _, t := range p["3-1 or 3-2"] {
			ps.advance[t] = true
		}
		for _, t := range p["0-3"] {
			ps.zero[t] = true
		}
		predSets[i] = ps
	}

	teams := sim.Teams
	ss := NewSwissSystem(teams, sim.Sigma)

	results := make(map[*Team]map[string]int)
	for _, t := range teams {
		results[t] = map[string]int{"3-0": 0, "3-1 or 3-2": 0, "0-3": 0}
	}
	success := make([]int, len(predictions))

	for iter := 0; iter < n; iter++ {
		ss.Reset()
		ss.SimulateTournament()

		outcomeGroups := map[string]map[string]bool{
			"3-0":        make(map[string]bool),
			"3-1 or 3-2": make(map[string]bool),
			"0-3":        make(map[string]bool),
		}
		for _, t := range teams {
			rec := ss.Records[t]
			if rec.Wins == 3 {
				if rec.Losses == 0 {
					results[t]["3-0"]++
					outcomeGroups["3-0"][t.Name] = true
				} else {
					results[t]["3-1 or 3-2"]++
					outcomeGroups["3-1 or 3-2"][t.Name] = true
				}
			} else if rec.Losses == 3 {
				results[t]["0-3"]++
				outcomeGroups["0-3"][t.Name] = true
			}
		}
		for idx, ps := range predSets {
			score := 0
			for tn := range outcomeGroups["3-0"] {
				if ps.perfect[tn] {
					score++
				}
			}
			for tn := range outcomeGroups["3-1 or 3-2"] {
				if ps.advance[tn] {
					score++
				}
			}
			for tn := range outcomeGroups["0-3"] {
				if ps.zero[tn] {
					score++
				}
			}
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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -f <data.json> [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&file, "f", "", "path to input data (.json)")
	flag.IntVar(&n, "n", 1_000_000, "number of iterations to run")
	flag.IntVar(&k, "k", runtime.NumCPU(), "number of cores to use")
	flag.IntVar(&p, "p", 1_000, "number of predictions to run")
	flag.IntVar(&s, "s", 0, "random seed")
	flag.Parse()

	if file == "" {
		flag.Usage()
		os.Exit(1)
	}

	if s != 0 {
		globalRand.r = rand.New(rand.NewSource(int64(s)))
	} else {
		globalRand.r = rand.New(rand.NewSource(time.Now().UnixNano()))
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
			globalRand.Shuffle(len(shuffled), func(i, j int) {
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
