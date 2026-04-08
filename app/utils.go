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
