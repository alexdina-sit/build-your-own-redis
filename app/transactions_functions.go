package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func (server *Server) HandleIncr(arr []string) string {
	if len(arr) < 2 {
		return "-Missing arguments. Try INCR <key>"
	}

	server.mu.Lock()
	defer server.mu.Unlock()
	item, prs := server.itemsMap[arr[1]]
	if !prs {
		server.itemsMap[arr[1]] = &Item{
			Value: "1",
		}
		return ":1\r\n"
	}

	value, err := strconv.Atoi(item.Value)
	if err != nil {
		return "-ERR value is not an integer or out of range\r\n"
	}

	value += 1
	server.itemsMap[arr[1]].Value = fmt.Sprintf("%d", value)
	return fmt.Sprintf(":%d\r\n", value)
}

func (server *Server) HandleMulti(session *Session) {
	reader := session.reader
	buf := session.buf
	conn := session.Connection

	_, err := conn.Write([]byte("+OK\r\n"))
	if err != nil {
		fmt.Println("Write error:", err)
	}

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("Read error:", err)
			break
		}

		text := string(buf[:n])
		if strings.Contains(text, "WATCH") {
			_, err = conn.Write([]byte("-ERR WATCH inside MULTI is not allowed\r\n"))
			return
		}

		if strings.Contains(text, "DISCARD") {
			_, err = conn.Write([]byte("+OK\r\n"))
			session.commandsQueue = make([]string, 0)
			session.WatchedKeys = make(map[string]int)
			return
		}

		if strings.Contains(text, "EXEC") {
			server.HandleExec(session)
			return
		}

		session.commandsQueue = append(session.commandsQueue, text)
		_, err = conn.Write([]byte("+QUEUED\r\n"))
	}
}

func (server *Server) HandleExec(session *Session) {
	var sb strings.Builder

	if len(session.commandsQueue) == 0 {
		session.Connection.Write([]byte("*0\r\n"))
		return
	}

	if !changeTracker(server, session) {
		session.Connection.Write([]byte("*-1\r\n"))
		session.WatchedKeys = make(map[string]int)
		session.commandsQueue = make([]string, 0)
		return
	}

	addRespArrayHeader(&sb, len(session.commandsQueue))
	for _, command := range session.commandsQueue {
		sb.WriteString(server.HandleCommand(session, command))
	}

	session.Connection.Write([]byte(sb.String()))
	session.WatchedKeys = make(map[string]int)
	session.commandsQueue = make([]string, 0)
}

func changeTracker(server *Server, session *Session) bool {
	for key, sessionValue := range session.WatchedKeys {
		if serverValue, exists := server.ServerKeys[key]; exists {
			if serverValue != sessionValue {
				return false
			}
		}
	}

	return true
}

func (server *Server) ExecWithoutMulti() string {
	return "-ERR EXEC without MULTI\r\n"
}

func (server *Server) DiscardWithoutMulti() string {
	return "-ERR DISCARD without MULTI\r\n"
}
