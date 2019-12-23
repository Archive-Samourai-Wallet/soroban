package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/server"

	"code.samourai.io/wallet/samourai-soroban/services"
)

var (
	prefix string
	seed   string
)

func init() {
	flag.StringVar(&prefix, "prefix", "", "Generate Onion with prefix")
	flag.StringVar(&seed, "seed", "", "Onion private key seed")
	flag.Parse()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(prefix) > 0 {
		server.GenKey(prefix)
		return nil
	}

	ctx := context.Background()

	soroban := server.New(ctx,
		soroban.Options{
			Directory: soroban.ServerInfo{
				Hostname: "localhost",
				Port:     6379,
			},
		},
	)
	if soroban == nil {
		return errors.New("Fails to create Soroban server")
	}

	err := services.RegisterAll(soroban)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println("Staring soroban...")
	err = soroban.Start(seed)
	if err != nil {
		return err
	}
	defer soroban.Stop()

	soroban.WaitForStart()

	fmt.Printf("Sordoban started: http://%s.onion\n", soroban.ID())

	<-ctx.Done()

	return nil
}
