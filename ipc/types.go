package ipc

import (
	"context"
	"time"
)

type MessageType string

type Message struct {
	Type      MessageType `json:"type"`
	Message   string      `json:"message"`
	Payload   string      `json:"payload,omitempty"`
	Timestamp *time.Time  `json:"timestamp,omitempty"`
}

type MessageHandler func(ctx context.Context, message Message) (Message, error)

const (
	MessageTypeDebug   MessageType = "debug"
	MessageTypeSoroban MessageType = "soroban"
)
