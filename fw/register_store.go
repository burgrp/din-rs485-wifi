package main

type registerStore struct {
	values []int32
	valid  []bool
}

func newRegisterStore(maxTag uint8) *registerStore {
	n := int(maxTag) + 1
	return &registerStore{
		values: make([]int32, n),
		valid:  make([]bool, n),
	}
}

func (s *registerStore) SetByTag(tag uint8, value int32, valid bool) bool {
	if int(tag) >= len(s.values) {
		return false
	}
	if s.values[tag] == value && s.valid[tag] == valid {
		return false
	}
	s.values[tag] = value
	s.valid[tag] = valid
	return true
}

func (s *registerStore) GetByTag(tag uint8) (value int32, valid bool, ok bool) {
	if int(tag) >= len(s.values) {
		return 0, false, false
	}
	return s.values[tag], s.valid[tag], true
}
