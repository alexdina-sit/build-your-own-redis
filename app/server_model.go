package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strings"
	"sync"
)

type Server struct {
	Role             string
	MasterReplId     string
	MasterReplOffset int
	Config           *Config
	aof              *os.File
	isAofLoading     bool

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

func (server *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	reader, buf := bufio.NewReader(conn), make([]byte, 1024)
	session := NewSession(conn, reader, buf)

	server.mu.RLock()
	defaultUser, exists := server.usersMap["default"]
	if exists {
		for _, flag := range defaultUser.Flags {
			if flag == "nopass" {
				session.IsAuthenticated = true
			}
		}
	}
	server.mu.RUnlock()

	for {
		cmd, err := readRespCommand(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Read error:", err)
			break
		}

		response := server.HandleCommand(session, string(cmd))
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Write error:", err)
		}
	}
}

func (server *Server) HandleCommand(session *Session, input string) string {
	if len(input) == 0 || input[0] != '*' {
		return ""
	}

	args := processRespArray(input)
	command := strings.ToUpper(args[0])

	if !session.IsAuthenticated && args[0] != "AUTH" {
		return "-NOAUTH Authentication required\r\n"
	}

	if len(session.SubscribedChannels) > 0 && !slices.Contains(allowedCommandsSubscribed, command) {
		return fmt.Sprintf("-ERR Can't execute '%s' while subscribed to a channel\r\n", command)
	}

	return server.ExecuteCommand(session, command, args, input)
}

func (server *Server) ExecuteCommand(session *Session, command string, args []string, input string) string {
	fnc, exists := server.commandsMap[command]
	if !exists {
		return "-ERR Unknown comannd\r\n"
	}

	if command == "PSYNC" || command == "MULTI" {
		fnc(command, session, args)
		return ""
	}

	response := fnc(command, session, args)
	if !slices.Contains(replicableCommands, command) {
		return response
	}

	if server.Role == "master" {
		server.mu.Lock()
		defer server.mu.Unlock()

		server.MasterReplOffset += len([]byte(input))
		propagate(server.replicas, input)
	}

	if strings.ToLower(server.Config.AppendOnly) == "yes" && strings.ToLower(server.Config.AppendFSync) == "always" {
		if !server.isAofLoading {
			writeToAof(server.aof, input)
		}
	}
	return response

}

func (server *Server) HandlePing(session *Session) string {
	if len(session.SubscribedChannels) > 0 {
		return "*2\r\n$4\r\npong\r\n$0\r\n\r\n"
	}
	return "+PONG\r\n"
}

func (server *Server) HandleEcho(args []string) string {
	if len(args) < 2 {
		return "-ERR Missing arguments. Please try: ECHO <text>\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])
}
