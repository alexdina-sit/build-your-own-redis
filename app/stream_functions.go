package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

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

var smap = make(map[string]*Stream)

func handleType(arr []string) string {
	mu.RLock()
	defer mu.RUnlock()

	if _, exists := mmap[arr[1]]; exists {
		return "+string\r\n"
	}

	if _, exists := smap[arr[1]]; exists {
		return "+stream\r\n"

	}
	return "+none\r\n"
}

func handleXadd(arr []string) string {
	if len(arr) < 5 || (len(arr)-3)%2 != 0 {
		return "-Missing arguments. Try: XADD <stream_key> <id> <key> <value>...\r\n"
	}

	streamKey := arr[1]
	id := arr[2]

	mu.Lock()
	defer mu.Unlock()

	stream, exists := smap[streamKey]
	if !exists {
		stream = &Stream{}
		smap[streamKey] = stream
	}

	if strings.Contains(arr[2], "*") {
		newId, err := generateId(stream, id)

		if err != nil {
			return err.Error()
		}

		id = newId
	} else {
		if isPossible, err := checkId(stream, id); !isPossible {
			return err.Error()
		}
	}

	lastTimestamp, lastSequenceNo, err := parseId(id)
	if err != nil {
		return err.Error()
	}

	streamEntry := StreamEntry{
		ID:         id,
		Values:     arr[3:],
		timestamp:  lastTimestamp,
		sequenceNo: lastSequenceNo,
	}

	stream.Entries = append(stream.Entries, &streamEntry)
	stream.lastSequenceNo = lastSequenceNo
	stream.lastTimestamp = lastTimestamp

	return fmt.Sprintf("$%d\r\n%s\r\n", len(id), id)
}

func handleXrange(arr []string) string {
	if len(arr) < 4 {
		return "-Missing arguments. Please try: XRANGE <stream_key> <start> <stop>\r\n"
	}

	stream_key, startId, stopId := arr[1], arr[2], arr[3]

	mu.RLock()
	defer mu.RUnlock()

	stream, exists := smap[stream_key]
	if !exists {
		return "*0\r\n"
	}

	minTimestamp, minSequence, err := parseLimitId(startId, false)
	if err != nil {
		return err.Error()
	}

	maxTimestamp, maxSequence, err := parseLimitId(stopId, true)
	if err != nil {
		return err.Error()
	}

	filteredEntries := getStreamEntries(stream, minTimestamp, minSequence, maxTimestamp, maxSequence)
	return buildEntriesResp(filteredEntries)
}

func parseLimitId(id string, isMax bool) (int64, int64, error) {
	if id == "-" {
		return 0, 0, nil
	}

	if id == "+" {
		return math.MaxInt64, math.MaxInt64, nil
	}

	parts := strings.Split(id, "-")
	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, errors.New("-ERR Failed to convert the miliseconds part\r\n")
	}

	if len(parts) == 1 {
		if isMax {
			return timestamp, math.MaxInt64, nil
		}
		return timestamp, 0, nil
	}

	sequence, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, errors.New("-ERR Failed to convert the sequence part\r\n")
	}

	return timestamp, sequence, nil
}

func returnStreamEntryResp(streamEntry *StreamEntry) string {
	returnStr := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n*%d\r\n", len(streamEntry.ID), streamEntry.ID, len(streamEntry.Values))
	for _, value := range streamEntry.Values {
		returnStr += fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	}

	return returnStr
}

func handleXread(arr []string) string {
	if len(arr) < 4 {
		return "-Missing arguments. Please try: XREAD STREAMS <key> <id>\r\n"
	}

	stream_key, startId := arr[2], arr[3]

	mu.RLock()
	defer mu.RUnlock()

	stream, exists := smap[stream_key]
	if !exists {
		return "*0\r\n"
	}

	parts := strings.Split(startId, "-")
	if len(parts) < 2 {
		return "-ERR Invalid ID, it should contains the timestamp and the sequence number.\r\n"
	}

	sequence, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "-ERR Failed to convert the sequence part\r\n"
	}

	startId = fmt.Sprintf("%s-%d", parts[0], sequence+1)
	stopId := fmt.Sprintf("%d-%d", stream.lastTimestamp, stream.lastSequenceNo)
	return fmt.Sprintf("*1\r\n*2\r\n$%d\r\n%s\r\n", len(stream_key), stream_key) + handleXrange([]string{"XRANGE", stream_key, startId, stopId})
}
