package main

// generateAllPartitions generates all ways to partition teams into groups of specified sizes.
// teams is a slice of unique team identifiers (seeds).
// groupSizes specifies the size of each group, in order.
// If maxCount > 0, the function stops after generating maxCount partitions.
// The returned slice contains at most maxCount partitions, each partition is a slice of groups,
// where groups[i] corresponds to groupSizes[i].
func generateAllPartitions(teams []int, groupSizes []int, maxCount int) [][][]int {
	var partitions [][][]int

	var recurse func(remaining []int, sizes []int, current [][]int)
	recurse = func(remaining []int, sizes []int, current [][]int) {
		if len(sizes) == 0 {
			if len(remaining) == 0 {
				// make a copy of the current partition
				part := make([][]int, len(current))
				for i, g := range current {
					gcopy := make([]int, len(g))
					copy(gcopy, g)
					part[i] = gcopy
				}
				partitions = append(partitions, part)
			}
			return
		}
		if maxCount > 0 && len(partitions) >= maxCount {
			return
		}
		size := sizes[0]
		n := len(remaining)
		if size > n {
			return
		}
		// generate all combinations of 'size' from 'remaining'
		indices := make([]int, size)
		for i := range indices {
			indices[i] = i
		}
		for {
			// build group from current indices
			group := make([]int, size)
			for i, idx := range indices {
				group[i] = remaining[idx]
			}
			// create newRemaining by removing the chosen elements
			// use a simple boolean map for O(n) removal
			chosen := make(map[int]bool)
			for _, v := range group {
				chosen[v] = true
			}
			newRemaining := make([]int, 0, n-size)
			for _, v := range remaining {
				if !chosen[v] {
					newRemaining = append(newRemaining, v)
				}
			}
			// continue recursion
			recurse(newRemaining, sizes[1:], append(current, group))

			if maxCount > 0 && len(partitions) >= maxCount {
				return
			}
			// move to next combination (lexicographic order)
			i := size - 1
			for ; i >= 0; i-- {
				if indices[i] != i+n-size {
					break
				}
			}
			if i < 0 {
				break // no more combinations
			}
			indices[i]++
			for j := i + 1; j < size; j++ {
				indices[j] = indices[j-1] + 1
			}
		}
	}

	recurse(teams, groupSizes, nil)
	return partitions
}
