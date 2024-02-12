package comsoc

import (
	"errors"
)

// Returns the index where alt is found in prefs
func rank(alt Alternative, prefs []Alternative) int {
	for i, a := range prefs {
		if a == alt {
			return i
		}
	}
	return -1
}

// Returns true if alt1 is preferred to alt2
func isPref(alt1, alt2 Alternative, prefs []Alternative) bool {
	for _, a := range prefs {
		if a == alt1 {
			return true
		}
		if a == alt2 {
			return false
		}
	}
	return false
}

// Returns the best alternatives for a given count
func maxCount(count Count) (bestAlts []Alternative) {
	var maxi = 0
	for i, v := range count {
		if v > maxi {
			maxi = v
			bestAlts = []Alternative{i}
		} else if v == maxi {
			bestAlts = append(bestAlts, i)
		}
	}
	return
}

// Checks the given profile, e.g., that they are all complete and each alternative appears only once per preference
func checkProfile(prefs Profile) error {
	if len(prefs) < 1 {
		return errors.New("no votes submitted")
	}
	if len(prefs[0]) < 2 {
		return errors.New("less than 2 candidates")
	}
	for _, v := range prefs {
		if len(v) != len(prefs[0]) {
			return errors.New("profile is not complete")
		}
	}
	for _, v := range prefs {
		for i, a := range v {
			for j, b := range v {
				if i != j && a == b {
					return errors.New("profile is not correct")
				}
			}
		}
	}
	return nil
}

// Checks the given profile, e.g., that they are all complete and each alternative of alts appears exactly once per preference
func checkProfileAlternative(prefs Profile, alts []Alternative) error {
	err := checkProfile(prefs)
	if err != nil {
		return err
	}

	for _, prof := range prefs {
		for _, a := range alts {
			var isPresent = false
			for _, b := range prof {
				if a == b {
					isPresent = true
				}
			}
			if !isPresent {
				return errors.New("profile is not correct: a alternative is missing")
			}
		}
	}
	return nil
}
