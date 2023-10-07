package soroban

import (
	"context"
	"sync"

	"github.com/cretz/bine/tor"

	log "github.com/sirupsen/logrus"
)

type ContextKey string

const (
	TorClientsKeys = ContextKey("soroban-torclients")
)

type torClientsInfo struct {
	sync.Mutex
	clients []*tor.Tor
}

func WithTorContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, TorClientsKeys, &torClientsInfo{})
}

func AddTorClient(ctx context.Context, torClient *tor.Tor) {
	torContext := ctx.Value(TorClientsKeys)
	if torContext == nil {
		panic("Create context with Tor")
	}
	if torClients, ok := torContext.(*torClientsInfo); ok {
		torClients.Lock()
		defer torClients.Unlock()

		torClients.clients = append(torClients.clients, torClient)
		log.Info("Tor client added")
	}
}

func Shutdown(ctx context.Context) {
	log.Warning("Shutting down all tor processes")
	torContext := ctx.Value(TorClientsKeys)
	if torContext == nil {
		panic("Create context with Tor")
	}
	if torClients, ok := torContext.(*torClientsInfo); ok {
		torClients.Lock()
		defer torClients.Unlock()

		for _, torClient := range torClients.clients {
			err := torClient.Close()
			if err != nil {
				log.WithError(err).Error("Failed to close tor client")
			} else {
				log.Warning("Tor client closed")
			}
		}
		torClients.clients = torClients.clients[:0]
	}
}
