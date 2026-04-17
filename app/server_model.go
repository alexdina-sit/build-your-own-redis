package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Role             string
	MasterReplId     string
	MasterReplOffset int
	mu               sync.RWMutex
	rdb              *Rdb
	aof              *AOF

	replicas []*Session

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
				rdb:              &Rdb{},
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

func (server *Server) handlePing(session *Session) string {
	if len(session.SubscribedChannels) > 0 {
		return "*2\r\n$4\r\npong\r\n$0\r\n\r\n"
	}
	return "+PONG\r\n"
}

func (server *Server) execWithoutMulti() string {
	return "-ERR EXEC without MULTI\r\n"
}

func (server *Server) discardWithoutMulti() string {
	return "-ERR DISCARD without MULTI\r\n"
}

func (server *Server) handleEcho(args []string) string {
	if len(args) < 2 {
		return "-ERR Missing arguments. Please try: ECHO <text>\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])
}

func (server *Server) handleSet(args []string) string {
	var expireType string
	var expireTime int

	if len(args) < 3 {
		return "-ERR Missing arguments. Your input should be: SET <key> <value>\r\n"
	}

	if len(args) > 4 {
		expireType = strings.ToUpper(args[3])
		intValue, err := strconv.Atoi(args[4])
		if err != nil {
			return "-ERR Error while converting your expire value\r\n"
		}

		expireTime = intValue
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	server.itemsMap[args[1]] = &Item{
		Value:      args[2],
		ExpireType: expireType,
		ExpireTime: expireTime,
		CreateDate: time.Now(),
	}

	server.ServerKeys[args[1]]++
	return "+OK\r\n"
}

func (server *Server) handleGet(arr []string) string {
	server.mu.Lock()
	defer server.mu.Unlock()

	item, prs := server.itemsMap[arr[1]]
	if !prs {
		return "$-1\r\n"
	}

	expireType := strings.ToUpper(item.ExpireType)
	timeSpent := time.Since(item.CreateDate)

	if (expireType == "EX" && timeSpent.Seconds() > float64(item.ExpireTime)) ||
		(expireType == "PX" && timeSpent.Milliseconds() > int64(item.ExpireTime)) {

		delete(server.itemsMap, arr[1])
		return "$-1\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(item.Value), item.Value)
}
