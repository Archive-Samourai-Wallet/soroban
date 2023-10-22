package server

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/ipc"

	log "github.com/sirupsen/logrus"
)

func startChildSoroban(ctx context.Context, options soroban.Options, childID int) {
	executablePath, err := os.Executable()
	if err != nil {
		log.WithError(err).
			Fatal("Failed to get current executable path")
	}

	if !strings.HasPrefix(executablePath, "/") {
		executablePath = path.Join("./", executablePath)
	}
	if !fileExists(executablePath) {
		log.Fatal("Soroban executable not found")
	}

	go ipc.StartProcessDaemon(ctx, fmt.Sprintf("soroban-child-%d", childID),
		executablePath,
		"--config", options.Config,
		"--ipcChildID", strconv.Itoa(childID),
		"--ipcNatsHost", options.IPC.NatsHost,
		"--ipcNatsPort", strconv.Itoa(options.IPC.NAtsPort),
		"--p2pBootstrap", options.P2P.Bootstrap,
		"--p2pRoom", options.P2P.Room,
		"--p2pListenPort", strconv.Itoa(options.P2P.ListenPort+childID),
		"--log", log.GetLevel().String(),
	)

}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}
