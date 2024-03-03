package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type IPCOptions struct {
	Mode     string
	Subject  string
	NatsHost string
	NatsPort int
}

type IPCService struct {
	options IPCOptions
	conn    *nats.Conn
}

// NewServer start a IPC server with embedeed NATS server
func New(ctx context.Context, options IPCOptions) *IPCService {
	return &IPCService{
		options: options,
	}
}

func (p *IPCService) Mode() string {
	return p.options.Mode
}

// Connect to IPC server with NATS.
func (p *IPCService) Connect(ctx context.Context) {
	if p.conn != nil {
		log.Warning("IPC Client is already connected")
		return
	}
	conn, err := natsConnect(ctx, p.options.NatsHost, p.options.NatsPort)
	if err != nil {
		log.WithError(err).
			Panic("Failed to connect to NATS")
	}
	p.conn = conn

	// check if responder is present
	_, err = p.Request(Message{
		Type:    MessageTypeDebug,
		Message: "Client Init",
		Payload: "{}",
	}, "up")
	if err != nil {
		log.WithError(err).
			Error("Failed to reach IPC server")
	}

	log.Debug("IPC Client Connected")
}

func (p *IPCService) Start(ctx context.Context, handler MessageHandler) {
	if handler == nil {
		log.Fatal("Invalid message handler")
		return
	}

	done := make(chan struct{})
	go startNatsServer(ctx, p.options.Subject, p.options.NatsHost, p.options.NatsPort, handler, done)

	<-done
	close(done)
}

func (p *IPCService) Request(request Message, direction string) (Message, error) {
	data, err := json.Marshal(&request)
	if err != nil {
		return Message{}, err
	}

	subject := fmt.Sprintf("%s.%s", p.options.Subject, direction)
	log.WithField("subject", subject).Debug("IPC Requests")
	msg, err := p.conn.Request(subject, data, 5*time.Second)
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

func (p *IPCService) ListenFromServer(ctx context.Context, ipcSubject string, ipcHandler MessageHandler) {
	subject := fmt.Sprintf("%s.%s", ipcSubject, "down")
	log.WithField("subject", subject).Info("Child register for server requests")
	for i := 0; i < 16; i++ {
		sub, err := p.conn.QueueSubscribe(subject, "queue."+subject, func(msg *nats.Msg) {
			log.Warning("Message recieved from IPC server")
			handleNatsMessage(ctx, msg, ipcHandler)
		})
		if err != nil {
			log.WithError(err).
				Fatal("Failed to subscribe")
			return
		}
		defer sub.Unsubscribe()
	}

	<-ctx.Done()
}
