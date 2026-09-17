package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type StreamEntry struct {
	ID     string
	Values []string
}

func (se *StreamEntry) DecodeID() (time.Time, int) {
	parts := strings.Split(se.ID, "-")

	ms, _ := strconv.ParseInt(parts[0], 10, 64)
	seq, _ := strconv.Atoi(parts[1])

	return time.UnixMilli(ms), seq
}

type Stream []StreamEntry

func NewStream() *Stream {
	s := make(Stream, 0)
	return &s
}

func (s *Stream) validateOrGenerateID(rawID string) (string, error) {
	lastEntry, exists := s.Last()
	var ID string

	currentTime := time.Now().Truncate(time.Millisecond)
	currentSeqNumb := 0

	// handling "*"
	if rawID == "*" {
		if exists {
			lastMs, lastSeqNumb := lastEntry.DecodeID()
			if lastMs == currentTime {
				currentSeqNumb = lastSeqNumb + 1
			}
		}
		ID = fmt.Sprintf("%d-%d", currentTime.UnixMilli(), currentSeqNumb)
		return ID, nil
	}

	// handling "ms-*"
	parts := strings.Split(rawID, "-")
	msInt, err := strconv.ParseInt(parts[0], 10, 64)

	if err != nil {
		return "", ErrInvalidStreamID
	}

	ms := time.UnixMilli(msInt)

	if parts[1] == "*" {
		if exists {
			lastMs, lastSeqNumb := lastEntry.DecodeID()
			timeCompare := ms.Compare(lastMs)

			if timeCompare == -1 {
				return "", ErrStreamIDSmaller
			} else if timeCompare == 0 {
				currentSeqNumb = lastSeqNumb + 1
			}

		} else if msInt == 0 {
			currentSeqNumb = 1
		}

		ID = fmt.Sprintf("%d-%d", msInt, currentSeqNumb)
		return ID, nil
	}

	// handling "ms-seqNumb"
	seqNumb, err := strconv.Atoi(parts[1])

	if err != nil {
		return "", ErrInvalidStreamID
	}

	if msInt == 0 && seqNumb == 0 {
		return "", ErrStreamIDZero
	}

	if exists {
		lastMs, lastSeqNumb := lastEntry.DecodeID()
		timeCompare := ms.Compare(lastMs)
		if timeCompare == -1 || (timeCompare == 0 && lastSeqNumb >= seqNumb) {
			return "", ErrStreamIDSmaller
		}
	}
	ID = fmt.Sprintf("%d-%d", msInt, seqNumb)
	return ID, nil
}

func (s *Stream) Append(rawID string, values []string) (string, error) {
	ID, err := s.validateOrGenerateID(rawID)

	if err != nil {
		return "", err
	}

	fmt.Println(ID)

	*s = append(*s, StreamEntry{ID: ID, Values: values})

	return ID, nil
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
