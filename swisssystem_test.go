package main

import (
	"math/rand"
	"testing"
)

// TestBudapest2025Stage2 replicates the exact Swiss bracket from the
// Budapest 2025 Stage 2 tournament using fixed probabilities.
func TestBudapest2025Stage1(t *testing.T) {
	// Teams with their actual seeds from Budapest 2025 Stage 1
	teams := []*Team{
		{Name: "Legacy", Seed: 1, Rating: []int{1500}},
		{Name: "FaZe Clan", Seed: 2, Rating: []int{1500}},
		{Name: "B8", Seed: 3, Rating: []int{1500}},
		{Name: "GamerLegion", Seed: 4, Rating: []int{1500}},
		{Name: "Fnatic", Seed: 5, Rating: []int{1500}},
		{Name: "PARIVISION", Seed: 6, Rating: []int{1500}},
		{Name: "Ninjas in Pyjamas", Seed: 7, Rating: []int{1500}},
		{Name: "Imperial Esports", Seed: 8, Rating: []int{1500}},
		{Name: "FlyQuest", Seed: 9, Rating: []int{1500}},
		{Name: "Lynn Vision Gaming", Seed: 10, Rating: []int{1500}},
		{Name: "M80", Seed: 11, Rating: []int{1500}},
		{Name: "Fluxo", Seed: 12, Rating: []int{1500}},
		{Name: "RED Canids", Seed: 13, Rating: []int{1500}},
		{Name: "The Huns Esports", Seed: 14, Rating: []int{1500}},
		{Name: "NRG", Seed: 15, Rating: []int{1500}},
		{Name: "Rare Atom", Seed: 16, Rating: []int{1500}},
	}

	prob := make([][]float64, 17)
	for i := range prob {
		prob[i] = make([]float64, 17)
	}
	setProbBySeed := func(seedA, seedB int, p float64) {
		prob[seedA][seedB] = p
		prob[seedB][seedA] = 1 - p
	}

	// Set probabilities to match actual results
	// Round 1
	setProbBySeed(1, 9, 0.0)  // Legacy loses to FlyQuest (actual: 10-13)
	setProbBySeed(2, 10, 1.0) // FaZe beats Lynn Vision
	setProbBySeed(3, 11, 0.0) // B8 loses to M80
	setProbBySeed(4, 12, 0.0) // GamerLegion loses to Fluxo
	setProbBySeed(5, 13, 1.0) // Fnatic beats RED Canids
	setProbBySeed(6, 14, 1.0) // PARIVISION beats The Huns
	setProbBySeed(7, 15, 0.0) // NiP loses to NRG
	setProbBySeed(8, 16, 1.0) // Imperial beats Rare Atom

	// Round 2
	setProbBySeed(2, 15, 0.0) // FaZe loses to NRG
	setProbBySeed(5, 12, 0.0) // Fnatic loses to Fluxo
	setProbBySeed(6, 11, 0.0) // PARIVISION loses to M80
	setProbBySeed(8, 9, 0.0)  // Imperial loses to FlyQuest
	setProbBySeed(1, 16, 1.0) // Legacy beats Rare Atom
	setProbBySeed(3, 14, 1.0) // B8 beats The Huns
	setProbBySeed(4, 13, 0.0) // GamerLegion loses to RED Canids
	setProbBySeed(7, 10, 1.0) // NiP beats Lynn Vision

	// Round 3
	// 2-0 pool (Bo3)
	setProbBySeed(9, 12, 1.0)  // FlyQuest beats Fluxo
	setProbBySeed(11, 15, 1.0) // M80 beats NRG
	// 1-1 pool (Bo1)
	setProbBySeed(5, 8, 1.0)  // Fnatic beats Imperial
	setProbBySeed(1, 13, 1.0) // Legacy beats RED Canids
	setProbBySeed(2, 7, 0.0)  // FaZe loses to NiP
	setProbBySeed(3, 6, 1.0)  // B8 beats PARIVISION
	// 0-2 pool (Bo3)
	setProbBySeed(4, 16, 1.0)  // GamerLegion beats Rare Atom
	setProbBySeed(10, 14, 0.0) // Lynn Vision loses to The Huns

	// Round 4
	// 3-0 Qualified: FlyQuest, M80 are already qualified
	// 2-1 pool (Bo3)
	setProbBySeed(12, 7, 0.0) // Fluxo loses to NiP
	setProbBySeed(15, 5, 0.0) // NRG loses to Fnatic
	setProbBySeed(3, 1, 1.0)  // B8 beats Legacy
	// 1-2 pool (Bo3)
	setProbBySeed(6, 4, 1.0)  // PARIVISION beats GamerLegion
	setProbBySeed(8, 14, 1.0) // Imperial beats The Huns
	setProbBySeed(13, 2, 0.0) // RED Canids loses to FaZe

	// Round 5
	// 3-1 to qualify: B8, Fnatic, NiP are already qualified
	// 2-2 pool (Bo3) - Qualification matches
	setProbBySeed(15, 8, 0.0) // NRG loses to Imperial
	setProbBySeed(12, 2, 0.0) // Fluxo loses to FaZe
	setProbBySeed(6, 1, 1.0)  // PARIVISION beats Legacy (to match final standings: Legacy 2-3, PARIVISION 3-2)

	sigma := []int{200}
	rng := rand.New(rand.NewSource(42))
	ss := NewSwissSystem(teams, sigma, rng, prob)

	// Run the tournament
	ss.SimulateTournament()

	// Check final records
	expectedWins := map[string]int{
		"M80":                3,
		"FlyQuest":           3,
		"B8":                 3,
		"Fnatic":             3,
		"Ninjas in Pyjamas":  3,
		"PARIVISION":         3,
		"Imperial Esports":   3,
		"FaZe Clan":          3,
		"NRG":                2,
		"Fluxo":              2,
		"Legacy":             2,
		"The Huns Esports":   1,
		"RED Canids":         1,
		"GamerLegion":        1,
		"Lynn Vision Gaming": 0,
		"Rare Atom":          0,
	}

	for _, team := range teams {
		rec := ss.Records[team.Seed]
		expected := expectedWins[team.Name]
		if rec.Wins != expected {
			t.Errorf("Team %s: expected %d wins, got %d (losses: %d)", team.Name, expected, rec.Wins, rec.Losses)
		}
	}
}

func TestBudapest2025Stage2(t *testing.T) {
	// Teams with their actual seeds (names for reference only)
	teams := []*Team{
		{Name: "Aurora Gaming", Seed: 1, Rating: []int{1500}},
		{Name: "Natus Vincere", Seed: 2, Rating: []int{1500}},
		{Name: "Team Liquid", Seed: 3, Rating: []int{1500}},
		{Name: "3DMAX", Seed: 4, Rating: []int{1500}},
		{Name: "Astralis", Seed: 5, Rating: []int{1500}},
		{Name: "TYLOO", Seed: 6, Rating: []int{1500}},
		{Name: "MIBR", Seed: 7, Rating: []int{1500}},
		{Name: "Passion UA", Seed: 8, Rating: []int{1500}},
		{Name: "M80", Seed: 9, Rating: []int{1500}},
		{Name: "FlyQuest", Seed: 10, Rating: []int{1500}},
		{Name: "B8", Seed: 11, Rating: []int{1500}},
		{Name: "Fnatic", Seed: 12, Rating: []int{1500}},
		{Name: "Ninjas in Pyjamas", Seed: 13, Rating: []int{1500}},
		{Name: "PARIVISION", Seed: 14, Rating: []int{1500}},
		{Name: "Imperial Esports", Seed: 15, Rating: []int{1500}},
		{Name: "FaZe Clan", Seed: 16, Rating: []int{1500}},
	}

	prob := make([][]float64, 17)
	for i := range prob {
		prob[i] = make([]float64, 17)
	}
	setProbBySeed := func(seedA, seedB int, p float64) {
		prob[seedA][seedB] = p
		prob[seedB][seedA] = 1 - p
	}

	// Set deterministic probabilities for every match that occurred in the real tournament.
	// Each entry: (winnerSeed, loserSeed, 1.0)
	// We list matches in chronological order (round by round) as extracted from the HTML.
	matches := []struct {
		winner int
		loser  int
	}{
		// Round 1 (0-0 Bo1)
		{1, 9},  // Aurora > M80
		{2, 10}, // NaVi > FlyQuest
		{11, 3}, // B8 > Liquid
		{12, 4}, // Fnatic > 3DMAX
		{13, 5}, // NiP > Astralis
		{6, 14}, // TYLOO > PARIVISION
		{15, 7}, // Imperial > MIBR
		{16, 8}, // FaZe > Passion UA

		// Round 2 (1-0 Bo1)
		{16, 1},  // FaZe > Aurora
		{2, 15},  // NaVi > Imperial
		{13, 6},  // NiP > TYLOO
		{11, 12}, // B8 > Fnatic

		// Round 2 (0-1 Bo1)
		{14, 3}, // PARIVISION > Liquid
		{4, 10}, // 3DMAX > FlyQuest
		{9, 5},  // M80 > Astralis
		{8, 7},  // Passion UA > MIBR

		// Round 3 (2-0 Bo3)
		{16, 13}, // FaZe > NiP
		{2, 11},  // NaVi > B8

		// Round 3 (1-1 Bo1)
		{14, 1},  // PARIVISION > Aurora
		{9, 6},   // M80 > TYLOO
		{15, 12}, // Imperial > Fnatic
		{4, 8},   // 3DMAX > Passion UA

		// Round 3 (0-2 Bo3)
		{3, 7},  // Liquid > MIBR
		{5, 10}, // Astralis > FlyQuest

		// Round 4 (2-1 Bo3)
		{11, 4},  // B8 > 3DMAX
		{14, 13}, // PARIVISION > NiP
		{15, 9},  // Imperial > M80

		// Round 4 (1-2 Bo3)
		{5, 1},  // Astralis > Aurora
		{3, 6},  // Liquid > TYLOO
		{8, 12}, // Passion UA > Fnatic

		// Round 5 (2-2 Bo3)
		{4, 13}, // 3DMAX > NiP
		{3, 5},  // Liquid > Astralis
		{8, 9},  // Passion UA > M80
	}

	for _, m := range matches {
		setProbBySeed(m.winner, m.loser, 1.0)
	}

	sigma := []int{100}
	rng := rand.New(rand.NewSource(42))
	ss := NewSwissSystem(teams, sigma, rng, prob)
	ss.SimulateTournament()

	// Expected final records (wins, losses) for each seed
	expected := map[int]struct{ wins, losses int }{
		1:  {1, 3}, // Aurora
		2:  {3, 0}, // NaVi
		3:  {3, 2}, // Liquid
		4:  {3, 2}, // 3DMAX
		5:  {2, 3}, // Astralis
		6:  {1, 3}, // TYLOO
		7:  {0, 3}, // MIBR
		8:  {3, 2}, // Passion UA
		9:  {2, 3}, // M80
		10: {0, 3}, // FlyQuest
		11: {3, 1}, // B8
		12: {1, 3}, // Fnatic
		13: {2, 3}, // NiP
		14: {3, 1}, // PARIVISION
		15: {3, 1}, // Imperial
		16: {3, 0}, // FaZe
	}

	for seed, exp := range expected {
		rec := ss.Records[seed]
		if rec.Wins != exp.wins || rec.Losses != exp.losses {
			t.Errorf("seed %d record = %d-%d, want %d-%d", seed, rec.Wins, rec.Losses, exp.wins, exp.losses)
		}
	}

	// Additionally, verify that each team faced the correct opponents in each round.
	// We'll just check total number of opponents per seed (should be equal to wins+losses)
	for seed, exp := range expected {
		faced := ss.Faced[seed]
		if len(faced) != exp.wins+exp.losses {
			t.Errorf("seed %d faced %d opponents, expected %d", seed, len(faced), exp.wins+exp.losses)
		}
	}

	/*
		Seed 1: 1-3 (Buchholz: 3)
		Seed 2: 3-0 (Buchholz: 1)
		Seed 3: 3-2 (Buchholz: -2)
		Seed 4: 3-2 (Buchholz: -3)
		Seed 5: 2-3 (Buchholz: -6)
		Seed 6: 1-3 (Buchholz: 1)
		Seed 7: 0-3 (Buchholz: 4)
		Seed 8: 3-2 (Buchholz: -2)
		Seed 9: 2-3 (Buchholz: -2)
		Seed 10: 0-3 (Buchholz: 3)
		Seed 11: 3-1 (Buchholz: 3)
		Seed 12: 1-3 (Buchholz: 6)
		Seed 13: 2-3 (Buchholz: 3)
		Seed 14: 3-1 (Buchholz: -4)
		Seed 15: 3-1 (Buchholz: -3)
		Seed 16: 3-0 (Buchholz: -2)
	*/
	expectedBuchholz := map[int]int{
		1:  3,
		2:  1,
		3:  -2,
		4:  -3,
		5:  -6,
		6:  1,
		7:  4,
		8:  -2,
		9:  -2,
		10: 3,
		11: 3,
		12: 6,
		13: 3,
		14: -4,
		15: -3,
		16: -2,
	}

	for seed, expBuch := range expectedBuchholz {
		buch := ss.CalculateBuchholz(seed)
		if buch != expBuch {
			t.Errorf("seed %d Buchholz = %d, want %d", seed, buch, expBuch)
		}
	}
}
