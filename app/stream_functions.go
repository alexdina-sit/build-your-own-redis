package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type StreamEntry struct {
	ID     string
	Values map[string]string
}

type Stream struct {
	Entries []*StreamEntry
}

var smap = make(map[string]*Stream)

func handleType(arr []string) string {
	mu.RLock()
	defer mu.RUnlock()

	_, prs := mmap[arr[1]]
	if !prs {
		_, prss := smap[arr[1]]
		if !prss {
			return "+none\r\n"
		}
		return "+stream\r\n"
	}

	return "+string\r\n"
}

func handleXadd(arr []string) string {
	if len(arr) < 5 {
		return "-Missing arguments. Try: XADD <stream_key> <key> <value>..."
	}

	mu.Lock()
	defer mu.Unlock()

	stream, prs := smap[arr[1]]
	if !prs {
		stream = &Stream{}
		smap[arr[1]] = stream
	}

	errMsg := ""
	id := arr[2]
	if strings.Contains(arr[2], "*") {
		id, errMsg = generateId(stream.Entries, id)

		if errMsg != "" {
			return errMsg
		}
	} else {
		isPossible, errMsg := checkId(stream.Entries, id)
		if !isPossible {
			return errMsg
		}
	}

	streamEntry := StreamEntry{
		ID:     id,
		Values: make(map[string]string),
	}

	for i := 3; i < len(arr); i += 2 {
		streamEntry.Values[arr[i]] = arr[i+1]
	}

	stream.Entries = append(stream.Entries, &streamEntry)
	return fmt.Sprintf("$%d\r\n%s\r\n", len(id), id)
}

func generateId(entries []*StreamEntry, rawId string) (string, string) {
	var timeMs int64

	if rawId == "*" {
		timeMs = time.Now().UnixMilli()
	} else {
		ts, err := strconv.ParseInt(strings.Split(rawId, "-")[0], 10, 64)
		if err != nil {
			return "", "-ERR Failed to convert the miliseconds, invalid input\r\n"
		}

		timeMs = ts
	}

	if len(entries) == 0 {
		if timeMs == 0 {
			return fmt.Sprintf("%d-1", timeMs), ""
		}
		return fmt.Sprintf("%d-0", timeMs), ""
	}

	lastElem := entries[len(entries)-1]
	msTime, seqNo, errMsg := parseId(lastElem.ID)

	if errMsg != "" {
		return "", errMsg
	}

	if timeMs == msTime {
		return fmt.Sprintf("%d-%d", timeMs, (seqNo + 1)), ""
	}

	if timeMs < msTime {
		return "", "-ERR Miliseconds part should be higher than the last elem's\r\n"
	}
	return fmt.Sprintf("%d-%d", timeMs, 0), ""
}

func checkId(entries []*StreamEntry, id string) (bool, string) {
	newMsTime, newSeqNo, errMsg := parseId(id)

	if errMsg != "" {
		return false, errMsg
	}

	if newMsTime == 0 && newSeqNo == 0 {
		return false, "-ERR The ID specified in XADD must be greater than 0-0\r\n"
	}

	if len(entries) == 0 {
		return true, ""
	}

	lastElem := entries[len(entries)-1]
	msTime, seqNo, errMsg := parseId(lastElem.ID)

	if errMsg != "" {
		return false, errMsg
	}

	if newMsTime < msTime {
		return false, "-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n"
	}

	if newMsTime == msTime {
		if newSeqNo > seqNo {
			return true, ""
		}
		return false, "-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n"
	}

	return true, ""
}

func parseId(id string) (int64, int64, string) {
	parts := strings.Split(id, "-")

	if len(parts) != 2 {
		return 0, 0, "-ERR Invalid stream ID format\r\n"
	}

	msTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, "-ERR Failed to convert the miliseconds part\r\n"
	}

	seqNo, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, "-ERR Failed to convert the sequence part\r\n"
	}

	return msTime, seqNo, ""
}
