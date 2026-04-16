package main

func (server *Server) initCommands() {
	server.commandsMap = map[string]func(cmd string, session *Session, args []string) string{
		"PING":    func(_ string, _ *Session, _ []string) string { return server.handlePing() },
		"EXEC":    func(_ string, _ *Session, _ []string) string { return server.execWithoutMulti() },
		"DISCARD": func(_ string, _ *Session, _ []string) string { return server.discardWithoutMulti() },

		"ECHO":      func(cmd string, _ *Session, args []string) string { return server.handleEcho(args) },
		"GET":       func(cmd string, _ *Session, args []string) string { return server.handleGet(args) },
		"LRANGE":    func(cmd string, _ *Session, args []string) string { return server.handleLrange(args) },
		"LLEN":      func(cmd string, _ *Session, args []string) string { return server.handleLlen(args) },
		"LPOP":      func(cmd string, _ *Session, args []string) string { return server.handleLpop(args) },
		"BLPOP":     func(cmd string, _ *Session, args []string) string { return server.handleBlpop(args) },
		"TYPE":      func(cmd string, _ *Session, args []string) string { return server.handleType(args) },
		"INFO":      func(cmd string, _ *Session, args []string) string { return server.handleInfo(args) },
		"INCR":      func(cmd string, _ *Session, args []string) string { return server.handleIncr(args) },
		"XADD":      func(cmd string, _ *Session, args []string) string { return server.handleXadd(args) },
		"XREAD":     func(cmd string, _ *Session, args []string) string { return server.handleXread(args) },
		"XRANGE":    func(cmd string, _ *Session, args []string) string { return server.handleXrange(args) },
		"ZADD":      func(cmd string, _ *Session, args []string) string { return server.handleZadd(args) },
		"ZRANGE":    func(cmd string, _ *Session, args []string) string { return server.handleZrange(args) },
		"ZRANK":     func(cmd string, _ *Session, args []string) string { return server.handleZrank(args) },
		"ZCARD":     func(cmd string, _ *Session, args []string) string { return server.handleZcard(args) },
		"GEOADD":    func(cmd string, _ *Session, args []string) string { return server.handleGeoadd(args) },
		"GEOPOS":    func(cmd string, _ *Session, args []string) string { return server.handleGeopos(args) },
		"GEODIST":   func(cmd string, _ *Session, args []string) string { return server.handleGeodist(args) },
		"GEOSEARCH": func(cmd string, _ *Session, args []string) string { return server.handleGeosearch(args) },
		"ZSCORE":    func(cmd string, _ *Session, args []string) string { return server.handleZscore(args) },
		"ZREM":      func(cmd string, _ *Session, args []string) string { return server.handleZrem(args) },
		"WAIT":      func(cmd string, _ *Session, args []string) string { return server.handleWait(args) },
		"RPUSH":     func(cmd string, _ *Session, args []string) string { return server.handlePush(args, Right) },
		"LPUSH":     func(cmd string, _ *Session, args []string) string { return server.handlePush(args, Left) },

		"ACL":      func(cmd string, session *Session, args []string) string { return server.handleAcl(session, args) },
		"AUTH":     func(cmd string, session *Session, args []string) string { return server.handleAuth(session, args) },
		"REPLCONF": func(cmd string, session *Session, args []string) string { return server.handleReplconf(session, args) },

		"MULTI": func(cmd string, session *Session, _ []string) string {
			server.handleMulti(session)
			return ""
		},

		"PSYNC": func(cmd string, session *Session, args []string) string {
			server.handlePsync(session)
			return ""
		},
	}
}
