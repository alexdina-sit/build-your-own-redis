package main

import (
	"fmt"
	"strings"
)

func (server *Server) handleSubscribe(session *Session, args []string) string {
	if len(args) < 2 {
		return "-ERR Missing arguments. Please try: SUBSCRIBE <channel_name>"
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	var sb strings.Builder
	addRespArrayHeader(&sb, 3)
	addRespString(&sb, "subscribe")

	channelName := args[1]
	addRespString(&sb, channelName)
	_, exists := session.SubscribedChannels[channelName]
	if !exists {
		session.SubscribedChannels[channelName] = true
		server.channels[channelName] = append(server.channels[channelName], session)
	}

	count := fmt.Sprintf(":%d\r\n", len(session.SubscribedChannels))
	sb.WriteString(count)
	return sb.String()
}
