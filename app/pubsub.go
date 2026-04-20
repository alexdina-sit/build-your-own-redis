package main

import (
	"fmt"
	"strings"
)

func (server *Server) HandleSubscribe(session *Session, args []string) string {
	if len(args) < 2 {
		return "-ERR Missing arguments. Please try: SUBSCRIBE <channel_name>\r\n"
	}

	var sb strings.Builder
	channelName := args[1]

	addRespArrayHeader(&sb, 3)
	addRespString(&sb, "subscribe")
	addRespString(&sb, channelName)

	server.mu.Lock()
	defer server.mu.Unlock()

	_, exists := session.SubscribedChannels[channelName]
	if !exists {
		session.SubscribedChannels[channelName] = true
		server.channels[channelName] = append(server.channels[channelName], session)
	}

	fmt.Fprintf(&sb, ":%d\r\n", len(session.SubscribedChannels))
	return sb.String()
}

func (server *Server) HandleUnsubscribe(session *Session, args []string) string {
	if len(args) < 2 {
		return "-ERR Missing arguments. Please try: SUBSCRIBE <channel_name>\r\n"
	}

	var sb strings.Builder
	channelName := args[1]

	addRespArrayHeader(&sb, 3)
	addRespString(&sb, "unsubscribe")
	addRespString(&sb, channelName)

	server.mu.Lock()
	defer server.mu.Unlock()

	_, exists := session.SubscribedChannels[channelName]
	if !exists {
		fmt.Fprintf(&sb, ":%d\r\n", len(session.SubscribedChannels))
		return sb.String()
	}

	delete(session.SubscribedChannels, channelName)
	channel := server.channels[channelName]
	sessionIndex := 0
	for index, ses := range channel {
		if ses == session {
			sessionIndex = index
		}
	}

	server.channels[channelName] = append(channel[:sessionIndex], channel[sessionIndex+1:]...)
	fmt.Fprintf(&sb, ":%d\r\n", len(session.SubscribedChannels))
	return sb.String()
}

func (server *Server) HandlePublish(session *Session, args []string) string {
	if len(args) < 3 {
		return "-ERR Missing arguments. Please try: PUBLISH <channel_name> <message>\r\n"
	}

	var sb strings.Builder
	channelName, message := args[1], args[2]

	server.mu.Lock()
	defer server.mu.Unlock()

	channel, exists := server.channels[channelName]
	if !exists {
		return ":0\r\n"
	}

	addRespArrayHeader(&sb, 3)
	addRespString(&sb, "message")
	addRespString(&sb, channelName)
	addRespString(&sb, message)

	err := propagate(channel, sb.String())
	if err != nil {
		return "-ERR There was an error while publishing your message\r\n"
	}

	return fmt.Sprintf(":%d\r\n", len(channel))
}
