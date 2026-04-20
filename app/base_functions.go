package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Item struct {
	Value      string
	ExpireType string
	ExpireTime int
	CreateDate time.Time
}

func (server *Server) HandleSet(args []string) string {
	var expireType string
	var expireTime int

	if len(args) < 3 {
		return "-ERR Missing arguments. Your input should be: SET <key> <value>\r\n"
	}

	if len(args) > 4 {
		expireType = strings.ToUpper(args[3])
		intValue, err := strconv.Atoi(args[4])
		if err != nil {
			return "-ERR Error while converting your expire value\r\n"
		}

		expireTime = intValue
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	server.itemsMap[args[1]] = &Item{
		Value:      args[2],
		ExpireType: expireType,
		ExpireTime: expireTime,
		CreateDate: time.Now(),
	}

	server.ServerKeys[args[1]]++
	return "+OK\r\n"
}

func (server *Server) HandleGet(arr []string) string {
	server.mu.Lock()
	defer server.mu.Unlock()

	item, prs := server.itemsMap[arr[1]]
	if !prs {
		return "$-1\r\n"
	}

	expireType := strings.ToUpper(item.ExpireType)
	timeSpent := time.Since(item.CreateDate)

	if (expireType == "EX" && timeSpent.Seconds() > float64(item.ExpireTime)) ||
		(expireType == "PX" && timeSpent.Milliseconds() > int64(item.ExpireTime)) {

		delete(server.itemsMap, arr[1])
		return "$-1\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(item.Value), item.Value)
}
