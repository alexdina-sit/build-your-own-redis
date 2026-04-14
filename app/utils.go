package main

import (
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
