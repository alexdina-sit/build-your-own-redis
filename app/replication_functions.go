package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

func getMasterAddress(replicaofArgs string) (string, string, error) {
	replicaParts := strings.Split(replicaofArgs, " ")
	if len(replicaParts) < 2 {
		return "", "", errors.New("-ERR Missing arguments, you should mention both MASTER_HOST and MASTER_PORT\r\n")
	}

	masterHost, masterPort := replicaParts[0], replicaParts[1]
	_, err := strconv.Atoi(masterPort)
	if err != nil {
		return "", "", errors.New("-ERR Failed to convert the MASTER_PORT. Please try again with an integer value\r\n")
	}

	return masterHost, masterPort, nil
}

func (server *Server) HandleInfo(arr []string) string {
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

func (server *Server) Handshake(masterHost string, masterPort string) {
	masterAddress := net.JoinHostPort(masterHost, masterPort)
	conn, err := net.Dial("tcp", masterAddress)
	if err != nil {
		fmt.Println("Failed to communicate with the master")
		return
	}

	reader := bufio.NewReader(conn)

	conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	reader.ReadString('\n')

	conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n"))
	reader.ReadString('\n')

	conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
	reader.ReadString('\n')

	conn.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
	reader.ReadString('\n')

	line, err := reader.ReadString('\n')
	rdbLen, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		fmt.Println("Failed to convert the RDB size")
	}

	buf := make([]byte, rdbLen)
	io.ReadFull(reader, buf)

	session := &Session{
		Connection:      conn,
		IsReplica:       true,
		IsAuthenticated: true,
	}

	for {
		cmd, err := readRespCommand(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err.Error())
			}
			break
		}

		response := server.HandleCommand(session, cmd)
		if strings.Contains(response, "REPLCONF") {
			conn.Write([]byte(response))
		}

		session.ReplOffset += len([]byte(cmd))
	}
}

func (server *Server) HandlePsync(session *Session) string {
	emptyRdbHex := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0dfb8ea"
	rdbBytes, _ := hex.DecodeString(emptyRdbHex)

	resyncMessage := fmt.Sprintf("+FULLRESYNC %s %d\r\n", server.MasterReplId, server.MasterReplOffset)
	session.Connection.Write([]byte(resyncMessage))

	session.Connection.Write(fmt.Appendf(nil, "$%d\r\n", len(rdbBytes)))
	session.Connection.Write(rdbBytes)

	session.IsReplica = true

	server.mu.Lock()
	server.replicas = append(server.replicas, session)
	server.mu.Unlock()

	return ""
}

func (server *Server) HandleReplconf(session *Session, arr []string) string {
	if len(arr) < 1 {
		return "-ERR Invalid arguments"
	}

	arg := strings.ToUpper(arr[1])

	if arg == "GETACK" {
		replOffset := strconv.Itoa(session.ReplOffset)
		return fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$%d\r\n%s\r\n", len(replOffset), replOffset)
	}

	if arg == "ACK" {
		offset, err := strconv.Atoi(arr[2])
		if err != nil {
			return "-ERR Failed to convert replica offset\r\n"
		}

		session.ReplOffset = offset
		return ""
	}

	return "+OK\r\n"
}

func (server *Server) HandleWait(arr []string) string {
	if len(arr) < 3 {
		return "-ERR Missing arguments. Please try: WAIT <num> <timeout>\r\n"
	}

	numReplicas, err := strconv.Atoi(arr[1])
	if err != nil {
		return "-ERR Invalid argument for num of replicas, it should be an integer\r\n"
	}
	timeout, err := strconv.Atoi(arr[2])
	if err != nil {
		return "-ERR Invalid timeout. It should be an integer\r\n"
	}

	if server.MasterReplOffset == 0 {
		server.mu.RLock()
		count := len(server.replicas)
		server.mu.RUnlock()
		return fmt.Sprintf(":%d\r\n", count)
	}

	server.mu.RLock()
	for _, replica := range server.replicas {
		replica.Connection.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$6\r\nGETACK\r\n$1\r\n*\r\n"))
	}
	server.mu.RUnlock()

	start := time.Now()
	for {
		done := 0

		server.mu.RLock()
		for _, replica := range server.replicas {
			if replica.ReplOffset >= server.MasterReplOffset {
				done++
			}
		}
		server.mu.RUnlock()

		if done >= numReplicas {
			return fmt.Sprintf(":%d\r\n", done)
		}

		if timeout > 0 && time.Since(start) > time.Duration(timeout)*time.Millisecond {
			return fmt.Sprintf(":%d\r\n", done)
		}

		time.Sleep(10 * time.Millisecond)
	}
}
