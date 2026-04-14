package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (server *Server) handleSet(arr []string) string {
	var expireType string
	var expireTime int

	if len(arr) < 3 {
		return "-ERR Missing arguments. Your input should be: SET <key> <value>\r\n"
	}

	if len(arr) > 4 {
		expireType = strings.ToUpper(arr[3])
		intValue, err := strconv.Atoi(arr[4])
		if err != nil {
			return "-ERR Error while converting your expire value\r\n"
		}

		expireTime = intValue
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	server.itemsMap[arr[1]] = &Item{
		Value:      arr[2],
		ExpireType: expireType,
		ExpireTime: expireTime,
		CreateDate: time.Now(),
	}

	return "+OK\r\n"
}

func (server *Server) handleGet(arr []string) string {
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

func respParser(session *ClientSession, input string) string {
	server := GetServerInstance()

	if input[0] == 42 {
		arr := processRespArray(input)

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
			return server.handleGet(arr)

		case "SET":
			return server.handleSet(arr)

		case "ACL":
			return server.handleAcl(session, arr)

		case "AUTH":
			return server.handleAuth(session, arr)

		case "RPUSH":
			return server.handlePush(arr, Right)

		case "LPUSH":
			return server.handlePush(arr, Left)

		case "LRANGE":
			return server.handleLrange(arr)

		case "LLEN":
			return server.handleLlen(arr)

		case "LPOP":
			return server.handleLpop(arr)

		case "BLPOP":
			return server.handleBlpop(arr)

		case "TYPE":
			return server.handleType(arr)

		case "XADD":
			return server.handleXadd(arr)

		case "INCR":
			return server.handleIncr(arr)

		case "MULTI":
			server.handleMulti(session)

		case "XRANGE":
			return server.handleXrange(arr)

		case "EXEC":
			return "-ERR EXEC without MULTI\r\n"

		case "DISCARD":
			return "-ERR DISCARD without MULTI\r\n"

		case "ZADD":
			return server.handleZadd(arr)

		case "XREAD":
			return server.handleXread(arr)

		case "ZRANK":
			return server.handleZrank(arr)

		case "ZCARD":
			return server.handleZcard(arr)

		case "ZRANGE":
			return server.handleZrange(arr)

		case "ZSCORE":
			return server.handleZscore(arr)

		case "ZREM":
			return server.handleZrem(arr)

		case "GEOADD":
			return server.handleGeoadd(arr)

		case "GEOPOS":
			return server.handleGeopos(arr)

		case "GEODIST":
			return server.handleGeodist(arr)

		case "GEOSEARCH":
			return server.handleGeosearch(arr)

		case "INFO":
			return server.handleInfo(arr)

		case "REPLCONF":
			return server.handleReplconf()

		case "PSYNC":
			server.handlePsync(session)
		}
	}

	return ""
}
