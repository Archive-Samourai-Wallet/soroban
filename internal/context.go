package internal

import (
	"context"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/ipc"
	"code.samourai.io/wallet/samourai-soroban/p2p"
)

type ContextKey string

const (
	SorobanDirectoryKey = ContextKey("soroban-directory")
	SorobanP2PKey       = ContextKey("soroban-p2p")
	SorobanIPCKey       = ContextKey("soroban-ipc")
)

func DirectoryFromContext(ctx context.Context) soroban.Directory {
	storage, _ := ctx.Value(SorobanDirectoryKey).(soroban.Directory)
	return storage
}

func P2PFromContext(ctx context.Context) *p2p.P2P {
	result, _ := ctx.Value(SorobanP2PKey).(*p2p.P2P)
	return result
}

func IPCFromContext(ctx context.Context) *ipc.IPCService {
	result, _ := ctx.Value(SorobanIPCKey).(*ipc.IPCService)
	return result
}
