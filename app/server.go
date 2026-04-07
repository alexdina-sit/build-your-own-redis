package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Item struct {
	Value      string
	ExpireType string
	ExpireTime int
	CreateDate time.Time
}

type User struct {
	Passwords     []string
	Flags         []string
	Authenticated bool
}

type ClientSession struct {
	Connection      net.Conn
	IsAuthenticated bool
	UserName        string
}

var mmap = make(map[string]*Item)
var lmap = make(map[string][]string)
var umap = make(map[string]*User)
var mu sync.RWMutex

func main() {
	fmt.Println("Logs from your program will appear here!")
	umap["default"] = &User{Flags: []string{"nopass"}}

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
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

	reader := bufio.NewReader(conn)
	buf := make([]byte, 1024)

	session := &ClientSession{
		Connection:      conn,
		IsAuthenticated: false,
		UserName:        "default",
	}

	mu.RLock()
	defaultUser, exists := umap["default"]
	if exists {
		for _, flag := range defaultUser.Flags {
			if flag == "nopass" {
				session.IsAuthenticated = true
			}
		}
	}

	mu.RUnlock()

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("Read error:", err)
			break
		}

		text := respParser(session, string(buf[:n]))

		_, err = conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Write error:", err)
		}
	}
}

func respParser(session *ClientSession, input string) string {
	if input[0] == 42 {
		arr := readArray(input)

		cmd := strings.ToUpper(arr[0])
		if !session.IsAuthenticated && arr[0] != "AUTH" {
			return "-NOAUTH Authentication required\r\n"
		}

		switch cmd {
		case "ECHO":
			return fmt.Sprintf("$%d\r\n%s\r\n", len(arr[1]), arr[1])

		case "PING":
			return "+PONG\r\n"

		case "GET":
			return handleGet(arr)

		case "SET":
			return handleSet(arr)

		case "ACL":
			return handleAcl(session, arr)

		case "AUTH":
			return handleAuth(session, arr)

		case "RPUSH":
			return handlePush(arr, "right")

		case "LPUSH":
			return handlePush(arr, "left")

		case "LRANGE":
			return handleLrange(arr)

		case "LLEN":
			return handleLlen(arr)

		case "LPOP":
			return handleLpop(arr)

		case "BLPOP":
			return handleBlpop(arr)

		case "TYPE":
			return handleType(arr)

		case "XADD":
			return handleXadd(arr)
		}
	}

	return ""
}
