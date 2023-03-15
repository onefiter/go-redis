package database

import "github.com/go-redis/interface/resp"

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close()
	AfterClientClose(c resp.Connection)
}

type DataEntity struct {
	Data interface{}
}
