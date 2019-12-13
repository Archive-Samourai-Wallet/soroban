package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	soroban "code.samourai.io/wallet/samourai-soroban"
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
		soroban.GenKey(prefix)
		return nil
	}

	server := soroban.New(soroban.Options{
		Redis: soroban.OptionRedis{
			Hostname: "localhost",
			Port:     6379,
		},
	})
	if server == nil {
		return errors.New(("Fails to create Soroban server"))
	}

	channel := new(services.Channel)
	err := server.Register(channel, "channel")
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println("Staring soroban...")
	err = server.Start(seed)
	if err != nil {
		return err
	}
	defer server.Stop()

	<-server.Ready

	fmt.Printf("Sordoban started: http://%s.onion\n", server.ID())

	<-context.Background().Done()

	return nil
}
