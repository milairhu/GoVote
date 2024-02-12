package comsoc

import "fmt"

// ApprovalSWF calculates the social welfare function for the approval voting method.
// It counts the votes for each alternative up to the threshold for each voter.
func ApprovalSWF(p Profile, thresholds []int) (count Count, err error) {
	err = checkProfile(p)
	if err != nil {
		return nil, err
	}
	count = make(Count, len(p[0])) //initializing the map
	for _, alt := range p[0] {
		// Initialize to 0
		count[alt] = 0
	}
	// Counting the votes of all profiles, from 0 to thresholds[i]
	for indVoter, voter := range p {
		// For all recorded votes
		for i, alt := range voter {
			// For each alternative of a vote
			if thresholds[indVoter] < 0 || thresholds[indVoter] > len(voter) {
				return nil, fmt.Errorf("threshold %d is incorrect with %d alternatives", thresholds[indVoter], len(voter))
			}
			if i < thresholds[indVoter] {
				// If we count this alternative
				_, ok := count[alt]
				if !ok {
					count[alt] = 1
				} else {
					count[alt]++
				}
			} else {
				break
			}
		}
	}
	return count, nil
}

// ApprovalSCF calculates the social choice function for the approval voting method.
// It returns the alternatives with the highest count.
func ApprovalSCF(p Profile, thresholds []int) (bestAlts []Alternative, err error) {
	count, err := ApprovalSWF(p, thresholds)
	if err != nil {
		return nil, err
	}
	return maxCount(count), nil
}
