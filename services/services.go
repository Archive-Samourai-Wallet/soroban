package services

import (
	"errors"
	"fmt"

	soroban "code.samourai.io/wallet/samourai-soroban"
)

var (
	RegistrationErr = errors.New("Service registration failed")
)

type NamedService struct {
	Name    string
	Service soroban.Service
}

func RegisterAll(server soroban.Soroban) error {
	services := []NamedService{
		{"directory", new(Directory)},
	}

	for _, ns := range services {
		err := server.Register(ns.Name, ns.Service)
		if err != nil {
			return RegistrationErr
		}
		fmt.Printf("%s registered\n", ns.Name)
	}

	return nil
}
