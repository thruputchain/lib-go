package librpc

import (
	"github.com/xdrpp/goxdr/rpc"
)

func NewRpcMsgPool() rpc.MessagePool {
	return rpc.NewMsgArena(5000)
}

func NewRpcCallPool() rpc.CallPool {
	return rpc.NewCallArena(5000)
}
