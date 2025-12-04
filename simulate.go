package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

type Category int

const (
	Cat3_0 Category = iota
	CatAdv
	Cat0_3
)

func (c Category) String() string {
	switch c {
	case Cat3_0:
		return "3-0"
	case CatAdv:
		return "3-1 or 3-2"
	case Cat0_3:
		return "0-3"
	}
	return "unknown"
}

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

type Simulation struct {
	Sigma   []int
	Teams   []*Team
	TeamMap map[string]*Team
	Prob    [][]float64
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
		team := &Team{Name: name, Seed: seed, Rating: ratings}
		teams = append(teams, team)
		teamMap[name] = team
	}

	// Sort teams by seed for deterministic order
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Seed < teams[j].Seed
	})

	// Precompute probability matrix
	// Find max seed
	maxSeed := 0
	for _, t := range teams {
		if t.Seed > maxSeed {
			maxSeed = t.Seed
		}
	}
	limit := maxSeed + 1
	prob := ComputeProbabilities(teams, sigma, limit)

	return &Simulation{Sigma: sigma, Teams: teams, TeamMap: teamMap, Prob: prob}, nil
}

type BatchResult struct {
	Results map[*Team]map[Category]int
	Success []int
}

func (sim *Simulation) Batch(n int, predictions []map[Category][]int) (*BatchResult, error) {
	teams := sim.Teams

	type predMask struct {
		perfectMask uint64
		advanceMask uint64
		zeroMask    uint64
	}
	predMasks := make([]predMask, len(predictions))
	for i, p := range predictions {
		var perfectMask, advanceMask, zeroMask uint64
		for _, seed := range p[Cat3_0] {
			perfectMask |= 1 << uint(seed)
		}
		for _, seed := range p[CatAdv] {
			advanceMask |= 1 << uint(seed)
		}
		for _, seed := range p[Cat0_3] {
			zeroMask |= 1 << uint(seed)
		}
		predMasks[i] = predMask{
			perfectMask: perfectMask,
			advanceMask: advanceMask,
			zeroMask:    zeroMask,
		}
	}

	results := make(map[*Team]map[Category]int)
	for _, t := range teams {
		results[t] = map[Category]int{Cat3_0: 0, CatAdv: 0, Cat0_3: 0}
	}
	success := make([]int, len(predictions))

	// Create a local random source for this batch
	localRand := rand.New(rand.NewSource(time.Now().UnixNano() + int64(n)))
	// single rng for this batch's iterations
	rng := rand.New(rand.NewSource(localRand.Int63()))
	// create a single SwissSystem and reuse across iterations
	ss := NewSwissSystem(teams, sim.Sigma, rng, sim.Prob)

	for range n {
		ss.Reset()
		ss.SimulateTournament()

		var masks [3]uint64 // 0:3-0, 1:3-1 or 3-2, 2:0-3
		for _, t := range teams {
			rec := ss.Records[t.Seed]
			if rec.Wins == 3 {
				if rec.Losses == 0 {
					masks[0] |= 1 << uint(t.Seed)
				} else {
					masks[1] |= 1 << uint(t.Seed)
				}
			} else if rec.Losses == 3 {
				masks[2] |= 1 << uint(t.Seed)
			}
		}

		for idx, pm := range predMasks {
			score := bits.OnesCount64(masks[0]&pm.perfectMask) +
				bits.OnesCount64(masks[1]&pm.advanceMask) +
				bits.OnesCount64(masks[2]&pm.zeroMask)
			if score >= 6 {
				success[idx]++
			}
		}
	}
	return &BatchResult{Results: results, Success: success}, nil
}

func (sim *Simulation) Run(n, k int, predictions []map[Category][]int) (map[*Team]map[Category]int, []float64) {
	batchSize := n / k
	remainder := n % k

	var wg sync.WaitGroup
	mu := sync.Mutex{}

	combinedResults := make(map[*Team]map[Category]int)
	for _, t := range sim.Teams {
		combinedResults[t] = map[Category]int{Cat3_0: 0, CatAdv: 0, Cat0_3: 0}
	}
	combinedSuccess := make([]int, len(predictions))

	for i := range k {
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

func hashPrediction(pred map[Category][]int) uint64 {
	g1 := append([]int{}, pred[Cat3_0]...)
	sort.Ints(g1)
	g2 := append([]int{}, pred[CatAdv]...)
	sort.Ints(g2)
	g3 := append([]int{}, pred[Cat0_3]...)
	sort.Ints(g3)
	// Use a fixed-size buffer for hashing
	var h uint64
	for i, s := range g1 {
		h ^= uint64(s) << (i * 4)
	}
	for i, s := range g2 {
		h ^= uint64(s) << ((i + 2) * 4)
	}
	for i, s := range g3 {
		h ^= uint64(s) << ((i + 8) * 4)
	}
	// Mix the bits for better distribution
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33
	return h
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

	teamSeeds := make([]int, 0, len(sim.Teams))
	seed2Name := make(map[int]string)
	for _, t := range sim.Teams {
		teamSeeds = append(teamSeeds, t.Seed)
		seed2Name[t.Seed] = t.Name
	}

	predictions := make([]map[Category][]int, 0, p)
	hashes := make(map[uint64]bool)

	for i := 0; i < p; i++ {
		for {
			shuffled := make([]int, len(teamSeeds))
			copy(shuffled, teamSeeds)
			// Use masterRand to shuffle
			masterRand.Shuffle(len(shuffled), func(i, j int) {
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			})
			pred := map[Category][]int{
				Cat3_0: shuffled[:2],
				CatAdv: shuffled[2:8],
				Cat0_3: shuffled[8:10],
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
		pred  map[Category][]int
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
		hStr := fmt.Sprintf("%x", h)
		if len(hStr) >= 5 {
			fmt.Printf("%s\n", hStr[len(hStr)-5:])
		} else {
			fmt.Printf("%s\n", hStr)
		}
		orderedPreds := []Category{Cat3_0, CatAdv, Cat0_3}
		for _, key := range orderedPreds {
			val := ps.pred[key]
			names := make([]string, len(val))
			for i, seed := range val {
				names[i] = seed2Name[seed]
			}
			fmt.Printf("'%s': %s\n", key, strings.Join(names, ", "))
		}
		fmt.Println()
	}
}
