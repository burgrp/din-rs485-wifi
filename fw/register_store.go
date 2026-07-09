package main

const maxWireTag = 64

type registerStore struct {
	values [maxWireTag + 1]int32
	valid  [maxWireTag + 1]bool
	used   [maxWireTag + 1]bool
}

func newRegisterStore() *registerStore {
	return &registerStore{}
}

func (s *registerStore) SetByTag(tag uint8, value int32, valid bool) bool {
	if int(tag) > maxWireTag {
		return false
	}
	if s.used[tag] && s.values[tag] == value && s.valid[tag] == valid {
		return false
	}
	s.values[tag] = value
	s.valid[tag] = valid
	s.used[tag] = true
	return true
}

func (s *registerStore) GetByTag(tag uint8) (value int32, valid bool, ok bool) {
	if int(tag) > maxWireTag {
		return 0, false, false
	}
	if !s.used[tag] {
		return 0, false, false
	}
	return s.values[tag], s.valid[tag], true
}
