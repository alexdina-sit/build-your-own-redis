package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

type Direction string

const (
	CRLF            = "\r\n"
	Left  Direction = "left"
	Right Direction = "right"
)

var (
	allowedCommandsSubscribed = []string{"SUBSCRIBE", "UNSUBSCRIBE", "PSUBSCRIBE", "PUNSUBSCRIBE", "PING", "QUIT"}

	portFlag    = flag.String("port", "6379", "The port to listen on")
	replicaFlag = flag.String("replicaof", "", "Master and slave ports")

	dirFlag            = flag.String("dir", "", "File dir")
	dbfilenameFlag     = flag.String("dbfilename", "", "RDB dbfilename")
	appendonlyFlag     = flag.String("appendonly", "", "Appendonly flag")
	appenddirnameFlag  = flag.String("appenddirname", "", "Appenddirname flag")
	appendfilenameFlag = flag.String("appendfilename", "", "Appendfilename flag")
	appendfsyncFlag    = flag.String("appendfsync", "", "Appendfsync flag")
)

func main() {
	flag.Parse()

	server := GetServerInstance()
	server.usersMap["default"] = &User{Flags: []string{"nopass"}}

	server.Config = NewConfig(
		dirFlag,
		dbfilenameFlag,
		appendonlyFlag,
		appenddirnameFlag,
		appendfilenameFlag,
		appendfsyncFlag,
	)

	server.LoadRdb()

	if server.Config.AppendOnly == "yes" {
		server.LoadAOF()
	}

	if replicaFlag != nil && *replicaFlag != "" {
		server.Role = "slave"

		masterHost, masterPort, err := getMasterAddress(*replicaFlag)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		go server.Handshake(masterHost, masterPort)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", *portFlag))
	if err != nil {
		fmt.Printf("Failed to bind to port: %s\n", *portFlag)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		go server.HandleConnection(conn)
	}
}
