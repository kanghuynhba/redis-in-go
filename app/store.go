package main

import (
	"errors"
	"sync"
	"time"
)

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

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

func (s *Store) Get(key string) (string, bool, error) {
	s.mu.RLock()
	obj, exists := s.db[key]
	s.mu.RUnlock()

	if !exists {
		return "", false, nil
	}

	if obj.ExpiredAt != nil && time.Now().After(*obj.ExpiredAt) {
		s.delete(key)
		return "", false, nil
	}

	if err := checkType(obj.Type, TypeString); err != nil {
		return "", false, err
	}

	return obj.Data.(string), true, nil
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

func checkType(objType, expectedType ObjectType) error {
	if objType != expectedType {
		return ErrWrongType
	}
	return nil
}

func (s *Store) RPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	obj, exist := s.db[key]

	list := []string{}

	if !exist || (obj.ExpiredAt != nil && time.Now().After(*obj.ExpiredAt)) {
		newList := append([]string{}, values...)
		s.db[key] = Object{
			Type: TypeList, Data: newList,
		}
		return len(newList), nil
	}

	if err := checkType(obj.Type, TypeList); err != nil {
		return 0, err
	}

	list = obj.Data.([]string)
	list = append(list, values...)

	obj.Data = list
	s.db[key] = obj

	return len(list), nil
}

func (s *Store) LRange(key string, lBound, rBound int) ([]string, error) {
	s.mu.RLock()
	obj, exist := s.db[key]
	s.mu.RUnlock()

	if !exist || (obj.ExpiredAt != nil && time.Now().After(*obj.ExpiredAt)) {
		s.delete(key)
		return nil, nil
	}

	if err := checkType(obj.Type, TypeList); err != nil {
		return nil, err
	}

	list_len := len(obj.Data.([]string))

	lBound = max(lBound, -list_len)
	rBound = min(rBound, list_len-1)

	if lBound < 0 {
		lBound += list_len
	}

	if rBound < 0 {
		rBound += list_len
	}

	values := obj.Data.([]string)

	return values[lBound : rBound+1], nil
}
