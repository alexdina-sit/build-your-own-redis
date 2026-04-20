package main

import (
	"fmt"
	"strings"
	"time"
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

func (server *Server) XreadLastCase(ids, keys []string) []string {
	server.mu.RLock()
	defer server.mu.RUnlock()

	newIds := make([]string, len(ids))
	for i, id := range ids {
		if id == "$" {
			stream, exists := server.streamsMap[keys[i]]
			if exists && len(stream.Entries) > 0 {
				newIds[i] = fmt.Sprintf("%d-%d", stream.lastTimestamp, stream.lastSequenceNo)
			} else {
				newIds[i] = "0-0"
			}
		} else {
			newIds[i] = id
		}
	}

	return newIds
}

func (server *Server) HandleType(arr []string) string {
	server.mu.RLock()
	defer server.mu.RUnlock()

	if _, exists := server.itemsMap[arr[1]]; exists {
		return "+string\r\n"
	}

	if _, exists := server.streamsMap[arr[1]]; exists {
		return "+stream\r\n"

	}
	return "+none\r\n"
}

func (server *Server) HandleXadd(arr []string) string {
	if len(arr) < 5 || (len(arr)-3)%2 != 0 {
		return "-Missing arguments. Try: XADD <stream_key> <id> <key> <value>...\r\n"
	}

	streamKey := arr[1]
	id := arr[2]

	server.mu.Lock()
	defer server.mu.Unlock()

	stream, exists := server.streamsMap[streamKey]
	if !exists {
		stream = &Stream{}
		server.streamsMap[streamKey] = stream
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

func (server *Server) HandleXrange(arr []string) string {
	if len(arr) < 4 {
		return "-Missing arguments. Please try: XRANGE <stream_key> <start> <stop>\r\n"
	}

	stream_key, startId, stopId := arr[1], arr[2], arr[3]

	server.mu.RLock()
	defer server.mu.RUnlock()

	stream, exists := server.streamsMap[stream_key]
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

func (server *Server) handleXread(arr []string) string {
	timeout, streamIndex, err := checkForBlockingArgs(arr)
	if err != nil {
		return err.Error()
	}

	arguments := arr[streamIndex+1:]
	if len(arr) < 4 {
		return "-Missing arguments. Please try: XREAD STREAMS <key_1> <key_2> <id_1> <id_2>\r\n"
	}

	numStreams := len(arguments) / 2
	keys := arguments[:numStreams]
	ids := server.XreadLastCase(arguments[numStreams:], keys)

	now := time.Now()
	for {
		var sb strings.Builder
		streamsWithDataCount := 0

		for index, stream_key := range keys {
			server.mu.RLock()

			stream, exists := server.streamsMap[stream_key]
			if !exists {
				server.mu.RUnlock()
				continue
			}

			startId := ids[index]
			parts := strings.Split(startId, "-")

			if len(parts) < 2 {
				server.mu.RUnlock()
				return "-ERR Invalid ID format\r\n"
			}

			timestamp, sequence, err := parseId(startId)
			if err != nil {
				return err.Error()
			}

			entries := getStreamEntries(stream, timestamp, (sequence + 1), stream.lastTimestamp, stream.lastSequenceNo)
			server.mu.RUnlock()

			if len(entries) > 0 {
				streamsWithDataCount++
				addRespArrayHeader(&sb, 2)
				addRespString(&sb, stream_key)
				sb.WriteString(buildEntriesResp(entries))
			}
		}

		if streamsWithDataCount > 0 {
			var finalResp strings.Builder
			addRespArrayHeader(&finalResp, streamsWithDataCount)
			finalResp.WriteString(sb.String())
			return finalResp.String()
		}

		if (timeout == -1) || (timeout > 0 && time.Since(now).Milliseconds() >= timeout) {
			return "*-1\r\n"
		}
		time.Sleep(10 * time.Millisecond)
	}
}
