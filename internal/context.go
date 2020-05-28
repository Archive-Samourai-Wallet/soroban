package internal

import (
	"context"

	soroban "code.samourai.io/wallet/samourai-soroban"
)

const (
	SorobanDirectoryKey = "soroban-directory"
)

func DirectoryFromContext(ctx context.Context) soroban.Directory {
	storage, _ := ctx.Value(SorobanDirectoryKey).(soroban.Directory)
	return storage
}
