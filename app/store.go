package main

import (
	"errors"
	"strconv"
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
	obj, exists := s.lookup(key)
	s.mu.RUnlock()

	if !exists {
		s.delete(key)
		return "", false, nil
	}

	if err := checkType(obj.Type, TypeString); err != nil {
		return "", false, err
	}

	return obj.Data.(string), true, nil
}

func (s *Store) RPush(key string, values ...string) (int, error) {
	return s.push(key, true, values...)
}

func (s *Store) LRange(key string, start, stop int) ([]string, error) {
	s.mu.RLock()
	obj, exist := s.lookup(key)
	s.mu.RUnlock()

	if !exist {
		s.delete(key)
		return []string{}, nil
	}

	if err := checkType(obj.Type, TypeList); err != nil {
		return nil, err
	}

	list := obj.Data.(*Deque)
	list_len := list.Len()

	start = max(start, -list_len)
	stop = min(stop, list_len-1)

	if start < 0 {
		start += list_len
	}

	if stop < 0 {
		stop += list_len
	}

	if start >= list_len || start > stop {
		return []string{}, nil
	}

	return list.Range(start, stop), nil
}

func (s *Store) LPush(key string, values ...string) (int, error) {
	return s.push(key, false, values...)
}

func (s *Store) LLen(key string) (int, error) {
	s.mu.RLock()
	obj, exists := s.lookup(key)
	s.mu.RUnlock()

	if !exists {
		s.delete(key)
		return 0, nil
	}

	if err := checkType(obj.Type, TypeList); err != nil {
		return 0, err
	}

	list := obj.Data.(*Deque)

	return list.Len(), nil
}

func (s *Store) LPop(key string, del_keys int) ([]string, error) {
	return s.pop(key, del_keys, false)
}

func (s *Store) Type(key string) string {
	s.mu.RLock()
	obj, exists := s.lookup(key)
	s.mu.RUnlock()

	if !exists {
		s.delete(key)
		return string(TypeNone)
	}

	return string(obj.Type)
}

func (s *Store) Incr(key string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	obj, exists := s.lookup(key)

	if !exists {
		s.db[key] = Object{Type: TypeString, Data: "1"}
		return 1, nil
	}

	if err := checkType(obj.Type, TypeString); err != nil {
		return 0, err
	}

	num, err := strconv.Atoi(obj.Data.(string))

	if err != nil {
		return 0, ErrNotInteger
	}

	num++
	obj.Data = strconv.Itoa(num)
	s.db[key] = obj

	return num, nil
}

// helper

func (s *Store) lookup(key string) (Object, bool) {
	obj, exists := s.db[key]

	if !exists {
		return Object{}, false
	}

	if obj.ExpiredAt != nil && time.Now().After(*obj.ExpiredAt) {
		return Object{}, false
	}

	return obj, true
}

func (s *Store) delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.db[key]

	if exists {
		delete(s.db, key)
	}
}

func checkType(objType, expectedType ObjectType) error {
	if objType != expectedType {
		return ErrWrongType
	}
	return nil
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

func (s *Store) push(key string, isBack bool, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	obj, exists := s.lookup(key)

	if !exists {
		newList := NewDeque()
		newList.PushMultipleValues(values, isBack)
		s.db[key] = Object{
			Type: TypeList, Data: newList,
		}
		return newList.Len(), nil
	}

	if err := checkType(obj.Type, TypeList); err != nil {
		return 0, err
	}

	list := obj.Data.(*Deque)
	list.PushMultipleValues(values, isBack)

	obj.Data = list
	s.db[key] = obj

	return list.Len(), nil
}

func (s *Store) pop(key string, del_keys int, isBack bool) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	obj, exists := s.lookup(key)

	values := make([]string, del_keys)

	if !exists {
		return values, nil
	}

	if err := checkType(obj.Type, TypeList); err != nil {
		return values, err
	}

	list := obj.Data.(*Deque)

	values = list.PopMultipleValues(del_keys, isBack)

	obj.Data = list
	s.db[key] = obj

	return values, nil

}
