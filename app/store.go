package main

import (
	"sync"
	"time"
)

type Store struct {
	mu sync.RWMutex
	db map[string]Object
}

func NewStore() *Store {
	s := &Store{
		db: make(map[string]Object),
	}

	go s.startActiveCleaner(100 * time.Millisecond)

	return s
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[key] = Object{TypeString, value, nil}
}

func (s *Store) SetWithExpiry(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiry := time.Now().Add(ttl)
	s.db[key] = Object{TypeString, value, &expiry}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	value, exists := s.db[key]
	s.mu.RUnlock()

	if !exists {
		return "", false
	}

	if value.ExpiredAt != nil && time.Now().After(*value.ExpiredAt) {
		s.delete(key)
		return "", false
	}

	return value.Data.(string), true
}

func (s *Store) delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.db[key]

	if exists {
		delete(s.db, key)
	}
}

func (s *Store) startActiveCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.sweepExpiredKeys()
	}
}

func (s *Store) sweepExpiredKeys() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, val := range s.db {
		if val.ExpiredAt != nil && time.Now().After(*val.ExpiredAt) {
			delete(s.db, key)
		}
	}
}

func checkType(objType, expectedType ObjectType) bool {
	if objType != expectedType {
		// return errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		return true
	}
	return false
}
