package tcp

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-redis/interface/tcp"
	"github.com/go-redis/lib/logger"
)

type Config struct {
	Address string
}

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {

	listenser, err := net.Listen("tcp", cfg.Address)

	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
			closeChan <- struct{}{}

		}
	}()

	if err != nil {
		return err
	}
	logger.Info("start listen")

	ListenAndServe(listenser, handler, closeChan)

	return nil

}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {

	// listen signal
	go func() {
		<-closeChan
		logger.Info("shutting down...")
		_ = listener.Close() // listener.Accept() will return err immediately
		_ = handler.Close()  // close connections
	}()

	// listen port
	defer func() {
		// close during unexpected error
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	// 1.panic之前，新链接不再接收，但是要等接收的链接处理完
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()

		if err != nil {
			break
		}

		logger.Info("accepted link")
		waitDone.Add(1)

		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()

	}

	waitDone.Wait()

}
