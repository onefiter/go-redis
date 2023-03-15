package tcp

import (
	"context"
	"net"
)

type Handler interface {
	// 处理连接的接口
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
