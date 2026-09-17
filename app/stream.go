package main

import (
	"strconv"
	"strings"
	"time"
)

type StreamEntry struct {
	ID     string
	Values []string
}

func (se *StreamEntry) DecodeID() (time.Time, int, error) {

	if se.ID == "*" {
		return time.Now().Truncate(time.Millisecond), -1, nil
	}

	parts := strings.Split(se.ID, "-")

	if len(parts) != 2 {
		return time.Time{}, 0, ErrInvalidStreamID
	}

	ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, 0, ErrInvalidStreamID
	}

	if parts[1] == "*" {
		return time.UnixMilli(ms), -1, nil
	}

	seq, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, 0, ErrInvalidStreamID
	}

	if ms == 0 && seq == 0 {
		return time.Time{}, 0, ErrStreamIDZero
	}

	return time.UnixMilli(ms), seq, nil
}

type Stream []StreamEntry

func NewStream() *Stream {
	s := make(Stream, 0)
	return &s
}

func (s *Stream) Append(ID string, values []string) error {
	streamEntry := StreamEntry{ID: ID, Values: values}

	ms, seqNumb, err := streamEntry.DecodeID()

	if err != nil {
		return err
	}

	lastEntry, exists := s.Top()

	if exists {
		lastMsTime, lastSeqNumb, _ := lastEntry.DecodeID()
		timeCompare := ms.Compare(lastMsTime)
		if timeCompare == -1 || (timeCompare == 0 && seqNumb <= lastSeqNumb) {
			return ErrStreamIDSmaller
		}
	}

	*s = append(*s, streamEntry)

	return nil
}

func (s *Stream) Top() (StreamEntry, bool) {

	if s == nil || s.Len() == 0 {
		return StreamEntry{}, false
	}

	return (*s)[s.Len()-1], true
}

func (s *Stream) Len() int {
	if s == nil {
		return 0
	}
	return len(*s)
}
