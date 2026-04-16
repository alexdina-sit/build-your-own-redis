package main

import "sort"

type Element struct {
	Member string
	Score  float64
}

type SortedSet struct {
	ElemSlice []*Element
	ElemDict  map[string]*Element
}

func NewSortedSet() *SortedSet {
	return &SortedSet{
		ElemSlice: make([]*Element, 0),
		ElemDict:  make(map[string]*Element),
	}
}

func (s *SortedSet) AddOrUpdate(member string, score float64) int {
	returnValue := 0

	if existingElem, exists := s.ElemDict[member]; exists {
		if existingElem.Score == score {
			return 0
		}

		for i, e := range s.ElemSlice {
			if e.Member == member {
				s.ElemSlice = append(s.ElemSlice[:i], s.ElemSlice[i+1:]...)
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
	s.ElemDict[member] = newElem

	index := sort.Search(len(s.ElemSlice), func(i int) bool {
		if s.ElemSlice[i].Score == score {
			return s.ElemSlice[i].Member >= member
		}
		return s.ElemSlice[i].Score > score
	})

	s.ElemSlice = append(s.ElemSlice, nil)
	copy(s.ElemSlice[index+1:], s.ElemSlice[index:])
	s.ElemSlice[index] = newElem

	return returnValue
}

func (s *SortedSet) GetMemberRank(member string) (int, bool) {
	for index, elem := range s.ElemSlice {
		if elem.Member == member {
			return index, true
		}
	}

	return 0, false
}

func (s *SortedSet) GetMemberScore(member string) (float64, bool) {
	elem, exists := s.ElemDict[member]
	if !exists {
		return 0, false
	}

	return elem.Score, true
}

func (s *SortedSet) RemoveMember(member string) int {
	_, exists := s.ElemDict[member]
	if !exists {
		return -1
	}

	delete(s.ElemDict, member)

	for index, elem := range s.ElemSlice {
		if elem.Member == member {
			s.ElemSlice = append(s.ElemSlice[:index], s.ElemSlice[index+1:]...)
			return 1
		}
	}
	return -1
}
