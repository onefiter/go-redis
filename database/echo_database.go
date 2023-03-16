package database

import (
	"github.com/go-redis/interface/resp"
	"github.com/go-redis/lib/logger"
	"github.com/go-redis/resp/reply"
)

type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	return reply.MakeMultiBulkReply(args)

}

func (e EchoDatabase) AfterClientClose(c resp.Connection) {
	logger.Info("EchoDatabase AfterClientClose")
}

func (e EchoDatabase) Close() {
	logger.Info("EchoDatabase Close")

}
