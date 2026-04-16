package main

type StreamEntry struct {
	ID         string
	Values     []string
	timestamp  int64
	sequenceNo int64
}

type Stream struct {
	Entries        []*StreamEntry
	lastTimestamp  int64
	lastSequenceNo int64
}
