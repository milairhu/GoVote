package comsoc

/*
* Single Transferable Vote (STV)
* Each individual gives their preference order
* For n candidates, we have n−1 rounds
* (unless a strict majority for one candidate is reached earlier)
* We assume that in each round each individual "votes" for their preferred candidate
* (among those still in the race)
* In each round, the candidate with the fewest votes is eliminated
 */

func STV_SWF(p Profile) (Count, error) {
	ok := checkProfile(p)
	if ok != nil {
		return nil, ok
	}
	copyP := make(Profile, len(p)) // We copy the profile to be able to perform deletions without affecting the original
	for i, votant := range p {
		copyP[i] = make([]Alternative, len(votant))
		copy(copyP[i], votant)
	}
	resMap := make(Count, len(p[0]))
	// We initialize the map to 0
	for _, alt := range copyP[0] {
		resMap[alt] = 0
	}

	for nbToursRestants := len(copyP[0]) - 1; nbToursRestants > 0; nbToursRestants-- {
		// We count the votes and eliminate the candidate with the fewest votes
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
		// We have the scores for each candidate for this round
		var miniCount int = len(copyP) + 1
		var miniAlt Alternative
		for alt, count := range comptMap {
			if count < miniCount {
				miniCount = count
				miniAlt = alt
			}
		}
		// We eliminate the candidate with the fewest votes from the votes
		for indP, votant := range copyP {
			for i, alt := range votant {
				if alt == miniAlt {
					votant[i] = votant[len(votant)-1]
				}
			}
			copyP[indP] = votant[:len(votant)-1]
		}
		// We increment
		for _, alt := range copyP[0] {
			if alt != miniAlt {
				resMap[alt]++
			}
		}

	}
	return resMap, nil
}

func STV_SCF(p Profile) (bestAlts []Alternative, err error) {
	count, err := STV_SWF(p)
	if err != nil {
		return nil, err
	}
	return maxCount(count), nil
}
