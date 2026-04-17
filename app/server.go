package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strings"
)

type Direction string

const CRLF = "\r\n"
const (
	Left  Direction = "left"
	Right Direction = "right"
)

var allowedCommandsSubscribed = []string{"SUBSCRIBE", "UNSUBSCRIBE", "PSUBSCRIBE", "PUNSUBSCRIBE", "PING", "QUIT"}

func main() {
	server := GetServerInstance()
	server.usersMap["default"] = &User{Flags: []string{"nopass"}}

	portFlag := flag.String("port", "6379", "The port to listen on")
	replicaFlag := flag.String("replicaof", "", "Master and slave ports")

	dirFlag := flag.String("dir", "", "File dir")
	dbfilenameFlag := flag.String("dbfilename", "", "RDB dbfilename")
	appendonlyFlag := flag.String("appendonly", "", "Appendonly flag")
	appenddirnameFlag := flag.String("appenddirname", "", "Appenddirname flag")
	appendfilenameFlag := flag.String("appendfilename", "", "Appendfilename flag")
	appendfsyncFlag := flag.String("appendfsync", "", "Appendfsync flag")

	fmt.Println(*appendonlyFlag, *appenddirnameFlag, *appendfilenameFlag, *appendfsyncFlag)

	flag.Parse()

	if *dirFlag != "" && *dbfilenameFlag != "" {
		server.rdb.Dir = *dirFlag
		server.rdb.DbFileName = *dbfilenameFlag
		server.loadRdb()
	}

	if *replicaFlag != "" {
		server.Role = "slave"

		masterHost, masterPort, err := getMasterAddress(*replicaFlag)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		go handshake(masterHost, masterPort)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", *portFlag))
	if err != nil {
		fmt.Printf("Failed to bind to port: %s\n", *portFlag)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	server := GetServerInstance()
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

		text := handleCommand(session, string(cmd))
		_, err = conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Write error:", err)
		}
	}
}

func handleCommand(session *Session, input string) string {
	server := GetServerInstance()

	if input[0] == 42 {
		args := processRespArray(input)

		command := strings.ToUpper(args[0])
		if !session.IsAuthenticated && args[0] != "AUTH" {
			return "-NOAUTH Authentication required\r\n"
		}

		if len(session.SubscribedChannels) > 0 && !slices.Contains(allowedCommandsSubscribed, command) {
			return fmt.Sprintf("-ERR Can't execute '%s' while subscribed to a channel\r\n", command)
		}

		fnc, exists := server.commandsMap[command]
		if exists {
			if command != "PSYNC" && command != "MULTI" {
				return fnc(command, session, args)
			}
			fnc(command, session, args)
		}

		switch command {
		case "SET":
			{
				response := server.handleSet(args)

				if server.Role == "master" {
					server.mu.Lock()
					defer server.mu.Unlock()

					server.MasterReplOffset += len([]byte(input))
					propagate(server.replicas, input)
				}

				return response
			}

		case "KEYS":
			{
				var sb strings.Builder
				keys := len(server.itemsMap)
				addRespArrayHeader(&sb, keys)
				for key := range server.itemsMap {
					addRespString(&sb, key)
				}

				return sb.String()
			}
		}
	}

	return ""
}
