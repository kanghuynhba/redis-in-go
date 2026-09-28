package main

import (
	"cmp"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type StreamID struct {
	MilliSec  int64
	SeqNumber int
}

func (sid StreamID) String() string {
	return fmt.Sprintf("%d-%d", sid.MilliSec, sid.SeqNumber)
}

func (sid StreamID) Compare(other StreamID) int {
	if c := cmp.Compare(sid.MilliSec, other.MilliSec); c != 0 {
		return c
	}

	return cmp.Compare(sid.SeqNumber, other.SeqNumber)
}

type StreamEntry struct {
	ID     StreamID
	Values []string
}

func (se StreamEntry) EncodeRESP() []byte {
	entryArr := NewEncodeArray()
	entryArr.Append(EncodeBulkString(se.ID.String()))

	valArr := NewEncodeArray()
	for _, val := range se.Values {
		valArr.Append(EncodeBulkString(val))
	}
	entryArr.Append(valArr.Bytes())

	return entryArr.Bytes()
}

type Stream []StreamEntry

func NewStream() *Stream {
	s := make(Stream, 0)
	return &s
}

func (s *Stream) validateOrGenerateID(rawID string) (StreamID, error) {
	lastEntry, exists := s.Last()
	lastID := lastEntry.ID

	currentTime := time.Now().UnixMilli()
	currentSeqNumb := 0

	// handling "*"
	if rawID == "*" {
		if exists && lastID.MilliSec == currentTime {
			currentSeqNumb = lastID.SeqNumber + 1
		}
		return StreamID{MilliSec: currentTime, SeqNumber: currentSeqNumb}, nil
	}

	// handling "ms-*"
	parts := strings.Split(rawID, "-")

	if len(parts) != 2 {
		return StreamID{}, ErrInvalidStreamID
	}

	ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return StreamID{}, ErrInvalidStreamID
	}

	if parts[1] == "*" {
		if exists {
			if ms < lastID.MilliSec {
				return StreamID{}, ErrStreamIDSmaller
			} else if ms == lastID.MilliSec {
				currentSeqNumb = lastID.SeqNumber + 1
			}
		} else if ms == 0 {
			currentSeqNumb = 1
		}

		return StreamID{MilliSec: ms, SeqNumber: currentSeqNumb}, nil
	}

	// handling "ms-seqNumb"
	seqNumb, err := strconv.Atoi(parts[1])
	if err != nil {
		return StreamID{}, ErrInvalidStreamID
	}

	if ms == 0 && seqNumb == 0 {
		return StreamID{}, ErrStreamIDZero
	}

	if exists {
		if ms < lastID.MilliSec || (ms == lastID.MilliSec && lastID.SeqNumber >= seqNumb) {
			return StreamID{}, ErrStreamIDSmaller
		}
	}

	return StreamID{MilliSec: ms, SeqNumber: seqNumb}, nil
}

func (s *Stream) Append(rawID string, values []string) (string, error) {
	id, err := s.validateOrGenerateID(rawID)
	if err != nil {
		return "", err
	}

	*s = append(*s, StreamEntry{ID: id, Values: values})

	return id.String(), nil
}

func (s *Stream) Last() (StreamEntry, bool) {
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

func (s *Stream) LowerBound(id StreamID) int {
	left, right := 0, s.Len()
	for left < right {
		mid := left + (right-left)/2
		if (*s)[mid].ID.Compare(id) >= 0 {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

func (s *Stream) UpperBound(id StreamID) int {
	left, right := 0, s.Len()
	for left < right {
		mid := left + (right-left)/2
		if (*s)[mid].ID.Compare(id) > 0 {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

func (s *Stream) validateRangeID(rawID string) (StreamID, error) {

	if rawID == "+" {
		lastEntry, exists := s.Last()
		if !exists {
			return StreamID{}, ErrInvalidStreamID
		}
		return lastEntry.ID, nil
	}

	parts := strings.Split(rawID, "-")

	if len(parts) != 2 {
		return StreamID{}, ErrInvalidStreamID
	}

	ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return StreamID{}, ErrInvalidStreamID
	}

	seqNumb, err := strconv.Atoi(parts[1])
	if err != nil {
		return StreamID{}, ErrInvalidStreamID
	}

	return StreamID{MilliSec: ms, SeqNumber: seqNumb}, nil
}

func (s *Stream) Range(rawStartID, rawEndID string) ([]StreamEntry, error) {

	start, err := s.validateRangeID(rawStartID)
	end, err := s.validateRangeID(rawEndID)

	if err != nil {
		return nil, err
	}

	startIdx := s.LowerBound(start)
	endIdx := s.UpperBound(end)
	return (*s)[startIdx:endIdx], nil
}

func (s *Stream) ReadAfter(rawID string) (Stream, error) {
	id, err := s.validateRangeID(rawID, false)

	if err != nil {
		return nil, err
	}

	idx := math.Max((float64(s.UpperBound(id) - 1)), 0)

	return (*s)[int(idx):], nil
}
