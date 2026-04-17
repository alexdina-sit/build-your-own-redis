package main

func (server *Server) HandleWatch(session *Session, args []string) string {
	if len(args) < 2 {
		return "-ERR Missing arguments. Please try: WATCH <key>"
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	for _, key := range args[1:] {
		session.WatchedKeys[key] = server.ServerKeys[key]
	}
	return "+OK\r\n"
}

func (server *Server) HandleUnwatch(session *Session) string {
	session.WatchedKeys = make(map[string]int)
	return "+OK\r\n"
}
