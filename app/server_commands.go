package main

func (server *Server) initCommands() {
	server.commandsMap = map[string]func(input string, session *Session, args []string) string{
		"EXEC":    func(_ string, _ *Session, _ []string) string { return server.execWithoutMulti() },
		"DISCARD": func(_ string, _ *Session, _ []string) string { return server.discardWithoutMulti() },

		"ECHO":      func(_ string, _ *Session, args []string) string { return server.handleEcho(args) },
		"GET":       func(_ string, _ *Session, args []string) string { return server.handleGet(args) },
		"LRANGE":    func(_ string, _ *Session, args []string) string { return server.handleLrange(args) },
		"LLEN":      func(_ string, _ *Session, args []string) string { return server.handleLlen(args) },
		"TYPE":      func(_ string, _ *Session, args []string) string { return server.handleType(args) },
		"INFO":      func(_ string, _ *Session, args []string) string { return server.handleInfo(args) },
		"XREAD":     func(_ string, _ *Session, args []string) string { return server.handleXread(args) },
		"XRANGE":    func(_ string, _ *Session, args []string) string { return server.handleXrange(args) },
		"ZRANGE":    func(_ string, _ *Session, args []string) string { return server.handleZrange(args) },
		"ZRANK":     func(_ string, _ *Session, args []string) string { return server.handleZrank(args) },
		"ZCARD":     func(_ string, _ *Session, args []string) string { return server.handleZcard(args) },
		"GEOPOS":    func(_ string, _ *Session, args []string) string { return server.handleGeopos(args) },
		"GEODIST":   func(_ string, _ *Session, args []string) string { return server.handleGeodist(args) },
		"GEOSEARCH": func(_ string, _ *Session, args []string) string { return server.handleGeosearch(args) },
		"ZSCORE":    func(_ string, _ *Session, args []string) string { return server.handleZscore(args) },
		"WAIT":      func(_ string, _ *Session, args []string) string { return server.handleWait(args) },

		"PING":    func(_ string, session *Session, _ []string) string { return server.handlePing(session) },
		"UNWATCH": func(_ string, session *Session, _ []string) string { return server.HandleUnwatch(session) },

		"ZREM":   func(input string, _ *Session, args []string) string { return server.handleZrem(args) },
		"RPUSH":  func(input string, _ *Session, args []string) string { return server.handlePush(args, Right) },
		"LPUSH":  func(input string, _ *Session, args []string) string { return server.handlePush(args, Left) },
		"CONFIG": func(input string, _ *Session, args []string) string { return server.handleConfig(args) },
		"GEOADD": func(input string, _ *Session, args []string) string { return server.handleGeoadd(args) },
		"ZADD":   func(input string, _ *Session, args []string) string { return server.handleZadd(args) },
		"INCR":   func(input string, _ *Session, args []string) string { return server.handleIncr(args) },
		"XADD":   func(input string, _ *Session, args []string) string { return server.handleXadd(args) },
		"LPOP":   func(input string, _ *Session, args []string) string { return server.handleLpop(args) },
		"BLPOP":  func(input string, _ *Session, args []string) string { return server.handleBlpop(args) },

		"ACL":         func(_ string, session *Session, args []string) string { return server.handleAcl(session, args) },
		"AUTH":        func(_ string, session *Session, args []string) string { return server.handleAuth(session, args) },
		"REPLCONF":    func(_ string, session *Session, args []string) string { return server.handleReplconf(session, args) },
		"SUBSCRIBE":   func(_ string, session *Session, args []string) string { return server.HandleSubscribe(session, args) },
		"PUBLISH":     func(_ string, session *Session, args []string) string { return server.HandlePublish(session, args) },
		"UNSUBSCRIBE": func(_ string, session *Session, args []string) string { return server.HandleUnsubscribe(session, args) },

		"MULTI": func(_ string, session *Session, _ []string) string {
			server.handleMulti(session)
			return ""
		},

		"PSYNC": func(_ string, session *Session, _ []string) string {
			server.handlePsync(session)
			return ""
		},

		"WATCH": func(input string, session *Session, args []string) string { return server.HandleWatch(session, args) },
	}
}
