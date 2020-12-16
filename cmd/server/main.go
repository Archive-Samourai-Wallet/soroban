package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/server"

	"code.samourai.io/wallet/samourai-soroban/services"
)

var (
	prefix string

	domain string
	seed   string
	export string

	withTor  bool
	hostname string
	port     int

	directoryType string
	directoryHost string
	directoryPort int
)

func init() {
	// GenKey
	flag.StringVar(&prefix, "prefix", "", "Generate Onion with prefix")

	// Server
	flag.StringVar(&domain, "domain", "", "Directory Domain")
	flag.StringVar(&seed, "seed", "", "Onion private key seed")
	flag.StringVar(&export, "export", "", "Export hidden service secret key from seed to file")

	flag.BoolVar(&withTor, "withTor", false, "Hidden service enabled (default false)")
	flag.StringVar(&hostname, "hostname", "localhost", "server address (default localhost)")
	flag.IntVar(&port, "port", 4242, "Server port (default 4242)")

	flag.StringVar(&directoryHost, "directoryType", "", "Directory Type (default, redis)")
	flag.StringVar(&directoryHost, "directoryHostname", "", "Directory host")
	flag.IntVar(&directoryPort, "directoryPort", 0, "Directory host")

	flag.Parse()

	if len(domain) == 0 {
		domain = "samourai"
	}

	if len(export) != 0 {
		withTor = true
	}
	if !withTor && len(seed) != 0 {
		log.Fatalf("Can't use seed without tor")
	}

	if len(directoryType) == 0 {
		directoryType = "default"
	}
	if len(directoryHost) == 0 {
		directoryHost = "localhost"
	}
	if directoryPort == 0 {
		directoryPort = 6379
	}
}

func main() {
	// export seed & exit
	if len(export) > 0 && len(seed) > 0 {
		data, err := server.ExportHiddenServiceSecret(seed)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(export, data, 0600)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

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
			Domain:        domain,
			DirectoryType: directoryType,
			Directory: soroban.ServerInfo{
				Hostname: directoryHost,
				Port:     directoryPort,
			},
			WithTor: withTor,
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
	if withTor {
		err = soroban.StartWithTor(port, seed)
	} else {
		err = soroban.Start(hostname, port)
	}
	if err != nil {
		return err
	}
	defer soroban.Stop()

	soroban.WaitForStart()

	if len(soroban.ID()) != 0 {
		fmt.Printf("Soroban started: http://%s.onion\n", soroban.ID())
	} else {
		fmt.Printf("Soroban started: http://%s:%d/\n", hostname, port)
	}

	WaitForExit(ctx)

	return nil
}

func WaitForExit(ctx context.Context) {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("Soroban exited")
		done <- true
	}()

	select {
	case <-done:
		return

	case <-ctx.Done():
		return
	}
}
