package comsoc

import "errors"

func winDuel(p Profile, alt1 Alternative, alt2 Alternative) (bool, error) {
	ok := checkProfile(p)
	if ok != nil {
		return false, errors.New("invalid profile")
	}

	var i int = 0
	var nbWin = 0
	for _, votant := range p {
		for votant[i] != alt1 && votant[i] != alt2 {
			i++
		}
		if votant[i] == alt1 {
			nbWin++
		}
		i = 0
	}
	return nbWin > len(p)-nbWin, nil
}

// Gives the Condorcet winner or nil if there is none
func CondorcetWinner(p Profile) (bestAlts []Alternative, err error) {
	ok := checkProfile(p)
	if ok != nil {
		return nil, errors.New("invalid profile")
	}
	var m int = len(p[0]) // number of alternatives
	var n int = len(p)    // number of individuals

	// Special cases
	if m == 1 {
		// If only one alternative
		bestAlts = []Alternative{p[0][0]}
		return bestAlts, nil
	}
	if n == 1 {
		// If only one individual
		return []Alternative{p[0][0]}, nil
	}

	// General case
	// We do all the duels. We see if one wins all its duels.
	// If yes, it's the Condorcet winner
	// If not, there is no Condorcet winner
	var i int = 0
	var j int = 0
	for i = 0; i < m; i++ {
		var nbWin int = 0
		for j = 0; j < m; j++ {
			if i != j {
				win, _ := winDuel(p, p[0][i], p[0][j])
				if win {
					nbWin++
				}
			}
		}
		if nbWin == m-1 {
			return []Alternative{p[0][i]}, nil
		}
	}
	return []Alternative{}, nil
}
