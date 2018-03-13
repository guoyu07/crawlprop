package core

import (
	"errors"
	"sync"
)

type SafeStringSet struct {
	sync.Mutex
	strings map[string]struct{}
}

func NewSafeStringSet() *SafeStringSet {
	s := SafeStringSet{
		strings: make(map[string]struct{}),
	}
	return &s
}

func (s *SafeStringSet) add(str string) {
	s.Lock()
	s.strings[str] = struct{}{}
	s.Unlock()
}

func (s *SafeStringSet) size() int {
	s.Lock()
	defer s.Unlock()
	return len(s.strings)
}

func (s *SafeStringSet) contains(str string) bool {
	s.Lock()
	_, contains := s.strings[str]
	s.Unlock()
	return contains
}

func (s *SafeStringSet) delete(str string) {
	s.Lock()
	delete(s.strings, str)
	s.Unlock()
}

func (s *SafeStringSet) popAny() (string, error) {
	s.Lock()
	defer s.Unlock()
	for str := range s.strings {
		delete(s.strings, str)
		return str, nil
	}
	return "", errors.New("No elements in set")
}

type SafeStringIntMap struct {
	sync.Mutex
	m map[string]int
}

func NewSafeStringIntMap() *SafeStringIntMap {
	s := SafeStringIntMap{
		m: make(map[string]int),
	}
	return &s
}

func (s *SafeStringIntMap) set(k string, v int) {
	s.Lock()
	s.m[k] = v
	s.Unlock()
}

func (s *SafeStringIntMap) size() int {
	s.Lock()
	defer s.Unlock()
	return len(s.m)
}

func (s *SafeStringIntMap) contains(str string) bool {
	s.Lock()
	_, contains := s.m[str]
	s.Unlock()
	return contains
}

func (s *SafeStringIntMap) delete(str string) {
	s.Lock()
	delete(s.m, str)
	s.Unlock()
}

func (s *SafeStringIntMap) get(str string) (int, bool) {
	s.Lock()
	value, ok := s.m[str]
	s.Unlock()
	return value, ok
}

func (s *SafeStringIntMap) popAny() (string, int, error) {
	s.Lock()
	defer s.Unlock()
	for k, v := range s.m {
		delete(s.m, k)
		return k, v, nil
	}
	return "", 0, errors.New("No elements in set")
}
