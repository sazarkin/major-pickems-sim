package main

import (
	"math"
	"math/rand"
	"sort"
)

type SwissSystem struct {
	Sigma     []int
	Teams     []*Team   // original team list
	Records   []*Record // indexed by Team.Seed
	Faced     [][]int   // indexed by Team.Seed, holds opponent Seeds
	Remaining []bool    // indexed by Team.Seed
	Finished  []bool    // indexed by Team.Seed
	rng       *rand.Rand
	prob      [][]float64 // indexed by [SeedA][SeedB]
}

func NewSwissSystem(teams []*Team, sigma []int, rng *rand.Rand) *SwissSystem {
	maxSeed := 0
	for _, t := range teams {
		if t.Seed > maxSeed {
			maxSeed = t.Seed
		}
	}
	limit := maxSeed + 1

	records := make([]*Record, limit)
	faced := make([][]int, limit)
	remaining := make([]bool, limit)
	finished := make([]bool, limit)

	// Initialize arrays
	for i := range limit {
		records[i] = &Record{}
		faced[i] = make([]int, 0, 3)
	}
	// Set active teams
	for _, t := range teams {
		remaining[t.Seed] = true
	}

	ss := &SwissSystem{
		Sigma:     sigma,
		Teams:     teams,
		Records:   records,
		Faced:     faced,
		Remaining: remaining,
		Finished:  finished,
		rng:       rng,
	}
	// compute probability matrix
	prob := make([][]float64, limit)
	for i := range prob {
		prob[i] = make([]float64, limit)
	}

	n := len(teams)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			tA := teams[i]
			tB := teams[j]
			p := winProb(tA, tB, sigma)
			prob[tA.Seed][tB.Seed] = p
			prob[tB.Seed][tA.Seed] = 1 - p
		}
	}
	ss.prob = prob
	return ss
}

func (ss *SwissSystem) Reset() {
	for _, t := range ss.Teams {
		idx := t.Seed
		ss.Records[idx].Wins = 0
		ss.Records[idx].Losses = 0
		ss.Faced[idx] = ss.Faced[idx][:0]
		ss.Remaining[idx] = true
		ss.Finished[idx] = false
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
	idx := t.Seed
	diff := -ss.Records[idx].Diff()
	buch := 0
	for _, oppSeed := range ss.Faced[idx] {
		buch += ss.Records[oppSeed].Diff()
	}
	buch = -buch
	return diff, buch, t.Seed
}

func (ss *SwissSystem) SimulateMatch(a, b *Team) {
	idxA := a.Seed
	idxB := b.Seed
	recA := ss.Records[idxA]
	recB := ss.Records[idxB]
	isBO3 := recA.Wins == 2 || recA.Losses == 2

	p := ss.prob[idxA][idxB]

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
			idx := t.Seed
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
	for _, t := range ss.Teams {
		if !ss.Remaining[t.Seed] {
			continue
		}
		diff := ss.Records[t.Seed].Diff()
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
		for _, t := range ss.Teams {
			if ss.Remaining[t.Seed] {
				remaining++
			}
		}
		if remaining == 0 {
			break
		}
		ss.SimulateRound()
	}
}
