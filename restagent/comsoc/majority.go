package comsoc

// Simple Majority Method
func MajoritySWF(p Profile) (count Count, err error) {
	err = checkProfile(p)
	if err != nil {
		return nil, err
	}
	count = make(Count, len(p[0])) // Initialize the map
	for _, alt := range p[0] {
		// Initialize to 0
		count[alt] = 0
	}
	// Counting votes from the profile
	for _, votant := range p {
		_, ok := count[votant[0]] // votant[0] is the favorite of votant
		if ok {
			count[votant[0]]++
		} else {
			count[votant[0]] = 1
		}
	}
	return count, nil
}

func MajoritySCF(p Profile) (bestAlts []Alternative, err error) {
	count, err := MajoritySWF(p)
	if err != nil {
		return nil, err
	}
	return maxCount(count), nil
}
