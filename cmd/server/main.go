package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/confidential"
	"code.samourai.io/wallet/samourai-soroban/server"

	"code.samourai.io/wallet/samourai-soroban/services"

	log "github.com/sirupsen/logrus"
)

var (
	logLevel string
	prefix   string

	config string
	domain string
	seed   string
	export string

	withTor  bool
	hostname string
	port     int

	directoryType string
	directoryHost string
	directoryPort int

	p2pSeed       string
	p2pBootstrap  string
	p2pListenPort int
	p2pRoom       string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.StringVar(&logLevel, "log", "info", "Log level (default info)")

	// GenKey
	flag.StringVar(&prefix, "prefix", "", "Generate Onion with prefix")

	// Server
	flag.StringVar(&config, "config", "", "Yaml configuration file for confidential keys")
	flag.StringVar(&domain, "domain", "", "Directory Domain")
	flag.StringVar(&seed, "seed", "", "Onion private key seed")
	flag.StringVar(&export, "export", "", "Export hidden service secret key from seed to file")

	flag.BoolVar(&withTor, "withTor", false, "Hidden service enabled (default false)")
	flag.StringVar(&hostname, "hostname", "localhost", "server address (default localhost)")
	flag.IntVar(&port, "port", 4242, "Server port (default 4242)")

	flag.StringVar(&directoryType, "directoryType", "", "Directory Type (default, redis, memory)")
	flag.StringVar(&directoryHost, "directoryHostname", "", "Directory host")
	flag.IntVar(&directoryPort, "directoryPort", 0, "Directory host")

	flag.StringVar(&p2pSeed, "p2pSeed", "auto", "P2P Onion private key seed")
	flag.StringVar(&p2pBootstrap, "p2pBootstrap", "", "P2P bootstrap")
	flag.IntVar(&p2pListenPort, "p2pListenPort", 1042, "P2P Listen Port")
	flag.StringVar(&p2pRoom, "p2pRoom", "samourai-p2p", "P2P Room")

	flag.Parse()

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)

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
	ctx = soroban.WithTorContext(ctx)

	if len(config) > 0 {
		go confidential.ConfigWatcher(ctx, config)
	}

	sorobanServer := server.New(ctx,
		soroban.Options{
			Domain:        domain,
			DirectoryType: directoryType,
			Directory: soroban.ServerInfo{
				Hostname: directoryHost,
				Port:     directoryPort,
			},
			WithTor: withTor,
			P2P: soroban.P2PInfo{
				Seed:       p2pSeed,
				Bootstrap:  p2pBootstrap,
				ListenPort: p2pListenPort,
				Room:       p2pRoom,
			},
		},
	)
	if sorobanServer == nil {
		return errors.New("Fails to create Soroban server")
	}

	err := services.RegisterAll(ctx, sorobanServer)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println("Staring soroban...")
	if withTor {
		err = sorobanServer.StartWithTor(ctx, port, seed)
	} else {
		err = sorobanServer.Start(ctx, hostname, port)
	}
	if err != nil {
		return err
	}
	defer sorobanServer.Stop(ctx)

	sorobanServer.WaitForStart(ctx)

	if len(sorobanServer.ID()) != 0 {
		fmt.Printf("Soroban started: http://%s.onion\n", sorobanServer.ID())
	} else {
		fmt.Printf("Soroban started: http://%s:%d/\n", hostname, port)
	}

	<-ctx.Done()
	return nil
}

func WaitForExit(ctx context.Context) {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		soroban.Shutdown(ctx)
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
