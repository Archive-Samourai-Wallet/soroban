package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal"
	"code.samourai.io/wallet/samourai-soroban/ipc"
	log "github.com/sirupsen/logrus"
)

func StartP2PDirectory(ctx context.Context, p2pSeed, bootstrap string, hostname string, listenPort int, room string, ready chan struct{}) {
	if len(bootstrap) == 0 {
		log.Error("Invalid bootstrap")
		return
	}
	if len(room) == 0 {
		log.Error("Invalid room")
		return
	}

	client := internal.IPCFromContext(ctx)
	sorobanMode := "peer"
	if client != nil {
		sorobanMode = client.Mode()
	}

	p2P := internal.P2PFromContext(ctx)
	if p2P == nil {
		log.Error("p2p - P2P not found")
		return
	}

	p2pReady := make(chan struct{})
	go func() {
		err := p2P.Start(ctx, p2pSeed, hostname, listenPort, bootstrap, room, p2pReady)
		if err != nil {
			log.WithError(err).Error("Failed to p2P.Start")
		}
		ready <- struct{}{}
	}()

	<-p2pReady

	timeoutDelay := 15 * time.Minute // first timeout is longer at startup
	lastHeartbeatTimestamp := time.Now().UTC()
	for {
		select {
		case message := <-p2P.OnMessage:
			var args DirectoryEntry

			err := message.ParsePayload(&args)
			if err != nil {
				log.WithError(err).Error("Failed to ParsePayload")
				continue
			}

			if args.Name == "p2p.heartbeat" {
				timeoutDelay = 3 * time.Minute // reduce timeout delay after first heartbeat received
				lastHeartbeatTimestamp = time.Now()

				log.Trace("p2p - heartbeat received")
				continue
			}

			log.WithField("message", fmt.Sprintf("%s: %s", message.Context, string(message.Payload))).Debug("Recieved message from p2p")

			switch sorobanMode {
			case "child":
				// foward P2P message to IPC server
				data, err := json.Marshal(message)
				if err != nil {
					log.WithError(err).Error("failed to marshal p2p message.")
					continue
				}
				message, err := client.Request(ipc.Message{
					Type:    ipc.MessageTypeSoroban,
					Payload: string(data),
				}, "up")
				if err != nil {
					log.WithError(err).Error("failed send ipc request.")
					continue
				}
				if message.Message != "success" {
					log.WithField("Message", message.Message).Warning("IPC Message failed")
				}
				log.WithField("Message", message.Message).Debug("IPC Message sent")
				continue

			default:
				// Default P2P mode, with directory available
				directory := internal.DirectoryFromContext(ctx)
				if directory == nil {
					log.Error("Directory not found")
					continue
				}

				switch message.Context {
				case "Directory.Add":
					err = addToDirectory(directory, &args)

				case "Directory.Remove":
					err = removeFromDirectory(directory, &args)
				}
				if err != nil {
					log.WithError(err).Error("failed to process message.")
					continue
				}

				if ipcClient := internal.IPCFromContext(ctx); ipcClient != nil {
					ipcClient.Request(ipc.Message{
						Type:    ipc.MessageTypeSoroban,
						Payload: string(message.Payload),
					}, "up")
				}

				continue
			}

		case <-time.After(30 * time.Second):
			if time.Since(lastHeartbeatTimestamp) > timeoutDelay {
				log.Warning("No message received from too long, exiting...")
				soroban.Shutdown(ctx)
				os.Exit(0)
			}

			err := p2P.PublishJson(ctx, "Directory.Add", DirectoryEntry{
				Name:  "p2p.heartbeat",
				Entry: fmt.Sprintf("%d", time.Now().Unix()),
				Mode:  "short",
			})
			if err != nil {
				// non fatal error
				log.Warningf("p2p - Failed to PublishJson. %s\n", err)
				continue
			}
			log.Trace("p2p - heartbeat sent")

		case <-ctx.Done():
			return
		}
	}
}
