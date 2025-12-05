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

const requiredCorrectPicks = 5 // by the rules, need at least 5 correct picks to be successful

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

func readTeamsData(filepath string) ([]*Team, error) {
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

	teamsData := data["teams"].(map[string]any)

	teams := make([]*Team, 0, len(teamsData))
	teamMap := make(map[string]*Team)

	for name, t := range teamsData {
		tmap := t.(map[string]any)
		seed := int(tmap["seed"].(float64))
		valveRating := int(tmap["valve"].(float64))
		team := &Team{Name: name, Seed: seed, Rating: []int{valveRating}}
		teams = append(teams, team)
		teamMap[name] = team
	}

	return teams, nil
}

// SwissSystemInterface defines the methods needed from a Swiss system
type SwissSystemInterface interface {
	Reset()
	SimulateTournament()
	Records() []*Record
}

type Simulation struct {
	Sigma   []int
	Teams   []*Team
	TeamMap map[string]*Team
	Prob    [][]float64
	baseRng *rand.Rand
	rngMu   sync.Mutex
	// swissFactory creates a new SwissSystemInterface instance
	swissFactory func(teams []*Team, sigma []int, rng *rand.Rand, prob [][]float64) SwissSystemInterface
}

func NewSimulation(valveSigma int, teams []*Team, rng *rand.Rand) (*Simulation, error) {
	return NewSimulationWithFactory(valveSigma, teams, rng, func(teams []*Team, sigma []int, rng *rand.Rand, prob [][]float64) SwissSystemInterface {
		return NewSwissSystem(teams, sigma, rng, prob)
	})
}

func NewSimulationWithFactory(valveSigma int, teams []*Team, rng *rand.Rand,
	factory func(teams []*Team, sigma []int, rng *rand.Rand, prob [][]float64) SwissSystemInterface) (*Simulation, error) {
	sigma := []int{valveSigma}
	teamMap := make(map[string]*Team)
	for _, team := range teams {
		teamMap[team.Name] = team
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

	if factory == nil {
		factory = func(teams []*Team, sigma []int, rng *rand.Rand, prob [][]float64) SwissSystemInterface {
			return NewSwissSystem(teams, sigma, rng, prob)
		}
	}

	return &Simulation{
		Sigma:        sigma,
		Teams:        teams,
		TeamMap:      teamMap,
		Prob:         prob,
		baseRng:      rng,
		swissFactory: factory,
	}, nil
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

	// Create a deterministic random source for this batch using simulation's baseRng
	sim.rngMu.Lock()
	seed := sim.baseRng.Int63()
	sim.rngMu.Unlock()
	// single rng for this batch's iterations
	rng := rand.New(rand.NewSource(seed))
	// create a single SwissSystem and reuse across iterations
	ss := sim.swissFactory(teams, sim.Sigma, rng, sim.Prob)

	for range n {
		ss.Reset()
		ss.SimulateTournament()

		var masks [3]uint64 // 0:3-0, 1:3-1 or 3-2, 2:0-3
		for _, t := range teams {
			rec := ss.Records()[t.Seed]
			if rec.Wins == 3 {
				if rec.Losses == 0 {
					masks[0] |= 1 << uint(t.Seed)
					results[t][Cat3_0]++
				} else {
					// rec.Losses == 1 or 2
					masks[1] |= 1 << uint(t.Seed)
					results[t][CatAdv]++
				}
			} else if rec.Losses == 3 && rec.Wins == 0 {
				masks[2] |= 1 << uint(t.Seed)
				results[t][Cat0_3]++
			}
		}

		for idx, pm := range predMasks {
			score := bits.OnesCount64(masks[0]&pm.perfectMask) +
				bits.OnesCount64(masks[1]&pm.advanceMask) +
				bits.OnesCount64(masks[2]&pm.zeroMask)
			if score >= requiredCorrectPicks {
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

func main() {
	var file string
	var n, k, p, s, sigma int
	var profilePath string
	var useColor bool
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -f <data.json> [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&file, "f", "", "path to input data (.json)")
	flag.IntVar(&n, "n", 1_000_000, "number of iterations to run")
	flag.IntVar(&k, "k", runtime.NumCPU(), "number of cores to use")
	flag.IntVar(&p, "p", 1_000, "number of predictions to run")
	flag.IntVar(&s, "s", 0, "random seed")
	flag.IntVar(&sigma, "sigma", 600, "sigma value for valve rating")
	flag.StringVar(&profilePath, "profile", "", "write cpu profile to file")
	flag.BoolVar(&useColor, "color", true, "enable colored output")
	flag.Parse()

	if file == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Color codes
	const (
		colorReset        = "\033[0m"
		colorRed          = "\033[31m"
		colorGreen        = "\033[32m"
		colorYellow       = "\033[33m"
		colorBlue         = "\033[34m"
		colorPurple       = "\033[35m"
		colorCyan         = "\033[36m"
		colorWhite        = "\033[37m"
		colorBrightRed    = "\033[91m"
		colorBrightGreen  = "\033[92m"
		colorBrightYellow = "\033[93m"
		colorBrightBlue   = "\033[94m"
		colorBrightPurple = "\033[95m"
		colorBrightCyan   = "\033[96m"
	)
	// All available colors - now 10 colors
	allColors := []string{
		colorRed,
		colorGreen,
		colorYellow,
		colorBlue,
		colorPurple,
		colorCyan,
		colorWhite,
		colorBrightRed,
		colorBrightGreen,
		colorBrightBlue,
	}
	// Helper function to colorize team name
	colorizeTeam := func(teamName string, seed int) string {
		if !useColor {
			return teamName
		}
		// Ensure we have a valid index
		numColors := len(allColors)
		colorIdx := (seed - 1) % numColors
		if colorIdx < 0 {
			colorIdx = 0
		}
		return allColors[colorIdx] + teamName + colorReset
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

	var masterRand *rand.Rand
	if s != 0 {
		masterRand = rand.New(rand.NewSource(int64(s)))
	} else {
		masterRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	teams, err := readTeamsData(file)
	if err != nil {
		panic(err)
	}

	sim, err := NewSimulation(sigma, teams, masterRand)
	if err != nil {
		panic(err)
	}

	teamSeeds := make([]int, 0, len(sim.Teams))
	seed2Name := make(map[int]string)
	for _, t := range sim.Teams {
		teamSeeds = append(teamSeeds, t.Seed)
		seed2Name[t.Seed] = t.Name
	}

	// Shuffle team seeds to ensure randomness in partitions
	teamSeedsShuffled := make([]int, len(teamSeeds))
	copy(teamSeedsShuffled, teamSeeds)
	masterRand.Shuffle(len(teamSeedsShuffled), func(i, j int) {
		teamSeedsShuffled[i], teamSeedsShuffled[j] = teamSeedsShuffled[j], teamSeedsShuffled[i]
	})
	// Generate all partitions of teams into categories (2,6,2,6)
	groupSizes := []int{2, 6, 2, 6}
	partitions := generateAllPartitions(teamSeedsShuffled, groupSizes, p)

	predictions := make([]map[Category][]int, 0, len(partitions))
	for _, part := range partitions {
		pred := map[Category][]int{
			Cat3_0: part[0],
			CatAdv: part[1],
			Cat0_3: part[2],
		}
		predictions = append(predictions, pred)
	}

	fmt.Printf("Simulating %d tournaments and testing %d predictions...\n", n, len(predictions))
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

	fmt.Printf("\nTop 5 Predictions:\n")
	fmt.Println("------------------")
	for idx := 0; idx < 5 && idx < len(psList); idx++ {
		ps := psList[idx]
		fmt.Printf("Success:    %.2f%%\n", ps.score)
		orderedPreds := []Category{Cat3_0, CatAdv, Cat0_3}
		for _, key := range orderedPreds {
			val := ps.pred[key]
			names := make([]string, len(val))
			for i, seed := range val {
				teamName := seed2Name[seed]
				if useColor {
					names[i] = colorizeTeam(teamName, seed)
				} else {
					names[i] = teamName
				}
			}
			keyWithDoubleColon := key.String() + ":"
			fmt.Printf("%-11s %s\n", keyWithDoubleColon, strings.Join(names, ", "))
		}
		fmt.Println()
	}
}
