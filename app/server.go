package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	server := GetServerInstance()
	server.usersMap["default"] = &User{Flags: []string{"nopass"}}

	portFlag := flag.String("port", "6379", "The port to listen on")
	replicaFlag := flag.String("replicaof", "", "Master and slave ports")
	flag.Parse()

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
	reader := bufio.NewReader(conn)
	buf := make([]byte, 1024)

	session := &ClientSession{
		Connection:      conn,
		IsAuthenticated: false,
		UserName:        "default",
		reader:          *reader,
		buf:             buf,
	}

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

		text := respParser(session, string(cmd))
		_, err = conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Write error:", err)
		}
	}
}
