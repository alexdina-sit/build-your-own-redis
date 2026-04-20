package main

import (
	"os"
	"sync"
)

type Server struct {
	Role             string
	MasterReplId     string
	MasterReplOffset int
	Config           *Config
	aof              *os.File

	mu sync.RWMutex

	replicas    []*Session
	itemsMap    map[string]*Item
	usersMap    map[string]*User
	listsMap    map[string][]string
	zsetsMap    map[string]*SortedSet
	streamsMap  map[string]*Stream
	commandsMap map[string]func(input string, session *Session, args []string) string
	channels    map[string][]*Session
	ServerKeys  map[string]int
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
				commandsMap:      make(map[string]func(input string, session *Session, args []string) string),
				channels:         make(map[string][]*Session),
				ServerKeys:       make(map[string]int),
			}

			serverInstance.initCommands()
		})
	return serverInstance
}
