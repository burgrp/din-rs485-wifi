package main

// regValue is intentionally integer-only for BleRiot wire model.
type regValue struct {
	value int32
	valid bool
}

type registerStore struct {
	byTag map[uint8]regValue
}

func newRegisterStore() *registerStore {
	return &registerStore{byTag: make(map[uint8]regValue, 64)}
}

func (s *registerStore) SetByTag(tag uint8, value int32, valid bool) bool {
	cur, ok := s.byTag[tag]
	if ok && cur.value == value && cur.valid == valid {
		return false
	}
	s.byTag[tag] = regValue{value: value, valid: valid}
	return true
}

func (s *registerStore) Snapshot() map[uint8]regValue {
	out := make(map[uint8]regValue, len(s.byTag))
	for k, v := range s.byTag {
		out[k] = v
	}
	return out
}
