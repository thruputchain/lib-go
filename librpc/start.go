package librpc

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"

	"github.com/groundhogchain/lib-go/liblog"
	"github.com/groundhogchain/lib-go/libnet"
	"github.com/groundhogchain/lib-go/must"
	"github.com/xdrpp/goxdr/rpc"
	"github.com/xdrpp/goxdr/xdr"
)

func NewDriverSerial(
	ctx context.Context,
	sock net.Conn,
	maxMsgSize int,
	recvQueueLen int,
	sendQueueLen int,
	numProc int,
) *rpc.Driver {
	mp := NewRpcMsgPool()
	driver := rpc.NewDriverTuned(
		ctx, mp, NewRpcCallPool(), rpc.NewStreamTransportWithPool(sock, mp, maxMsgSize),
		recvQueueLen, sendQueueLen, numProc,
	)
	_ = liblog.Writer
	// driver.Log = log.Writer()
	driver.PanicHandler = rpc.PanicHandlerFunc(must.LogPanic)
	return driver
}

func NewDriverParallel(
	ctx context.Context,
	sock net.Conn,
	maxMsgSize int,
	recvQueueLen int,
	sendQueueLen int,
	numProc int,
) *rpc.Driver {
	mp := NewRpcMsgPool()
	driver := rpc.NewDriverTuned(
		ctx, mp, NewRpcCallPool(), rpc.NewStreamTransportWithPool(sock, mp, maxMsgSize),
		recvQueueLen, sendQueueLen, numProc,
	)
	driver.Lock = nil // process all calls in parallel
	_ = liblog.Writer
	// driver.Log = log.Writer()
	driver.PanicHandler = rpc.PanicHandlerFunc(must.LogPanic)
	return driver
}

func StartSerial(
	ctx context.Context,
	n libnet.Network,
	maxMsgSize int,
	recvQueueLen int,
	sendQueueLen int,
	numProc int,
	addr string,
	srv ...xdr.XdrSrv,
) net.Listener {
	l, err := n.ListenAddress(ctx, addr)
	must.NoError(ctx, err)
	go func() {
		// the only exit from the loop is when l is already closed
		// defer l.Close()
		for {
			sock, err := l.Accept()
			if (err != nil && strings.Contains(err.Error(), "use of closed network connection")) ||
				errors.Is(err, io.EOF) {
				liblog.Errorf("rpc server exiting (%v)", err)
				break
			}
			must.NoError(ctx, err)
			driver := NewDriverSerial(ctx, sock, maxMsgSize, recvQueueLen, sendQueueLen, numProc)
			for _, srv := range srv {
				driver.Register(srv)
			}
			go driver.Go()
		}
	}()
	return l
}

func StartParallel(
	ctx context.Context,
	n libnet.Network,
	maxMsgSize int,
	recvQueueLen int,
	sendQueueLen int,
	numProc int,
	addr string,
	srv ...xdr.XdrSrv,
) net.Listener {
	l, err := n.ListenAddress(ctx, addr)
	must.NoError(ctx, err)
	go func() {
		// the only exit from the loop is when l is already closed
		// defer l.Close()
		for {
			sock, err := l.Accept()
			if (err != nil && strings.Contains(err.Error(), "use of closed network connection")) ||
				errors.Is(err, io.EOF) {
				liblog.Errorf("rpc server exiting (%v)", err)
				break
			}
			must.NoError(ctx, err)
			driver := NewDriverParallel(ctx, sock, maxMsgSize, recvQueueLen, sendQueueLen, numProc)
			for _, srv := range srv {
				driver.Register(srv)
			}
			go driver.Go()
		}
	}()
	return l
}
