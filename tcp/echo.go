package tcp

import (
	"bufio"
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/go-redis/lib/logger"
	"github.com/go-redis/lib/sync/atomic"
	"github.com/go-redis/lib/sync/wait"
)

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

func (e *EchoClient) Close() error {
	e.Waiting.WaitWithTimeout(10 * time.Second)
	_ = e.Conn.Close()
	return nil
}

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

// MakeEchoHandler creates EchoHandler
func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		// closing handler refuse new connection
		_ = conn.Close()
	}

	// 新链接包装成EchoClient，EchoClient就代表一个客户端
	client := &EchoClient{
		Conn: conn,
	}
	// 这里只需要一个hashset，只需要key
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for { // 不断读网络报文
		// may occurs: client EOF, client timeout, server early close
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF { // EOF
				logger.Info("connection close")
				h.activeConn.Delete(client)
			} else { //
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		// 转换为字节流
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}

}

func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// sync.Map
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		return true
	})
	return nil
}
