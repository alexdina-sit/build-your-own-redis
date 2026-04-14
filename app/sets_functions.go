package main

import (
	"fmt"
	"strconv"
	"strings"
)

func (server *Server) handleZadd(arr []string) string {
	if len(arr) < 4 || (len(arr)-2)%2 != 0 {
		return "-ERR syntax error. Please try: ZADD <zset_key> <score> <member>\r\n"
	}

	setKey := arr[1]

	server.mu.Lock()
	defer server.mu.Unlock()

	set, prs := server.zsetsMap[setKey]
	if !prs {
		set = NewSortedSet()
		server.zsetsMap[setKey] = set
	}

	addedCount := 0

	for i := 2; i < len(arr); i += 2 {
		score, err := strconv.ParseFloat(arr[i], 64)
		if err != nil {
			return "-ERR value is not a valid float\r\n"
		}

		member := arr[i+1]
		addedCount += set.AddOrUpdate(member, score)
	}
	return fmt.Sprintf(":%d\r\n", addedCount)
}

func (server *Server) handleZrank(arr []string) string {
	if len(arr) < 3 {
		return "-ERR syntax error. Please try: ZRANK <zset_key> <zset_member>\r\n"
	}

	zsetKey, zsetMember := arr[1], arr[2]

	server.mu.RLock()
	defer server.mu.RUnlock()

	set, exists := server.zsetsMap[zsetKey]
	if !exists {
		return "$-1\r\n"
	}

	memberRank, memberExists := set.GetMemberRank(zsetMember)
	if !memberExists {
		return "$-1\r\n"
	}

	return fmt.Sprintf(":%d\r\n", memberRank)
}

func (server *Server) handleZcard(arr []string) string {
	if len(arr) < 2 {
		return "-ERR Invalid syntax. Please try: ZCARD <zset_key>"
	}

	server.mu.RLock()
	defer server.mu.RUnlock()

	zsetKey := arr[1]
	set, exists := server.zsetsMap[zsetKey]
	if !exists {
		return ":0\r\n"
	}

	return fmt.Sprintf(":%d\r\n", len(set.elements))
}

func (server *Server) handleZrange(arr []string) string {
	if len(arr) < 4 {
		return "-ERR Syntax error. Please try: ZRANG <zset_key> <start> <stop>\r\n"
	}

	zsetKey := arr[1]
	startIndex, err1 := strconv.Atoi(arr[2])
	stopIndex, err2 := strconv.Atoi(arr[3])

	if err1 != nil || err2 != nil {
		return "-ERR Invalid start / stop indexes. Please try again with integer values\r\n"
	}

	server.mu.RLock()
	defer server.mu.RUnlock()

	set, exists := server.zsetsMap[zsetKey]
	if !exists {
		return "*0\r\n"
	}

	lenSetElem := len(set.elements)
	if startIndex < 0 {
		if -startIndex > lenSetElem {
			startIndex = 0
		} else {
			startIndex = lenSetElem + startIndex
		}
	}

	if stopIndex < 0 {
		if -stopIndex > lenSetElem {
			stopIndex = 0
		} else {
			stopIndex = lenSetElem + stopIndex
		}
	}

	if startIndex >= lenSetElem || startIndex > stopIndex {
		return "*0\r\n"
	}

	if stopIndex >= lenSetElem {
		stopIndex = lenSetElem - 1
	}

	var sb strings.Builder
	addRespArrayHeader(&sb, (stopIndex - startIndex + 1))
	for i := startIndex; i <= stopIndex; i++ {
		addRespString(&sb, set.elements[i].Member)
	}

	return sb.String()
}

func (server *Server) handleZscore(arr []string) string {
	if len(arr) < 3 {
		return "-ERR Invalid syntax. Please try: ZSCORE <zset_key> <zset_member>"
	}

	server.mu.RLock()
	defer server.mu.RUnlock()

	zsetKey, zsetMember := arr[1], arr[2]
	set, exists := server.zsetsMap[zsetKey]
	if !exists {
		return "$-1\r\n"
	}

	score, memberExists := set.GetMemberScore(zsetMember)
	if !memberExists {
		return "$-1\r\n"
	}

	scoreStr := strconv.FormatFloat(score, 'f', -1, 64)
	return fmt.Sprintf("$%d\r\n%s\r\n", len(scoreStr), scoreStr)
}

func (server *Server) handleZrem(arr []string) string {
	if len(arr) < 3 {
		return "-ERR Invalid syntax. Please try: ZREM <zset_key> <zset_member1>..\r\n"
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	zsetKey := arr[1]
	set, exists := server.zsetsMap[zsetKey]
	if !exists {
		return ":0\r\n"
	}

	removedMembersCount := 0
	for _, member := range arr[2:] {
		if set.RemoveMember(member) != -1 {
			removedMembersCount++
		}
	}

	if len(set.elements) == 0 {
		delete(server.zsetsMap, zsetKey)
	}

	return fmt.Sprintf(":%d\r\n", removedMembersCount)
}
