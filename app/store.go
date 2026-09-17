package main

import (
	"strconv"
	"sync"
	"time"
)

type Store struct {
	mu      sync.RWMutex
	waiters map[string][]chan bool
	db      map[string]Object
	expires map[string]time.Time
}

func NewStore() *Store {
	s := &Store{
		waiters: make(map[string][]chan bool),
		db:      make(map[string]Object),
		expires: make(map[string]time.Time),
	}

	go s.startActiveCleaner(100 * time.Millisecond)

	return s
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[key] = Object{Type: TypeString, Data: value}
	delete(s.expires, key)
}

func (s *Store) SetWithExpiry(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[key] = Object{Type: TypeString, Data: value}
	s.expires[key] = time.Now().Add(ttl)
}

func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lookup(key); !exists {
		return false
	}

	s.expires[key] = time.Now().Add(ttl)
	return true
}

func (s *Store) TTL(key string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lookup(key); !exists {
		return -2 // key does not exists
	}

	expiry, hasExpiry := s.expires[key]
	if !hasExpiry {
		return -1 // key exists but has no expiry
	}

	remaining := time.Until(expiry)
	if remaining <= 0 {
		return -2
	}

	return int64(remaining.Seconds())
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
	obj, exists := s.lookup(key)
	s.mu.RUnlock()

	if !exists {
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

func (s *Store) BLPop(key string, timeout time.Duration) (string, bool, error) {
	hasValue, err := s.registerBLPopWaiter(key)
	if err != nil {
		return "", false, err
	}

	var timeoutChan <-chan time.Time
	if timeout > 0 {
		timeoutChan = time.After(timeout)
	}

	select {
	case <-hasValue:
		values, err := s.LPop(key, 1)
		if err != nil || len(values) == 0 {
			return "", true, err
		}
		return values[0], false, nil

	case <-timeoutChan:
		if !s.removeBLPopWaiter(key, hasValue) {
			select {
			case <-hasValue:
				values, err := s.LPop(key, 1)
				if err == nil && len(values) > 0 {
					return values[0], false, nil
				}
			default:
			}
		}
		return "", true, nil
	}
}

func (s *Store) registerBLPopWaiter(key string) (chan bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hasValue := make(chan bool, 1)

	if len(s.waiters[key]) > 0 {
		s.waiters[key] = append(s.waiters[key], hasValue)
		return hasValue, nil
	}

	obj, exists := s.lookup(key)
	if exists {
		if err := checkType(obj.Type, TypeList); err != nil {
			return nil, err
		}

		list := obj.Data.(*Deque)
		if list.Len() > 0 {
			hasValue <- true
			return hasValue, nil
		}
	}

	s.waiters[key] = append(s.waiters[key], hasValue)
	return hasValue, nil
}

func (s *Store) removeBLPopWaiter(key string, ch chan bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	waitList, exists := s.waiters[key]
	if !exists {
		return false
	}

	for i, c := range waitList {
		if c == ch {
			s.waiters[key] = append(waitList[:i], waitList[i+1:]...)
			return true
		}
	}
	return false
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

func (s *Store) lookupWriteOrCreate(key string, expectedType ObjectType) (any, error) {
	obj, exists := s.lookup(key)
	if !exists {
		var data any

		switch expectedType {
		case TypeString:
			str := ""
			data = &str
		case TypeList:
			data = NewDeque()
		case TypeStream:
			data = NewStream()
		default:
			return nil, ErrUnsupportedType
		}

		s.db[key] = *NewObject(expectedType, data)
		delete(s.expires, key)
		return data, nil
	}

	if err := checkType(obj.Type, expectedType); err != nil {
		return nil, err
	}

	return obj.Data, nil
}

func (s *Store) getStreamForWrite(key string) (*Stream, error) {
	data, err := s.lookupWriteOrCreate(key, TypeStream)
	if err != nil {
		return nil, err
	}
	return data.(*Stream), nil
}

func (s *Store) getListForWrite(key string) (*Deque, error) {
	data, err := s.lookupWriteOrCreate(key, TypeList)
	if err != nil {
		return nil, err
	}
	return data.(*Deque), nil
}

func (s *Store) XAdd(key, ID string, values []string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stream, err := s.getStreamForWrite(key)
	if err != nil {
		return "", err
	}

	err = stream.Append(ID, values)

	if err != nil {
		return "", err
	}

	return ID, nil
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

	if expiry, hasExpiry := s.expires[key]; hasExpiry && time.Now().After(expiry) {
		return Object{}, false
	}

	return obj, true
}

func (s *Store) delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.db, key)
	delete(s.expires, key)
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

	now := time.Now()
	for key, expiry := range s.expires {
		if now.After(expiry) {
			delete(s.db, key)
			delete(s.expires, key)
		}
	}
}

func (s *Store) push(key string, isBack bool, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.getListForWrite(key)
	if err != nil {
		return 0, err
	}

	list.PushMultipleValues(values, isBack)

	// Notify waiting BLPOP clients (at most len(values) waiters)
	waitList, exists := s.waiters[key]

	if exists && len(waitList) > 0 {
		numToNotify := len(values)

		if numToNotify > len(waitList) {
			numToNotify = len(waitList)
		}

		for i := 0; i < numToNotify; i++ {
			waitList[i] <- true
		}
		s.waiters[key] = waitList[numToNotify:]
	}

	return list.Len(), nil
}

func (s *Store) pop(key string, del_keys int, isBack bool) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	obj, exists := s.lookup(key)

	values := make([]string, del_keys)

	if !exists {
		return nil, nil
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
