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

var mmap = map[string]*Item{}
var lmap = map[string][]string{}
var umap = map[string]*User{}
var mu sync.RWMutex

func main() {
	fmt.Println("Logs from your program will appear here!")
	umap["default"] = &User{Flags: []string{"nopass"}}

	listener, err := net.Listen("tcp", "0.0.0.0:6379")
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

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("Read error:", err)
			break
		}

		text := respParser(string(buf[:n]))

		_, err = conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Write error:", err)
		}
	}
}

func respParser(input string) string {
	if input[0] == 42 {
		arr := readArray(input)

		switch strings.ToUpper(arr[0]) {
		case "ECHO":
			return fmt.Sprintf("$%d\r\n%s\r\n", len(arr[1]), arr[1])

		case "PING":
			return "+PONG\r\n"

		case "GET":
			return handleGet(arr)

		case "SET":
			return handleSet(arr)

		case "ACL":
			return handleAcl(arr)

		case "AUTH":
			return handleAuth(arr)

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
		}
	}

	return ""
}
