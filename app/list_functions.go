package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (server *Server) HandlePush(arr []string, direction Direction) string {
	if len(arr) < 3 {
		return "-ERR Missing arguments, your input should be: RPUSH <lname> <value_1>.."
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	for _, val := range arr[2:] {
		if direction == Right {
			server.listsMap[arr[1]] = append(server.listsMap[arr[1]], val)
			continue
		}

		server.listsMap[arr[1]] = append([]string{val}, server.listsMap[arr[1]]...)

	}
	return fmt.Sprintf(":%d\r\n", len(server.listsMap[arr[1]]))
}

func (server *Server) HandleLrange(arr []string) string {
	if len(arr) < 4 {
		return "-ERR Missing arguments, your input should be: LRANGE <lname> <start> <stop>"
	}

	start, err := strconv.Atoi(arr[2])
	if err != nil {
		return "-ERR Failed to convert the start index"
	}

	stop, err := strconv.Atoi(arr[3])
	if err != nil {
		return "-ERR Failed to convert the stop index"
	}

	server.mu.RLock()
	defer server.mu.RUnlock()

	value, prs := server.listsMap[arr[1]]
	if !prs {
		return "*0\r\n"
	}

	listLen := len(value)
	if start < 0 {
		if -start > listLen {
			start = 0
		} else {
			start = listLen + start
		}
	}

	if stop < 0 {
		if -stop > listLen {
			stop = 0
		} else {
			stop = listLen + stop
		}
	}

	if start >= listLen || start > stop {
		return "*0\r\n"
	}

	if stop >= listLen {
		stop = listLen - 1
	}

	var sb strings.Builder
	addRespArrayHeader(&sb, (stop - start + 1))

	for i := start; i <= stop; i++ {
		addRespString(&sb, value[i])
	}
	return sb.String()
}

func (server *Server) HandleLlen(arr []string) string {
	if len(arr) < 2 {
		return "-Missing arguments. Your input should be: LLEN <list_key>\r\n"
	}

	server.mu.RLock()
	defer server.mu.RUnlock()
	return fmt.Sprintf(":%d\r\n", len(server.listsMap[arr[1]]))
}

func (server *Server) HandleLpop(arr []string) string {
	if len(arr) < 2 {
		return "-Missing arguments. Your input should be: LPOP <list_key>\r\n"
	}

	server.mu.Lock()
	defer server.mu.Unlock()
	sl, prs := server.listsMap[arr[1]]
	if !prs || len(sl) == 0 {
		return "$-1\r\n"
	}

	if len(arr) == 2 {
		value := sl[0]
		sl = sl[1:]

		if len(sl) == 0 {
			delete(server.listsMap, arr[1])
		} else {
			server.listsMap[arr[1]] = sl
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	}
	stop, err := strconv.Atoi(arr[2])
	if err != nil {
		return "-ERR Error while converting your argument"
	}

	if stop > len(sl) {
		stop = len(sl)
	}

	var sb strings.Builder
	addRespArrayHeader(&sb, stop)

	for i := 0; i < stop; i++ {
		addRespString(&sb, sl[i])
	}

	sl = sl[stop:]
	if len(sl) == 0 {
		delete(server.listsMap, arr[1])
	} else {
		server.listsMap[arr[1]] = sl
	}
	return sb.String()

}

func (server *Server) HandleBlpop(arr []string) string {
	if len(arr) < 3 {
		return "-Missing arguments. Your input should be: BLPOP <list_key> <timeout>\r\n"
	}

	timeout, err := strconv.ParseFloat(arr[2], 64)
	if err != nil {
		return "-ERR Error while converting the timeout"
	}

	now := time.Now()
	for timeout == 0 || time.Since(now).Seconds() < timeout {
		server.mu.Lock()
		sl := server.listsMap[arr[1]]
		if len(sl) > 0 {
			value := sl[0]
			sl = sl[1:]
			if len(sl) == 0 {
				delete(server.listsMap, arr[1])
			} else {
				server.listsMap[arr[1]] = sl
			}

			server.mu.Unlock()
			return fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(arr[1]), arr[1], len(value), value)
		}
		server.mu.Unlock()
		time.Sleep(50 * time.Millisecond)
	}

	return "*-1\r\n"
}
