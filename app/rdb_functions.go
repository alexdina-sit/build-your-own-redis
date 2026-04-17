package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (server *Server) LoadRdb() {
	if server.Config.DbFileName == "" {
		return
	}

	path := filepath.Join(server.Config.Dir, server.Config.DbFileName)

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to open the RDB file: " + err.Error())
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	header := make([]byte, 9)
	if _, err := io.ReadFull(reader, header); err != nil {
		fmt.Println("There was an error while trying to read the RDB header: " + err.Error())
		return
	}

	for {
		opcode, err := reader.ReadByte()
		if err != nil || opcode == 0xFF {
			break
		}

		switch opcode {
		case 0xFB:
			{
				keysNum, _ := reader.ReadByte()
				reader.ReadByte()

				for range keysNum {
					var expireTimeMs uint64
					hasExpireTime := false
					b, _ := reader.ReadByte()

					switch b {
					case 0xFC:
						{
							hasExpireTime = true
							buf := make([]byte, 8)
							io.ReadFull(reader, buf)
							expireTimeMs = uint64(binary.LittleEndian.Uint64(buf))
							b, _ = reader.ReadByte()
						}
					case 0xFD:
						{
							hasExpireTime = true
							buf := make([]byte, 4)
							io.ReadFull(reader, buf)
							expireTimeMs = uint64(binary.LittleEndian.Uint64(buf)) * 1000
							b, _ = reader.ReadByte()
						}
					}

					if b == 0 {
						server.readKeyValueItem(reader, hasExpireTime, expireTimeMs)
					}
				}
				break
			}
		}
	}
}

func (server *Server) readKeyValueItem(reader *bufio.Reader, hasExpireTime bool, expireTimeMs uint64) string {
	keyLen, errK := reader.ReadByte()
	key := make([]byte, keyLen)
	io.ReadFull(reader, key)

	valueLen, errV := reader.ReadByte()
	value := make([]byte, valueLen)
	io.ReadFull(reader, value)

	if errK != nil || errV != nil {
		return "$-1\r\n"
	}

	if hasExpireTime {
		timeNowMs := time.Now().UnixMilli()

		if timeNowMs >= int64(expireTimeMs) {
			return ""
		}

		difference := int64(expireTimeMs) - timeNowMs
		server.handleSet([]string{"SET", string(key), string(value), "PX", fmt.Sprint(difference)})
	} else {
		server.handleSet([]string{"SET", string(key), string(value)})
	}
	return ""
}

func (server *Server) configGet(args []string) string {
	if len(args) < 3 {
		return "-ERR Missing arguments. Please try: CONFIG GET <parameter>\r\n"
	}

	parameter := strings.ToLower(args[2])

	var sb strings.Builder
	addRespArrayHeader(&sb, 2)
	addRespString(&sb, parameter)

	switch parameter {
	case "dir":
		addRespString(&sb, server.Config.Dir)

	case "dbfilename":
		addRespString(&sb, server.Config.DbFileName)

	case "appendonly":
		addRespString(&sb, server.Config.AppendOnly)

	case "appenddirname":
		addRespString(&sb, server.Config.AppendDirname)

	case "appendfilename":
		addRespString(&sb, server.Config.AppendFilename)

	case "appendfsync":
		addRespString(&sb, server.Config.AppendFSync)

	default:
		return "$-1\r\n"
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
