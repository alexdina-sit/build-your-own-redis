package main

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

func addRespString(sb *strings.Builder, str string) {
	sb.WriteString("$")
	sb.WriteString(strconv.Itoa(len(str)))
	sb.WriteString(CRLF)
	sb.WriteString(str)
	sb.WriteString(CRLF)
}

func addRespArrayHeader(sb *strings.Builder, length int) {
	sb.WriteString("*")
	sb.WriteString(strconv.Itoa(length))
	sb.WriteString(CRLF)
}

func processRespArray(respArray string) []string {
	splitedString := strings.Split(respArray, "\r\n")
	var returnArray []string

	for _, elem := range splitedString[2:] {
		if elem == "" || (len(elem) > 1 && elem[0] == 36) {
			continue
		}
		returnArray = append(returnArray, elem)
	}
	return returnArray
}

func readRespCommand(reader *bufio.Reader) (string, error) {
	var sb strings.Builder

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if len(line) == 0 {
		return "", io.EOF
	}

	sb.WriteString(line)

	switch line[0] {
	case '*':
		count, err := strconv.Atoi(strings.TrimSpace(line[1:]))
		if err != nil {
			return "", err
		}

		for range count {
			elemHeader, err := reader.ReadString('\n')
			if err != nil {
				return sb.String(), err
			}
			sb.WriteString(elemHeader)

			if elemHeader[0] == '$' {
				length, err := strconv.Atoi(strings.TrimSpace(elemHeader[1:]))
				if err != nil {
					return sb.String(), err
				}

				if length != -1 {
					buf := make([]byte, length+2)
					_, err = io.ReadFull(reader, buf)
					if err != nil {
						return sb.String(), err
					}
					sb.Write(buf)
				}
			}

		}

	case '$':
		length, err := strconv.Atoi(strings.TrimSpace(line[1:]))
		if err != nil {
			return "", err
		}

		if length != -1 {
			buf := make([]byte, length+2)
			_, err = io.ReadFull(reader, buf)
			if err != nil {
				return sb.String(), err
			}
			sb.Write(buf)
		}
	}

	return sb.String(), nil
}

func propagate(sessions []*Session, data string) error {
	for _, session := range sessions {
		_, err := session.Connection.Write([]byte(data))
		if err != nil {
			return err
		}
	}

	return nil
}
