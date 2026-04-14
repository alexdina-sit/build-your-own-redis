package main

import (
	"bufio"
	"net"
	"sort"
	"sync"
	"time"
)

type Direction string

const CRLF = "\r\n"

const (
	Left  Direction = "left"
	Right Direction = "right"
)

type Item struct {
	Value      string
	ExpireType string
	ExpireTime int
	CreateDate time.Time
}

type User struct {
	Passwords     []string
	Flags         []string
	Authenticated bool
}

type ClientSession struct {
	Connection      net.Conn
	IsAuthenticated bool
	UserName        string
	reader          bufio.Reader
	buf             []byte
	commandsQueue   []string
}

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

type Server struct {
	Role             string
	MasterReplId     string
	MasterReplOffset int
	mu               sync.RWMutex

	itemsMap   map[string]*Item
	usersMap   map[string]*User
	listsMap   map[string][]string
	zsetsMap   map[string]*SortedSet
	streamsMap map[string]*Stream
}

var (
	serverInstance *Server
	once           sync.Once
)

func GetServerInstance() *Server {
	once.Do(
		func() {

			serverInstance = &Server{
				Role:             "master",
				MasterReplId:     "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
				MasterReplOffset: 0,
				itemsMap:         make(map[string]*Item),
				listsMap:         make(map[string][]string),
				usersMap:         make(map[string]*User),
				zsetsMap:         make(map[string]*SortedSet),
				streamsMap:       make(map[string]*Stream),
			}
		})

	return serverInstance
}
