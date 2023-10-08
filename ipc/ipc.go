package ipc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type IPCOptions struct {
	Subject  string
	NatsHost string
	NatsPort int
}

type IPCServer struct {
	options IPCOptions
	conn    *nats.Conn
}

// NewServer start a IPC server with embedeed NATS server
func NewServer(ctx context.Context, options IPCOptions) *IPCServer {
	return &IPCServer{
		options: options,
	}
}

// New return a IPC Client connected to IPC server with NATS.
func New(ctx context.Context, options IPCOptions) *IPCServer {
	conn, err := natsConnect(ctx, options.NatsHost, options.NatsPort)
	if err != nil {
		log.WithError(err).
			Panic("Failed to connect to NATS")
	}

	client := IPCServer{
		options: options,
		conn:    conn,
	}

	// check if responder is present
	_, err = client.Request(Message{
		Type:    MessageTypeDebug,
		Message: "Client Init",
		Payload: "{}",
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to reach IPC server")
	}

	return &client

}

func (p *IPCServer) Start(ctx context.Context, handler MessageHandler) {
	if handler == nil {
		log.Fatal("Invalid message handler")
		return
	}

	done := make(chan struct{})
	go startNatsServer(ctx, p.options.Subject, p.options.NatsHost, p.options.NatsPort, handler, done)

	<-done
	close(done)
}

func (p *IPCServer) Request(request Message) (Message, error) {
	data, err := json.Marshal(&request)
	if err != nil {
		return Message{}, err
	}
	msg, err := p.conn.Request(p.options.Subject, data, 5*time.Second)
	if err != nil {
		return Message{}, err
	}

	var resp Message
	err = json.Unmarshal(msg.Data, &resp)
	if err != nil {
		return Message{}, err
	}

	return resp, nil
}
