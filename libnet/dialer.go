package libnet

import (
	"context"
	"io"
	"net"
	"syscall"
	"time"
)

type Network interface {
	Listener
	Dialer
	String() string
}

type Conn interface {
	io.Reader
	io.Writer
	io.Closer
}

type Listener interface {
	Listen(ctx context.Context, address string) (net.Listener, error)
}

type Dialer interface {
	Dial(ctx context.Context, address string) (net.Conn, error)
}

type TCPNet struct{}

func (TCPNet) String() string {
	return "tcp"
}

func (TCPNet) Listen(ctx context.Context, address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}

func (TCPNet) Dial(ctx context.Context, address string) (net.Conn, error) {
	deadline, ok := ctx.Deadline()
	if ok {
		return net.DialTimeout("tcp", address, time.Until(deadline))
	} else {
		return net.Dial("tcp", address)
	}
}

type TCPReuseNet struct{}

func (TCPReuseNet) String() string {
	return "tcp"
}

func (TCPReuseNet) Listen(ctx context.Context, address string) (net.Listener, error) {
	listenConfig := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var sockOptErr error
			if err := c.Control(func(fd uintptr) {
				if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
					sockOptErr = err
					return
				}
				if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1); err != nil {
					sockOptErr = err
				}
			}); err != nil {
				return err
			}
			return sockOptErr
		},
	}
	return listenConfig.Listen(ctx, "tcp", address)
}

func (TCPReuseNet) Dial(ctx context.Context, address string) (net.Conn, error) {
	deadline, ok := ctx.Deadline()
	if ok {
		return net.DialTimeout("tcp", address, time.Until(deadline))
	} else {
		return net.Dial("tcp", address)
	}
}
