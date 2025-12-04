package main

import (
	"math/rand"
	"testing"
)

// mockSwissSystem is a test implementation with predefined outcomes
type mockSwissSystem struct {
	teams   []*Team
	records []*Record
}

func (m *mockSwissSystem) Reset() {
	// Reset records to initial state
	for _, rec := range m.records {
		rec.Wins = 0
		rec.Losses = 0
	}
}

func (m *mockSwissSystem) SimulateTournament() {
	// Set predefined outcomes for testing
	// For example, let's set first team to 3-0, second to 3-1, third to 0-3
	// We need to find teams by their seed
	// Since records are indexed by seed, and maxSeed may be larger than number of teams
	// Let's find the indices for seeds 1, 2, 3
	// In our test, we'll use seeds 1, 2, 3 which are present
	// Reset first
	m.Reset()

	// Set specific outcomes
	// Assuming records are sized to maxSeed+1
	// For seed 1 (index 1)
	if len(m.records) > 1 {
		m.records[1].Wins = 3
		m.records[1].Losses = 0
	}
	// For seed 2 (index 2)
	if len(m.records) > 2 {
		m.records[2].Wins = 3
		m.records[2].Losses = 1
	}
	// For seed 3 (index 3)
	if len(m.records) > 3 {
		m.records[3].Wins = 0
		m.records[3].Losses = 3
	}
}

func (m *mockSwissSystem) Records() []*Record {
	return m.records
}

func newMockSwissSystem(teams []*Team, sigma []int, rng *rand.Rand, prob [][]float64) SwissSystemInterface {
	// Find max seed
	maxSeed := 0
	for _, t := range teams {
		if t.Seed > maxSeed {
			maxSeed = t.Seed
		}
	}
	limit := maxSeed + 1
	records := make([]*Record, limit)
	for i := range records {
		records[i] = &Record{}
	}
	return &mockSwissSystem{
		teams:   teams,
		records: records,
	}
}

func TestSimulationWithMock(t *testing.T) {
	// Use the same teams as in TestBudapest2025Stage2
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

	rng := rand.New(rand.NewSource(42))
	sim, err := NewSimulationWithFactory(600, teams, rng, func(teams []*Team, sigma []int, rng *rand.Rand, prob [][]float64) SwissSystemInterface {
		// Create a mock that always produces the exact results from TestBudapest2025Stage2
		maxSeed := 0
		for _, t := range teams {
			if t.Seed > maxSeed {
				maxSeed = t.Seed
			}
		}
		limit := maxSeed + 1
		records := make([]*Record, limit)
		for i := range records {
			records[i] = &Record{}
		}
		return &budapestMockSwissSystem{
			teams:   teams,
			records: records,
		}
	})
	if err != nil {
		t.Fatalf("Failed to create simulation: %v", err)
	}

	// Create predictions based on the actual outcomes from TestBudapest2025Stage2
	// According to the mock, the final records are:
	// 3-0: seeds 2 (NaVi), 16 (FaZe)
	// 3-1 or 3-2: seeds 3 (Liquid), 4 (3DMAX), 8 (Passion UA), 11 (B8), 14 (PARIVISION), 15 (Imperial)
	// 0-3: seeds 7 (MIBR), 10 (FlyQuest)
	// Other seeds have records that don't fit into these categories (1-3, 2-3, etc.)

	// Let's create three predictions, all respecting the category limits:
	// 1. A perfect prediction that matches all outcomes exactly
	perfectPrediction := map[Category][]int{
		Cat3_0: {2, 16},               // 2 teams: correct
		CatAdv: {3, 4, 8, 11, 14, 15}, // 6 teams: correct
		Cat0_3: {7, 10},               // 2 teams: correct
	}

	// 2. A prediction with some mistakes but still valid
	imperfectPrediction := map[Category][]int{
		Cat3_0: {2, 11},               // Wrong: seed 11 is 3-1, not 3-0
		CatAdv: {3, 4, 8, 14, 15, 16}, // Wrong: seed 16 is 3-0, not CatAdv
		Cat0_3: {7, 10},               // Correct
	}

	// 3. A prediction with exactly 5 correct picks (should be successful)
	fiveCorrectPrediction := map[Category][]int{
		Cat3_0: {2, 16},            // Both correct: 2 correct
		CatAdv: {3, 4, 1, 5, 6, 9}, // Only seeds 3 and 4 are correct: 2 correct
		Cat0_3: {7, 1},             // Seed 7 is correct, seed 1 is not 0-3: 1 correct
	}

	// 4. A prediction with exactly 4 correct picks (should fail)
	fourCorrectPrediction := map[Category][]int{
		Cat3_0: {2, 1},             // Seed 2 correct, seed 1 incorrect: 1 correct
		CatAdv: {3, 4, 1, 5, 6, 9}, // Seeds 3 and 4 correct: 2 correct
		Cat0_3: {7, 5},             // Seed 7 correct, seed 5 incorrect: 1 correct
	}

	// 5. A prediction with many mistakes
	failedPrediction := map[Category][]int{
		Cat3_0: {1, 5},                // Both are not 3-0
		CatAdv: {6, 7, 9, 10, 12, 13}, // All are not advancing teams
		Cat0_3: {2, 3},                // Both are not 0-3
	}

	predictions := []map[Category][]int{
		perfectPrediction,
		imperfectPrediction,
		fiveCorrectPrediction,
		fourCorrectPrediction,
		failedPrediction,
	}

	// Run the simulation
	teamResults, percentages := sim.Run(1, 1, predictions)

	// Check teamResults
	// Since we ran exactly 1 simulation, each team should have count 1 in exactly one category
	// based on the mock's fixed outcome, and 0 in the other two categories.
	// Teams that don't fall into any category should have 0 in all categories.
	expectedCategoryForSeed := map[int]Category{
		2:  Cat3_0,
		16: Cat3_0,
		3:  CatAdv,
		4:  CatAdv,
		8:  CatAdv,
		11: CatAdv,
		14: CatAdv,
		15: CatAdv,
		7:  Cat0_3,
		10: Cat0_3,
	}
	// For seeds not in the map, they should have 0 in all categories
	for _, team := range teams {
		seed := team.Seed
		results, ok := teamResults[team]
		if !ok {
			t.Errorf("teamResults missing entry for team %s (seed %d)", team.Name, seed)
			continue
		}
		expectedCat, shouldHaveCat := expectedCategoryForSeed[seed]
		for _, cat := range []Category{Cat3_0, CatAdv, Cat0_3} {
			count := results[cat]
			if shouldHaveCat && cat == expectedCat {
				if count != 1 {
					t.Errorf("team %s (seed %d) expected count 1 for category %v, got %d", team.Name, seed, cat, count)
				}
			} else {
				if count != 0 {
					t.Errorf("team %s (seed %d) expected count 0 for category %v, got %d", team.Name, seed, cat, count)
				}
			}
		}
	}

	expectedPercentages := []float64{
		100.0, // Perfect prediction: all picks correct
		100.0, // Imperfect prediction: 8 correct picks >= 5 required
		100.0, // Five correct prediction: 5 correct picks == 5 required
		0.0,   // Four correct prediction: 4 correct picks < 5 required
		0.0,   // Failed prediction: likely 0 correct picks
	}

	if len(percentages) != len(expectedPercentages) {
		t.Fatalf("Expected %d percentages, got %d", len(expectedPercentages), len(percentages))
	}
	for i, expPerc := range expectedPercentages {
		if percentages[i] != expPerc {
			t.Errorf("Prediction %d: expected percentage %.2f, got %.2f", i, expPerc, percentages[i])
		}
	}

}

// budapestMockSwissSystem always produces the exact results from TestBudapest2025Stage2
type budapestMockSwissSystem struct {
	teams   []*Team
	records []*Record
}

func (m *budapestMockSwissSystem) Reset() {
	// Reset all records to zero
	for _, rec := range m.records {
		rec.Wins = 0
		rec.Losses = 0
	}
}

func (m *budapestMockSwissSystem) SimulateTournament() {
	// Reset first
	m.Reset()

	// Set records according to TestBudapest2025Stage2 expected outcomes
	// Seed: wins, losses
	records := map[int]struct{ wins, losses int }{
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

	for seed, rec := range records {
		if seed < len(m.records) {
			m.records[seed].Wins = rec.wins
			m.records[seed].Losses = rec.losses
		}
	}
}

func (m *budapestMockSwissSystem) Records() []*Record {
	return m.records
}
