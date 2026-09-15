package main

type StreamEntry struct {
	ID     string
	Values []string
}

type Stream struct {
	Entries []StreamEntry
}

func NewStream() *Stream {
	return &Stream{
		Entries: make([]StreamEntry, 0),
	}
}

func (s *Stream) PushBack(ID string, values []string) {
	streamEntry := StreamEntry{ID: ID, Values: values}
	s.Entries = append(s.Entries, streamEntry)
}
