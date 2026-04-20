package main

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
)

type User struct {
	Passwords     []string
	Flags         []string
	Authenticated bool
}

func (server *Server) handleAcl(session *Session, args []string) string {
	if strings.ToUpper(args[1]) == "WHOAMI" {
		name := session.UserName
		return fmt.Sprintf("$%d\r\n%s\r\n", len(name), name)
	}

	if strings.ToUpper(args[1]) == "GETUSER" {
		return handleGetUser(server, args)
	}

	if strings.ToUpper(args[1]) == "SETUSER" {
		return handleSetUser(server, args)
	}

	return ""
}

func handleGetUser(server *Server, args []string) string {
	if len(args) < 3 {
		return "-You must provide the user name. Syntax: ACL GETUSER <username>.\r\n"
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	user, prs := server.usersMap[args[2]]
	if !prs {
		return "*0\r\n"
	}

	var sb strings.Builder
	sb.WriteString("*4\r\n$5\r\nflags\r\n")
	addRespArrayHeader(&sb, len(user.Flags))

	for _, flag := range user.Flags {
		addRespString(&sb, flag)
	}

	sb.WriteString("$9\r\npasswords\r\n")
	addRespArrayHeader(&sb, len(user.Passwords))
	for _, password := range user.Passwords {
		addRespString(&sb, password)
	}

	return sb.String()
}

func handleSetUser(server *Server, args []string) string {
	if len(args) < 3 {
		return "-You must provide the user name. Syntax: ACL SETUSER <username>.\r\n"
	}

	var password string
	h := sha256.New()
	if len(args) == 4 && args[3][0] == '>' {
		h.Write([]byte(args[3][1:]))
		password = fmt.Sprintf("%x", h.Sum(nil))
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	server.usersMap[args[2]] = &User{}
	if len(password) > 0 {
		server.usersMap[args[2]].Passwords = append(server.usersMap[args[2]].Passwords, password)
		server.usersMap[args[2]].Flags = make([]string, 0)
	} else {
		server.usersMap[args[2]].Passwords = make([]string, 0)
		server.usersMap[args[2]].Flags = append(server.usersMap[args[2]].Flags, "nopass")
	}
	return "+OK\r\n"
}

func (server *Server) handleAuth(session *Session, arr []string) string {
	if len(arr) < 3 {
		return "-Missing arguments. Please try ry AUTH <username> <password>.\r\n"
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	user, prs := server.usersMap[arr[1]]
	if !prs {
		return "-ERR User doesn't exist\r\n"
	}

	h := sha256.New()
	h.Write([]byte(arr[2]))
	encoded := fmt.Sprintf("%x", h.Sum(nil))

	if slices.Contains(user.Passwords, encoded) {
		session.IsAuthenticated = true
		session.UserName = arr[1]
		return "+OK\r\n"
	}

	return "-WRONGPASS invalid username-password pair or user is disabled.\r\n"
}
