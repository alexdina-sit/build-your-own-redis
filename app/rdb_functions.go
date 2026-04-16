package main

import (
	"fmt"
	"strings"
)

func (server *Server) configGet(args []string) string {
	if len(args) < 3 {
		return "-ERR Missing arguments. Please try: CONFIG GET <parameter>\r\n"
	}

	var sb strings.Builder
	parameter := strings.ToLower(args[2])

	switch parameter {
	case "dir":
		{
			addRespArrayHeader(&sb, 2)
			addRespString(&sb, "dir")
			addRespString(&sb, server.rdb.Dir)
		}
	case "dbfilename":
		{
			addRespArrayHeader(&sb, 2)
			addRespString(&sb, "dbfilename")
			addRespString(&sb, server.rdb.DbFileName)

		}
	}
	return sb.String()
}

func (server *Server) handleConfig(args []string) string {
	if len(args) < 2 {
		return "-ERR Missing arguments.\r\n"
	}

	action := args[1]
	if action == "GET" {
		return server.configGet(args)
	}
	return ""
}

func readRdb(data []byte) map[string]string {
	result := make(map[string]string)
	i := 9

	for i < len(data) {
		if data[i] == 0xFF {
			break
		}

		if data[i] == 0xFB {
			i += 2
			continue
		}

		if data[i] != 0x00 {
			i++
			continue
		}

		keyLen := int(data[i])
		fmt.Println(keyLen)
	}

	return result
}
