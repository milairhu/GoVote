package comsoc

// Borda Method
func BordaSWF(p Profile) (Count, error) {
	err := checkProfile(p)
	if err != nil {
		return nil, err
	}
	count := make(Count, len(p[0])) // Initialize the map
	for _, alt := range p[0] {
		count[alt] = 0
	}
	// Counting votes from the profile
	var nbAlt = len(p[0])
	for _, votant := range p {
		for i, alt := range votant {
			_, ok := count[alt]
			if !ok {
				count[alt] = nbAlt - 1 - i
			} else {
				count[alt] += nbAlt - 1 - i
			}
		}
	}
	return count, nil
}

func BordaSCF(p Profile) (bestAlts []Alternative, err error) {
	count, err := BordaSWF(p)
	if err != nil {
		return nil, err
	}
	return maxCount(count), nil
}
