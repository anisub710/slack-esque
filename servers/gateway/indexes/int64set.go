package indexes

type int64set map[int64]struct{}

func (s int64set) add(value int64) bool {
	_, exists := s[value]
	s[value] = struct{}{}
	return !exists
}

func (s int64set) remove(value int64) bool {
	_, exists := s[value]
	delete(s, value)
	return exists
}

func (s int64set) has(value int64) bool {
	_, exists := s[value]
	return exists
}

func (s int64set) all() []int64 {
	ret := make([]int64, 0, len(s))
	for k := range s {
		ret = append(ret, k)
	}
	// to sort
	// sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return ret
}
