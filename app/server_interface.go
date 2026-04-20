package main

import "net"

type IServer interface {
	DiscardWithoutMulti() string
	ExecWithoutMulti() string
	ExecuteCommand(session *Session, command string, args []string, input string) string
	LoadRdb()
	HandleConnection(conn net.Conn)
	HandleCommand(session *Session, input string) string
	HandlePing(session *Session) string
	HandleEcho(args []string) string
	HandleAcl(session *Session, args []string) string
	HandleAuth(session *Session, arr []string) string
	HandleGeoadd(arr []string) string
	HandleGeopos(arr []string) string
	HandleGeodist(arr []string) string
	HandleGeosearch(arr []string) string
	HandlePush(arr []string, direction Direction) string
	HandleLrange(arr []string) string
	HandleLlen(arr []string) string
	HandleLpop(arr []string) string
	HandleBlpop(arr []string) string
	HandleWatch(session *Session, args []string) string
	HandleUnwatch(session *Session) string
	HandleSubscribe(session *Session, args []string) string
	HandleUnsubscribe(session *Session, args []string) string
	HandlePublish(session *Session, args []string) string
	HandleConfig(args []string) string
	HandleKeys() string
	HandleInfo(arr []string) string
	Handshake(masterHost string, masterPort string)
	HandlePsync(session *Session) string
	HandleReplconf(session *Session, arr []string) string
	HandleWait(arr []string) string
	HandleZadd(arr []string) string
	HandleZrank(arr []string) string
	HandleZcard(arr []string) string
	HandleZrange(arr []string) string
	HandleZscore(arr []string) string
	HandleZrem(arr []string) string
	HandleType(arr []string) string
	HandleXadd(arr []string) string
	HandleXrange(arr []string) string
	HandleIncr(arr []string) string
	HandleMulti(session *Session)
	HandleExec(session *Session)
	XreadLastCase(ids, keys []string) []string
}
