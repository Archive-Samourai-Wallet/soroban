package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

func natsConnect(ctx context.Context, natsHost string, natsPort int) (*nats.Conn, error) {
	natsURL := fmt.Sprintf("nats://%s:%d", natsHost, natsPort)
	return natsURLConnect(ctx, natsURL)
}

func natsURLConnect(ctx context.Context, natsURL string) (*nats.Conn, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to NATS")
		return nil, err
	}
	log.WithField("URL", natsURL).
		Debug("Connected to NATS")

	return nc, nil
}

func startNatsServer(ctx context.Context, ipcSubject, natsHost string, natsPort int, handler MessageHandler, done chan struct{}) {
	// start enbedded nats server
	ns, err := server.NewServer(&server.Options{
		Host: natsHost,
		Port: natsPort,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to start embedded nats")
	}

	go ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		log.Fatal("Not ready for connection")
	}

	// Connect to nats server
	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to NATS")
	}

	subject := fmt.Sprintf("%s.%s", ipcSubject, "up")
	log.WithField("subject", subject).Info("Server register for child requests")

	for i := 0; i < 16; i++ {
		sub, err := nc.QueueSubscribe(subject, "queue."+subject, func(msg *nats.Msg) {
			handleNatsMessage(ctx, msg, handler)
		})
		if err != nil {
			log.WithError(err).
				Fatal("Failed to subscribe")
			return
		}
		defer sub.Unsubscribe()
	}

	log.Info("IPC server ready")
	done <- struct{}{}

	ns.WaitForShutdown()
	log.Info("IPC server stopped")
}

func handleNatsMessage(ctx context.Context, msg *nats.Msg, handler MessageHandler) {
	var response Message

	// return message response if reply is needed
	defer func(resp *Message) {
		if len(msg.Reply) == 0 {
			// NOOP
			return
		}

		// send response
		var data []byte
		if resp != nil {
			var err error
			data, err = json.Marshal(resp)
			if err != nil {
				log.WithError(err).
					Warning("Failed to Marshal response")
				return
			}
		}
		err := msg.Respond(data)
		if err != nil {
			log.WithError(err).
				Warning("NATS Respond Failed")
			return
		}
	}(&response)

	// process message
	var message Message
	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		log.WithError(err).
			WithField("Data", string(msg.Data)).
			Warning("Failed to Unmarshal message")
		response = Message{
			Type:    "error",
			Message: "Failed to Unmarshal message",
		}
		return
	}
	if len(message.Type) == 0 {
		log.Warning("Unknown Message Type")
		response = Message{
			Type:    "error",
			Message: "Unknown Message Type",
		}
		return
	}

	response, err = handler(ctx, message)
	if err != nil {
		log.WithError(err).
			Error("Message handler failed")
		return
	}
}
