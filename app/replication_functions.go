package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func getMasterAddress(replicaofArgs string) (string, string, error) {
	replicaParts := strings.Split(replicaofArgs, " ")
	if len(replicaParts) < 2 {
		return "", "", errors.New("-ERR Missing arguments, you should mention both MASTER_HOST and MASTER_PORT")
	}

	masterHost, masterPort := replicaParts[0], replicaParts[1]
	_, err := strconv.Atoi(masterPort)
	if err != nil {
		return "", "", errors.New("-ERR Failed to convert the MASTER_PORT. Please try again with an integer value")
	}

	return masterHost, masterPort, nil
}

func (server *Server) handleInfo(arr []string) string {
	var sb strings.Builder

	if len(arr) < 1 {
		return ""
	}

	if arr[1] == "replication" {
		response := fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d\r\n", server.Role, server.MasterReplId, server.MasterReplOffset)
		addRespString(&sb, response)
		return sb.String()
	}

	return ""
}

func handshake(masterHost string, masterPort string) {
	masterAddress := net.JoinHostPort(masterHost, masterPort)
	conn, err := net.Dial("tcp", masterAddress)
	if err != nil {
		fmt.Println("Failed to communicate with the master")
		return
	}

	defer conn.Close()
	reader := bufio.NewReader(conn)
	buf := make([]byte, 1024)

	// Handshake #1 - Send PING
	conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	_, err = reader.Read(buf)
	if err != nil {
		fmt.Println("Failed to receive a response for PING")
		return
	}

	// Handshake #2 - Send REPLCONF #1
	conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n"))
	_, err = reader.Read(buf)
	if err != nil {
		fmt.Println("Failed to receive a response for REPLCONF #1")
		return
	}

	// Handshake #2 - Send REPLCONF #2
	conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
	_, err = reader.Read(buf)
	if err != nil {
		fmt.Println("Failed to receive a response for REPLCONF #2")
		return
	}

	// Handshake #3 - Send PSYNC
	conn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
	_, err = reader.Read(buf)
	if err != nil {
		fmt.Println("Failed to receive a response for REPSYNC")
		return
	}
}

func (server *Server) handlePsync(session *ClientSession) {
	emptyRdbHex := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0dfb8ea"
	rdbBytes, _ := hex.DecodeString(emptyRdbHex)

	resyncMessage := fmt.Sprintf("+FULLRESYNC %s %d\r\n", server.MasterReplId, server.MasterReplOffset)
	session.Connection.Write([]byte(resyncMessage))

	rdbMessage := fmt.Sprintf("$%d\r\n%s", len(rdbBytes), rdbBytes)
	session.Connection.Write([]byte(rdbMessage))
}

func (server *Server) handleReplconf() string {
	return "+OK\r\n"
}
