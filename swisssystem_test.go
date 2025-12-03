package main

import (
	"math/rand"
	"testing"
)

// TestBudapest2025Stage2 replicates the exact Swiss bracket from the
// Budapest 2025 Stage 2 tournament using fixed probabilities.
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

	sigma := []int{100}
	rng := rand.New(rand.NewSource(42))
	ss := NewSwissSystem(teams, sigma, rng)

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
		ss.SetProbBySeed(m.winner, m.loser, 1.0)
	}

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
