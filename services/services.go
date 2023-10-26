package services

import (
	"context"
	"errors"

	soroban "code.samourai.io/wallet/samourai-soroban"
	log "github.com/sirupsen/logrus"
)

var (
	ErrRegistration = errors.New("service registration failed")
)

type NamedService struct {
	Name    string
	Service soroban.Service
}

func RegisterAll(ctx context.Context, server soroban.Soroban) error {
	services := []NamedService{
		{"directory", new(Directory)},
	}

	for _, ns := range services {
		err := server.Register(ctx, ns.Name, ns.Service)
		if err != nil {
			return ErrRegistration
		}
		log.Debugf("%s registered\n", ns.Name)
	}

	return nil
}
