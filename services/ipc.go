package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal"
	"code.samourai.io/wallet/samourai-soroban/ipc"
	"code.samourai.io/wallet/samourai-soroban/p2p"
	log "github.com/sirupsen/logrus"
)

func StartIPCService(ctx context.Context, ready chan struct{}) {
	if ipcServer := internal.IPCFromContext(ctx); ipcServer != nil {
		ipcServer.Start(ctx, func(ctx context.Context, message ipc.Message) (ipc.Message, error) {
			directory := internal.DirectoryFromContext(ctx)

			return ipcHandler(ctx, directory, message)
		})
	} else {
		log.Fatal("IPC Server not found in context")
	}

	ready <- struct{}{}
}

func ipcHandler(ctx context.Context, directory soroban.Directory, message ipc.Message) (ipc.Message, error) {
	switch message.Type {
	case ipc.MessageTypeSoroban:
		var p2pMessage p2p.Message
		err := json.Unmarshal([]byte(message.Payload), &p2pMessage)
		if err != nil {
			log.WithError(err).
				Error("Failed to parse P2P message")
		}

		log.WithField("p2pMessage", fmt.Sprintf("%s: %s", p2pMessage.Context, string(p2pMessage.Payload))).Debug("Recieve message from IPC")

		var args DirectoryEntry

		err = p2pMessage.ParsePayload(&args)
		if err != nil {
			log.WithError(err).
				Error("Failed to Parse P2P payload")
			return ipc.Message{
				Type:    message.Type,
				Message: "error",
			}, nil
		}

		switch p2pMessage.Context {
		case "Directory.Add":
			err = addToDirectory(directory, &args)

		case "Directory.Remove":
			err = removeFromDirectory(directory, &args)

		default:
			err = errors.New("unknown p2p message context")

		}
		if err != nil {
			log.WithError(err).Error("failed to process message.")
			return ipc.Message{
				Type:    message.Type,
				Message: "error",
			}, nil
		}

		return ipc.Message{
			Type:    message.Type,
			Message: "success",
		}, nil
	default:
		// NOOP
		return ipc.Message{
			Type:    message.Type,
			Message: "success",
		}, nil
	}
}
