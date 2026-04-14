package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func (server *Server) handleAcl(session *ClientSession, arr []string) string {
	if strings.ToUpper(arr[1]) == "WHOAMI" {
		name := session.UserName
		return fmt.Sprintf("$%d\r\n%s\r\n", len(name), name)
	}

	if strings.ToUpper(arr[1]) == "GETUSER" {
		if len(arr) < 3 {
			return "-You must provide the user name. Syntax: ACL GETUSER <username>.\r\n"
		}

		server.mu.RLock()
		defer server.mu.RUnlock()

		user, prs := server.usersMap[arr[2]]
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

	if strings.ToUpper(arr[1]) == "SETUSER" {
		if len(arr) < 3 {
			return "-You must provide the user name. Syntax: ACL SETUSER <username>.\r\n"
		}

		var password string
		h := sha256.New()
		if len(arr) == 4 && arr[3][0] == '>' {
			h.Write([]byte(arr[3][1:]))
			password = fmt.Sprintf("%x", h.Sum(nil))
		}

		server.mu.Lock()
		defer server.mu.Unlock()

		server.usersMap[arr[2]] = &User{}
		if len(password) > 0 {
			server.usersMap[arr[2]].Passwords = append(server.usersMap[arr[2]].Passwords, password)
			server.usersMap[arr[2]].Flags = make([]string, 0)
		} else {
			server.usersMap[arr[2]].Passwords = make([]string, 0)
			server.usersMap[arr[2]].Flags = append(server.usersMap[arr[2]].Flags, "nopass")
		}
		return "+OK\r\n"
	}

	return ""
}

func (server *Server) handleAuth(session *ClientSession, arr []string) string {
	if len(arr) < 3 {
		return "-Missing parameters. Try AUTH <username> <password>.\r\n"
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	user, prs := server.usersMap[arr[1]]
	if !prs {
		return "-User doesn't exist"
	}

	h := sha256.New()
	h.Write([]byte(arr[2]))
	encoded := fmt.Sprintf("%x", h.Sum(nil))
	if encoded == user.Passwords[0] {
		session.IsAuthenticated = true
		session.UserName = arr[1]
		return "+OK\r\n"
	}

	return "-WRONGPASS invalid username-password pair or user is disabled.\r\n"
}
