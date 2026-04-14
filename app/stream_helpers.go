package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func checkId(stream *Stream, id string) (bool, error) {
	newTimestamp, newSequenceNo, err := parseId(id)

	if err != nil {
		return false, err
	}

	if newTimestamp == 0 && newSequenceNo == 0 {
		return false, errors.New("-ERR The ID specified in XADD must be greater than 0-0\r\n")
	}

	if newTimestamp < stream.lastTimestamp {
		return false, errors.New("-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n")
	}

	if newTimestamp == stream.lastTimestamp && newSequenceNo <= stream.lastSequenceNo {
		return false, errors.New("-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n")
	}

	return true, nil
}

func parseId(id string) (int64, int64, error) {
	parts := strings.Split(id, "-")

	if len(parts) != 2 {
		return 0, 0, errors.New("-ERR Invalid stream ID format\r\n")
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, errors.New("-ERR Failed to convert the timestamp part\r\n")
	}

	sequenceNo, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, errors.New("-ERR Failed to convert the sequence part\r\n")
	}

	return timestamp, sequenceNo, nil
}

func generateId(stream *Stream, rawId string) (string, error) {
	var newTimestamp int64

	if rawId == "*" {
		newTimestamp = time.Now().UnixMilli()
	} else {
		ts, err := strconv.ParseInt(strings.Split(rawId, "-")[0], 10, 64)
		if err != nil {
			return "", errors.New("-ERR Failed to convert the timestamp, invalid input\r\n")
		}

		newTimestamp = ts
	}

	if len(stream.Entries) == 0 {
		if newTimestamp == 0 {
			return fmt.Sprintf("%d-1", newTimestamp), nil
		}
		return fmt.Sprintf("%d-0", newTimestamp), nil
	}

	if newTimestamp < stream.lastTimestamp {
		return "", errors.New("-ERR The timestamp should be higher than the last elem's\r\n")
	}

	if newTimestamp == stream.lastTimestamp {
		return fmt.Sprintf("%d-%d", newTimestamp, (stream.lastSequenceNo + 1)), nil
	}

	return fmt.Sprintf("%d-%d", newTimestamp, 0), nil
}

func getStreamEntries(stream *Stream, minTimestamp, minSequence, maxTimestamp, maxSequence int64) []*StreamEntry {
	var results []*StreamEntry

	for _, entry := range stream.Entries {
		if entry.timestamp < minTimestamp || (entry.timestamp == minTimestamp && entry.sequenceNo < minSequence) {
			continue
		}

		if entry.timestamp > maxTimestamp || (entry.timestamp == maxTimestamp && entry.sequenceNo > maxSequence) {
			break
		}

		results = append(results, entry)
	}
	return results
}

func buildEntriesResp(entries []*StreamEntry) string {
	var sb strings.Builder
	addRespArrayHeader(&sb, len(entries))

	for _, entry := range entries {
		sb.WriteString(returnStreamEntryResp(entry))
	}

	return sb.String()
}

func returnStreamEntryResp(streamEntry *StreamEntry) string {
	var sb strings.Builder

	addRespArrayHeader(&sb, 2)
	addRespString(&sb, streamEntry.ID)
	addRespArrayHeader(&sb, len(streamEntry.Values))

	for _, value := range streamEntry.Values {
		addRespString(&sb, value)
	}

	return sb.String()
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

func checkForBlockingArgs(arr []string) (int64, int, error) {
	var timeout int64 = -1
	var blockIndex int = -1

	for i, v := range arr {
		if strings.ToUpper(v) == "BLOCK" {
			blockIndex = i
			break
		}
	}

	if blockIndex != -1 {
		t, err := strconv.ParseInt(arr[blockIndex+1], 10, 64)
		if err != nil {
			return -1, -1, errors.New("-ERR Invalid timeout value\r\n")
		}
		timeout = t
	}

	streamIndex := -1
	for i, v := range arr {
		if strings.ToUpper(v) == "STREAMS" {
			streamIndex = i
			break
		}
	}

	if streamIndex == -1 {
		return -1, -1, errors.New("-ERR syntax error\r\n")
	}

	return timeout, streamIndex, nil
}
