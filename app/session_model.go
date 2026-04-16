package main

import (
	"bufio"
	"net"
)

type Session struct {
	Connection      net.Conn
	IsAuthenticated bool
	UserName        string
	reader          bufio.Reader
	buf             []byte
	commandsQueue   []string
	IsReplica       bool
	ReplOffset      int
}
