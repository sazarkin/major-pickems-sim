package main

import (
	"math"
	"math/rand"
	"slices"
	"sort"
)

type SwissSystem struct {
	Sigma           []int
	Teams           []*Team   // original team list
	records         []*Record // indexed by Team.Seed
	Faced           [][]int   // indexed by Team.Seed, holds opponent Seeds
	Remaining       []bool    // indexed by Team.Seed
	Finished        []bool    // indexed by Team.Seed
	Round           int       // current round number (1‑based)
	rng             *rand.Rand
	prob            [][]float64 // indexed by [SeedA][SeedB]
	CurrentBuchholz []int       // stores Buchholz scores for current round
}

func NewSwissSystem(teams []*Team, sigma []int, rng *rand.Rand, prob [][]float64) *SwissSystem {
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
		Sigma:           sigma,
		Teams:           teams,
		records:         records,
		Faced:           faced,
		Remaining:       remaining,
		Finished:        finished,
		Round:           0,
		rng:             rng,
		CurrentBuchholz: nil,
	}
	if prob != nil {
		ss.prob = prob
	} else {
		ss.prob = ComputeProbabilities(teams, sigma, limit)
	}
	return ss
}

func (ss *SwissSystem) Reset() {
	for _, t := range ss.Teams {
		idx := t.Seed
		ss.records[idx].Wins = 0
		ss.records[idx].Losses = 0
		ss.Faced[idx] = ss.Faced[idx][:0]
		ss.Remaining[idx] = true
		ss.Finished[idx] = false
	}
	ss.Round = 0
	ss.CurrentBuchholz = nil
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

func ComputeProbabilities(teams []*Team, sigma []int, limit int) [][]float64 {
	// compute probability matrix
	prob := make([][]float64, limit)
	for i := range prob {
		prob[i] = make([]float64, limit)
	}

	n := len(teams)
	for i := range n {
		for j := i + 1; j < n; j++ {
			tA := teams[i]
			tB := teams[j]
			p := winProb(tA, tB, sigma)
			prob[tA.Seed][tB.Seed] = p
			prob[tB.Seed][tA.Seed] = 1 - p
		}
	}

	return prob
}

// CalculateBuchholz calculates the Buchholz score for a given seed.
// Buchholz score is the sum of the score differences of all opponents faced.
func (ss *SwissSystem) CalculateBuchholz(seed int) int {
	buchholz := 0
	for _, oppSeed := range ss.Faced[seed] {
		rec := ss.records[oppSeed]
		buchholz += rec.Diff()
	}
	return buchholz
}

func (ss *SwissSystem) SimulateMatch(a, b *Team) {
	idxA := a.Seed
	idxB := b.Seed
	recA := ss.records[idxA]
	recB := ss.records[idxB]
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
			r := ss.records[idx]
			if r.Wins == 3 || r.Losses == 3 {
				ss.Remaining[idx] = false
				ss.Finished[idx] = true
			}
		}
	}
}

func (ss *SwissSystem) haveFaced(seedA, seedB int) bool {
	return slices.Contains(ss.Faced[seedA], seedB)
}

// pairingTable for Rounds 4 and 5 (6 teams).
// Indices are 0-based relative to the sorted group (0=Highest Seed, 5=Lowest Seed).
var pairingTable = [][3][2]int{
	{{0, 5}, {1, 4}, {2, 3}}, // 1v6 2v5 3v4
	{{0, 5}, {1, 3}, {2, 4}}, // 1v6 2v4 3v5
	{{0, 4}, {1, 5}, {2, 3}}, // 1v5 2v6 3v4
	{{0, 4}, {1, 3}, {2, 5}}, // 1v5 2v4 3v6
	{{0, 3}, {1, 5}, {2, 4}}, // 1v4 2v6 3v5
	{{0, 3}, {1, 4}, {2, 5}}, // 1v4 2v5 3v6
	{{0, 5}, {1, 2}, {3, 4}}, // 1v6 2v3 4v5
	{{0, 4}, {1, 2}, {3, 5}}, // 1v5 2v3 4v6
	{{0, 2}, {1, 5}, {3, 4}}, // 1v3 2v6 4v5
	{{0, 2}, {1, 4}, {3, 5}}, // 1v3 2v5 4v6
	{{0, 3}, {1, 2}, {4, 5}}, // 1v4 2v3 5v6
	{{0, 2}, {1, 3}, {4, 5}}, // 1v3 2v4 5v6
	{{0, 1}, {2, 5}, {3, 4}}, // 1v2 3v6 4v5
	{{0, 1}, {2, 4}, {3, 5}}, // 1v2 3v5 4v6
	{{0, 1}, {2, 3}, {4, 5}}, // 1v2 3v4 5v6
}

// pairGroup returns pairs (a,b) for the given group according to the
// official Swiss rules.
func (ss *SwissSystem) pairGroup(group []*Team) []pair {
	pairs := []pair{}
	if len(group) == 0 {
		return pairs
	}

	// Rules: "Match-ups shall be determined by seed."
	// Interpretation: Primary sort by Buchholz Score (Sum of opponents' score difference), Secondary by Initial Seed.
	sort.Slice(group, func(i, j int) bool {
		seedI := group[i].Seed
		seedJ := group[j].Seed

		if ss.CurrentBuchholz == nil {
			panic("CurrentBuchholz should be calculated before pairing")
		}
		// Lookup Buchholz scores from current round
		buchI := ss.CurrentBuchholz[seedI]
		buchJ := ss.CurrentBuchholz[seedJ]

		if buchI != buchJ {
			return buchI > buchJ // Higher Buchholz first
		}
		return seedI < seedJ // Lower Seed first
	})

	// Round 1: Fixed 1v9, 2v10... (i vs i + N/2)
	if ss.Round == 1 {
		half := len(group) / 2
		for i := range half {
			pairs = append(pairs, pair{group[i], group[i+half]})
		}
		return pairs
	}

	// Rounds 4 & 5: Use the priority table if we have exactly 6 teams.
	if (ss.Round == 4 || ss.Round == 5) && len(group) == 6 {
		for _, row := range pairingTable {
			valid := true
			// Check if this row results in any rematches
			for _, p := range row {
				t1 := group[p[0]]
				t2 := group[p[1]]
				if ss.haveFaced(t1.Seed, t2.Seed) {
					valid = false
					break
				}
			}
			if valid {
				for _, p := range row {
					pairs = append(pairs, pair{group[p[0]], group[p[1]]})
				}
				return pairs
			}
		}
		// If no table row is valid (unlikely), fall through to greedy High-Low.
	}

	// Rounds 2, 3, and fallback: Greedy High-Low.
	// "Highest seeded team faces the lowest seeded team available"
	used := make([]bool, len(group))
	for i := range group {
		if used[i] {
			continue
		}
		// Find opponent for group[i] (Highest available seed)
		// Look for lowest seed available (from end of list)
		found := false
		for j := len(group) - 1; j > i; j-- {
			if used[j] {
				continue
			}
			if !ss.haveFaced(group[i].Seed, group[j].Seed) {
				used[i] = true
				used[j] = true
				pairs = append(pairs, pair{group[i], group[j]})
				found = true
				break
			}
		}
		if !found {
			// If no valid opponent found without rematch, force pair with lowest available.
			// (Rules say "if possible", implying if not possible, we must pair anyway).
			for j := len(group) - 1; j > i; j-- {
				if used[j] {
					continue
				}
				used[i] = true
				used[j] = true
				pairs = append(pairs, pair{group[i], group[j]})
				found = true
				break
			}
		}
	}
	return pairs
}

type pair struct{ a, b *Team }

func (ss *SwissSystem) SimulateRound() {
	ss.Round++
	pos := []*Team{}
	even := []*Team{}
	neg := []*Team{}
	for _, t := range ss.Teams {
		if !ss.Remaining[t.Seed] {
			continue
		}
		diff := ss.records[t.Seed].Diff()
		if diff > 0 {
			pos = append(pos, t)
		} else if diff < 0 {
			neg = append(neg, t)
		} else {
			even = append(even, t)
		}
	}

	// Calculate and store Buchholz scores for this round
	roundBuchholz := make([]int, len(ss.records))
	for seed := range ss.records {
		roundBuchholz[seed] = ss.CalculateBuchholz(seed)
	}
	ss.CurrentBuchholz = roundBuchholz

	// Collect all pairs for the round first, BEFORE simulating any matches.
	// This ensures Buchholz scores are calculated based on the state at the START of the round.
	allPairs := []pair{}
	for _, group := range [][]*Team{pos, even, neg} {
		groupPairs := ss.pairGroup(group)
		allPairs = append(allPairs, groupPairs...)
	}

	// Now simulate all matches
	for _, p := range allPairs {
		ss.SimulateMatch(p.a, p.b)
	}
}

func (ss *SwissSystem) SimulateNextRound() bool {
	remaining := 0
	for _, t := range ss.Teams {
		if ss.Remaining[t.Seed] {
			remaining++
		}
	}
	if remaining == 0 {
		return false
	}
	ss.SimulateRound()
	return true
}

func (ss *SwissSystem) Records() []*Record {
	return ss.records
}

func (ss *SwissSystem) SimulateTournament() {
	for {
		if !ss.SimulateNextRound() {
			break
		}
	}
}
