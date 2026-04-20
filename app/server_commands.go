package main

var replicableCommands = []string{"SET", "RPUSH", "LPUSH", "ZREM", "GEOADD", "ZADD", "INCR", "XADD", "LPOP"}

func (server *Server) initCommands() {
	server.commandsMap = map[string]func(input string, session *Session, args []string) string{
		"EXEC":    func(_ string, _ *Session, _ []string) string { return server.ExecWithoutMulti() },
		"DISCARD": func(_ string, _ *Session, _ []string) string { return server.DiscardWithoutMulti() },
		"KEYS":    func(_ string, session *Session, _ []string) string { return server.HandleKeys() },

		"ECHO":      func(_ string, _ *Session, args []string) string { return server.HandleEcho(args) },
		"GET":       func(_ string, _ *Session, args []string) string { return server.HandleGet(args) },
		"LRANGE":    func(_ string, _ *Session, args []string) string { return server.HandleLrange(args) },
		"LLEN":      func(_ string, _ *Session, args []string) string { return server.HandleLlen(args) },
		"TYPE":      func(_ string, _ *Session, args []string) string { return server.HandleType(args) },
		"INFO":      func(_ string, _ *Session, args []string) string { return server.HandleInfo(args) },
		"XREAD":     func(_ string, _ *Session, args []string) string { return server.handleXread(args) },
		"XRANGE":    func(_ string, _ *Session, args []string) string { return server.HandleXrange(args) },
		"ZRANGE":    func(_ string, _ *Session, args []string) string { return server.HandleZrange(args) },
		"ZRANK":     func(_ string, _ *Session, args []string) string { return server.HandleZrank(args) },
		"ZCARD":     func(_ string, _ *Session, args []string) string { return server.HandleZcard(args) },
		"GEOPOS":    func(_ string, _ *Session, args []string) string { return server.HandleGeopos(args) },
		"GEODIST":   func(_ string, _ *Session, args []string) string { return server.HandleGeodist(args) },
		"GEOSEARCH": func(_ string, _ *Session, args []string) string { return server.HandleGeosearch(args) },
		"ZSCORE":    func(_ string, _ *Session, args []string) string { return server.HandleZscore(args) },
		"WAIT":      func(_ string, _ *Session, args []string) string { return server.HandleWait(args) },

		"PING":    func(_ string, session *Session, _ []string) string { return server.HandlePing(session) },
		"UNWATCH": func(_ string, session *Session, _ []string) string { return server.HandleUnwatch(session) },

		"SET":    func(input string, session *Session, args []string) string { return server.HandleSet(args) },
		"ZREM":   func(input string, _ *Session, args []string) string { return server.HandleZrem(args) },
		"RPUSH":  func(input string, _ *Session, args []string) string { return server.HandlePush(args, Right) },
		"LPUSH":  func(input string, _ *Session, args []string) string { return server.HandlePush(args, Left) },
		"CONFIG": func(input string, _ *Session, args []string) string { return server.HandleConfig(args) },
		"GEOADD": func(input string, _ *Session, args []string) string { return server.HandleGeoadd(args) },
		"ZADD":   func(input string, _ *Session, args []string) string { return server.HandleZadd(args) },
		"INCR":   func(input string, _ *Session, args []string) string { return server.HandleIncr(args) },
		"XADD":   func(input string, _ *Session, args []string) string { return server.HandleXadd(args) },
		"LPOP":   func(input string, _ *Session, args []string) string { return server.HandleLpop(args) },
		"BLPOP":  func(input string, _ *Session, args []string) string { return server.HandleBlpop(args) },

		"ACL":         func(_ string, session *Session, args []string) string { return server.HandleAcl(session, args) },
		"AUTH":        func(_ string, session *Session, args []string) string { return server.HandleAuth(session, args) },
		"REPLCONF":    func(_ string, session *Session, args []string) string { return server.HandleReplconf(session, args) },
		"SUBSCRIBE":   func(_ string, session *Session, args []string) string { return server.HandleSubscribe(session, args) },
		"PUBLISH":     func(_ string, session *Session, args []string) string { return server.HandlePublish(session, args) },
		"UNSUBSCRIBE": func(_ string, session *Session, args []string) string { return server.HandleUnsubscribe(session, args) },

		"MULTI": func(_ string, session *Session, _ []string) string {
			server.HandleMulti(session)
			return ""
		},

		"PSYNC": func(_ string, session *Session, _ []string) string {
			server.HandlePsync(session)
			return ""
		},

		"WATCH": func(input string, session *Session, args []string) string { return server.HandleWatch(session, args) },
	}
}
