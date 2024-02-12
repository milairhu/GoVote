package comsoc

import (
	"errors"
)

///// Tie breakers
/**
 * In case of ties, the SCFs must return a single element
 * and the SWFs must return a total order without ties.
 * For this purpose, tie-breaking functions are used
 * which, given a set of alternatives, return the best one.
 * They respect the following signature
 * (an error may occur if the slice of alternatives is empty):
**/
func TieBreakFactory(strictOrder []Alternative) func([]Alternative) (Alternative, error) {
	// The provided alternatives are used to break ties -> strict order
	return func(alts []Alternative) (Alternative, error) {
		// We assume that the provided alternatives are initially tied
		if len(alts) == 0 {
			return -1, errors.New("no alternatives provided")
		} else {
			order := make(map[Alternative]int, len(strictOrder))
			for i, alt := range strictOrder {
				order[alt] = len(strictOrder) - i
			}
			// The order map contains the alternatives associated with their rank
			var maxVal int
			var maxAlt Alternative
			for _, alt := range alts {
				if order[alt] > maxVal {
					maxVal = order[alt]
					maxAlt = alt
				}
			}
			return maxAlt, nil
		}
	}
}

// To obtain SWFs without ties
func SWFFactory(swf func(p Profile) (Count, error), tieBreaker func([]Alternative) (Alternative, error)) func(Profile) ([]Alternative, error) {
	// Returns a function that returns the ordered alternatives
	return func(p Profile) ([]Alternative, error) {
		count, err := swf(p)
		if err != nil {
			return nil, err
		}
		res := make([]Alternative, len(count))
		// We fill res with the alternatives: the higher count[alt] is, the higher alt is ranked

		// Professor's idea: we multiply all the alternatives by
		// nbAlt, and for each tied alternative,
		// we do +1, +2 etc to differentiate
		// -> here, we don't do that
		invCount := make(map[int][]Alternative, len(count)) // dict {score: [candidates]}
		var maxScore int
		var minScore int
		for alt, score := range count {
			// We fill the invCount dictionary and record the max and min scores
			invCount[score] = append(invCount[score], alt)
			if score > maxScore {
				maxScore = score
			} else if score < minScore {
				minScore = score
			}

		}
		var currIndex = 0
		for i := maxScore; i >= minScore; i-- {
			tab, ok := invCount[i]
			if ok {
				// If we have candidates corresponding to this score,
				// we sort them according to the tiebreaker and add them
				// to the res array
				for len(tab) > 1 {
					// as long as there are multiple tied elements,
					// we remove the best one from the list
					// and add it to res
					best, err := tieBreaker(tab)
					if err != nil {
						return nil, err
					}
					res[currIndex] = best
					currIndex++
					// We remove the element from tab
					for i, alt := range tab {
						if alt == best {
							tab[i] = tab[len(tab)-1]
							tab = tab[:len(tab)-1]
							break
						}
					}
				}
				if len(tab) == 1 {
					// We add the last element to the array
					res[currIndex] = tab[0]
					currIndex++
				}
			}
		}
		return res, nil
	}
}
func SCFFactory(scf func(p Profile) ([]Alternative, error), tieBreaker func([]Alternative) (Alternative, error)) func(Profile) (Alternative, error) {
	// Applies the scf function on the profile then breaks ties with tieBreaker. Returns the function that applies scf but without ties
	return func(p Profile) (Alternative, error) {
		bestAlts, err := scf(p)
		if err != nil {
			return -1, err
		}
		// We have the best alternatives. We use tiebreaker to break ties
		return tieBreaker(bestAlts)
	}
}

// Note: it is necessary to create a specific SWF function with a particular Tie-break for approval because the threshold must be taken into account
func MakeApprovalRankingWithTieBreak(p Profile, threshold []int, tieBreaker func([]Alternative) (Alternative, error)) ([]Alternative, error) {
	count, err := ApprovalSWF(p, threshold)
	if err != nil {
		return nil, err
	}
	res := make([]Alternative, len(count))
	// We fill res with the alternatives: the higher count[alt] is, the higher alt is ranked

	// Professor's idea: we multiply all the alternatives by
	// nbAlt, and for each tied alternative,
	// we do +1, +2 etc to differentiate
	// -> here, we don't do that
	invCount := make(map[int][]Alternative, len(count)) // dict {score: [candidates]}
	var maxScore int
	var minScore int
	for alt, score := range count {
		// We fill the invCount dictionary and record the max and min scores
		invCount[score] = append(invCount[score], alt)
		if score > maxScore {
			maxScore = score
		} else if score < minScore {
			minScore = score
		}

	}
	var currIndex = 0
	for i := maxScore; i >= minScore; i-- {
		tab, ok := invCount[i]
		if ok {
			// If we have candidates corresponding to this score,
			// we sort them according to the tiebreaker and add them
			// to the res array
			for len(tab) > 1 {
				// as long as there are multiple tied elements,
				// we remove the best one from the list
				// and add it to res
				best, err := tieBreaker(tab)
				if err != nil {
					return nil, err
				}
				res[currIndex] = best
				currIndex++
				// We remove the element from tab
				for i, alt := range tab {
					if alt == best {
						tab[i] = tab[len(tab)-1]
						tab = tab[:len(tab)-1]
						break
					}
				}
			}
			if len(tab) == 1 {
				// We add the last element to the array
				res[currIndex] = tab[0]
				currIndex++
			}
		}
	}
	return res, nil

}

// Note: We need to create a specific SWF function with a tie-break for STV because the tie-breaking process is different. We use the tie-break within the algorithm itself.
func STV_SWF_TieBreak(p Profile, tieBreak []Alternative) ([]Alternative, error) {
	// Check if the profile is valid
	ok := checkProfile(p)
	if ok != nil {
		return nil, ok
	}

	// Create a copy of the profile to avoid modifying the original
	copyP := make(Profile, len(p))
	for i, votant := range p {
		copyP[i] = make([]Alternative, len(votant))
		copy(copyP[i], votant)
	}

	// Create a map to associate each value with its position in the tie-break
	tieBreakMap := make(map[Alternative]int, len(tieBreak))
	for i, alt := range tieBreak {
		tieBreakMap[alt] = len(tieBreak) - i - 1
	}

	// Initialize the result map
	resMap := make(Count, len(p[0]))
	for _, alt := range copyP[0] {
		resMap[alt] = 0
	}

	// Perform the STV process
	for nbToursRestants := len(copyP[0]) - 1; nbToursRestants > 0; nbToursRestants-- {
		// Count the votes for each alternative and eliminate the worst candidate

		comptMap := make(Count, len(copyP[0]))
		for _, alt := range copyP[0] {
			comptMap[alt] = 0
		}
		for _, votant := range copyP {
			_, ok := comptMap[votant[0]]
			if ok {
				comptMap[votant[0]]++
			} else {
				comptMap[votant[0]] = 1
			}
		}
		// Get the scores for this round
		var miniCount int = len(copyP) + 1
		miniAlts := make([]Alternative, 0)
		for alt, count := range comptMap {
			if count < miniCount {
				miniCount = count
				miniAlts = []Alternative{alt}
			} else if count == miniCount {
				miniAlts = append(miniAlts, alt)
			}
		}
		// Get the worst candidates and eliminate one based on the provided tie-break
		var miniAlt Alternative
		miniValInTieBreak := len(tieBreak) + 1
		for _, alt := range miniAlts {
			if tieBreakMap[alt] < miniValInTieBreak {
				miniValInTieBreak = tieBreakMap[alt]
				miniAlt = alt
			}
		}
		// Eliminate the selected candidate from all votes
		for indP, votant := range copyP {
			var found bool
			for i, alt := range votant {
				if alt == miniAlt {
					found = true
				}
				if found && i < len(votant)-1 {
					votant[i] = votant[i+1]
				}
			}
			copyP[indP] = votant[:len(votant)-1]
		}
		// Increment the score of each candidate passing to the next round
		for _, alt := range copyP[0] {
			if alt != miniAlt {
				resMap[alt]++
			}
		}
	}
	// Create a slice of alternatives ordered according to the result map
	res := make([]Alternative, len(resMap))
	for alt, score := range resMap {
		res[len(resMap)-1-score] = alt
	}
	return res, nil
}
