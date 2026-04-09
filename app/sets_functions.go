package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Element struct {
	Member string
	Score  float64
}

type SortedSet struct {
	elements []*Element
	dict     map[string]*Element
}

func NewSortedSet() *SortedSet {
	return &SortedSet{
		elements: make([]*Element, 0),
		dict:     make(map[string]*Element),
	}
}

func (s *SortedSet) AddOrUpdate(member string, score float64) int {
	returnValue := 0

	if existingElem, exists := s.dict[member]; exists {
		if existingElem.Score == score {
			return 0
		}

		for i, e := range s.elements {
			if e.Member == member {
				s.elements = append(s.elements[:i], s.elements[i+1:]...)
				break
			}
		}
	} else {
		returnValue = 1
	}

	newElem := &Element{
		Member: member,
		Score:  score,
	}
	s.dict[member] = newElem

	index := sort.Search(len(s.elements), func(i int) bool {
		if s.elements[i].Score == score {
			return s.elements[i].Member >= member
		}
		return s.elements[i].Score > score
	})

	s.elements = append(s.elements, nil)
	copy(s.elements[index+1:], s.elements[index:])
	s.elements[index] = newElem

	return returnValue
}

func (s *SortedSet) GetMemberRank(member string) (int, bool) {
	for index, elem := range s.elements {
		if elem.Member == member {
			return index, true
		}
	}

	return 0, false
}

func (s *SortedSet) GetMemberScore(member string) (float64, bool) {
	elem, exists := s.dict[member]
	if !exists {
		return 0, false
	}

	return elem.Score, true
}

func (s *SortedSet) RemoveMember(member string) int {
	_, exists := s.dict[member]
	if !exists {
		return -1
	}

	delete(s.dict, member)

	for index, elem := range s.elements {
		if elem.Member == member {
			s.elements = append(s.elements[:index], s.elements[index+1:]...)
			return 1
		}
	}
	return -1
}

var setMap = make(map[string]*SortedSet)

func handleZadd(arr []string) string {
	if len(arr) < 4 || (len(arr)-2)%2 != 0 {
		return "-ERR syntax error. Please try: ZADD <zset_key> <score> <member>\r\n"
	}

	setKey := arr[1]

	mu.Lock()
	defer mu.Unlock()

	set, prs := setMap[setKey]
	if !prs {
		set = NewSortedSet()
		setMap[setKey] = set
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

func handleZrank(arr []string) string {
	if len(arr) < 3 {
		return "-ERR syntax error. Please try: ZRANK <zset_key> <zset_member>\r\n"
	}

	zsetKey, zsetMember := arr[1], arr[2]

	mu.RLock()
	defer mu.RUnlock()

	set, exists := setMap[zsetKey]
	if !exists {
		return "$-1\r\n"
	}

	memberRank, memberExists := set.GetMemberRank(zsetMember)
	if !memberExists {
		return "$-1\r\n"
	}

	return fmt.Sprintf(":%d\r\n", memberRank)
}

func handleZcard(arr []string) string {
	if len(arr) < 2 {
		return "-ERR Invalid syntax. Please try: ZCARD <zset_key>"
	}

	mu.RLock()
	defer mu.RUnlock()

	zsetKey := arr[1]
	set, exists := setMap[zsetKey]
	if !exists {
		return ":0\r\n"
	}

	return fmt.Sprintf(":%d\r\n", len(set.elements))
}

func handleZrange(arr []string) string {
	if len(arr) < 4 {
		return "-ERR Syntax error. Please try: ZRANG <zset_key> <start> <stop>\r\n"
	}

	zsetKey := arr[1]
	startIndex, err1 := strconv.Atoi(arr[2])
	stopIndex, err2 := strconv.Atoi(arr[3])

	if err1 != nil || err2 != nil {
		return "-ERR Invalid start / stop indexes. Please try again with integer values\r\n"
	}

	mu.RLock()
	defer mu.RUnlock()

	set, exists := setMap[zsetKey]
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

func handleZscore(arr []string) string {
	if len(arr) < 3 {
		return "-ERR Invalid syntax. Please try: ZSCORE <zset_key> <zset_member>"
	}

	mu.RLock()
	defer mu.RUnlock()

	zsetKey, zsetMember := arr[1], arr[2]
	set, exists := setMap[zsetKey]
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

func handleZrem(arr []string) string {
	if len(arr) < 3 {
		return "-ERR Invalid syntax. Please try: ZREM <zset_key> <zset_member1>..\r\n"
	}

	mu.Lock()
	defer mu.Unlock()

	zsetKey := arr[1]
	set, exists := setMap[zsetKey]
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
		delete(setMap, zsetKey)
	}

	return fmt.Sprintf(":%d\r\n", removedMembersCount)
}
