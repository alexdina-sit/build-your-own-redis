package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func handleIncr(arr []string) string {
	if len(arr) < 2 {
		return "-Missing arguments. Try INCR <key>"
	}

	mu.Lock()
	defer mu.Unlock()
	item, prs := mmap[arr[1]]
	if !prs {
		mmap[arr[1]] = &Item{
			Value: "1",
		}
		return ":1\r\n"
	}

	value, err := strconv.Atoi(item.Value)
	if err != nil {
		return "-ERR value is not an integer or out of range\r\n"
	}

	value += 1
	mmap[arr[1]].Value = fmt.Sprintf("%d", value)
	return fmt.Sprintf(":%d\r\n", value)
}

func handleMulti(session *ClientSession) {
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
		if strings.Contains(text, "DISCARD") {
			_, err = conn.Write([]byte("+OK\r\n"))
			session.commandsQueue = make([]string, 0)
			return
		}

		if strings.Contains(text, "EXEC") {
			if len(session.commandsQueue) == 0 {
				_, err = conn.Write([]byte("*0\r\n"))
				return
			}

			returnStr := fmt.Sprintf("*%d\r\n", len(session.commandsQueue))
			for _, command := range session.commandsQueue {
				returnStr += respParser(session, command)
			}

			_, err = conn.Write([]byte(returnStr))
			return
		}

		session.commandsQueue = append(session.commandsQueue, text)
		_, err = conn.Write([]byte("+QUEUED\r\n"))
	}
}
