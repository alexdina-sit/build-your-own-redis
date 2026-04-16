package main

import (
	"bufio"
	"net"
)

type Session struct {
	Connection         net.Conn
	IsAuthenticated    bool
	UserName           string
	reader             bufio.Reader
	buf                []byte
	commandsQueue      []string
	IsReplica          bool
	ReplOffset         int
	SubscribedChannels map[string]bool
}

func NewSession(conn net.Conn, reader *bufio.Reader, buf []byte) *Session {
	return &Session{
		Connection:         conn,
		IsAuthenticated:    false,
		UserName:           "default",
		reader:             *reader,
		buf:                buf,
		SubscribedChannels: make(map[string]bool),
	}
}
