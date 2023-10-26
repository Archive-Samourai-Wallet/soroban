package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/server"

	"code.samourai.io/wallet/samourai-soroban/services"

	log "github.com/sirupsen/logrus"
)

var (
	options soroban.Options = soroban.DefaultOptions

	prefix string
	export string
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.StringVar(&options.LogLevel, "log", options.LogLevel, "Log level (default info)")
	flag.StringVar(&options.LogFile, "logfile", options.LogFile, "Log file (default -)")

	// GenKey
	flag.StringVar(&prefix, "prefix", prefix, "Generate Onion with prefix")
	flag.StringVar(&export, "export", "", "Export hidden service secret key from seed to file")

	// Server
	flag.StringVar(&options.Soroban.Config, "config", options.Soroban.Config, "Yaml configuration file for soroban")
	flag.StringVar(&options.Soroban.Confidential, "confidential", options.Soroban.Confidential, "Yaml configuration file for confidential keys")
	flag.StringVar(&options.Soroban.Domain, "domain", options.Soroban.Domain, "Directory Domain")
	flag.StringVar(&options.Soroban.Seed, "seed", options.Soroban.Seed, "Onion private key seed")

	flag.BoolVar(&options.Soroban.WithTor, "withTor", options.Soroban.WithTor, "Hidden service enabled (default false)")
	flag.StringVar(&options.Soroban.Hostname, "hostname", options.Soroban.Hostname, "server address (default localhost)")
	flag.IntVar(&options.Soroban.Port, "port", options.Soroban.Port, "Server port (default 4242)")

	flag.StringVar(&options.Soroban.DirectoryType, "directoryType", options.Soroban.DirectoryType, "Directory Type (default, redis, memory)")

	flag.StringVar(&options.P2P.Seed, "p2pSeed", options.P2P.Seed, "P2P Onion private key seed")
	flag.StringVar(&options.P2P.Bootstrap, "p2pBootstrap", options.P2P.Bootstrap, "P2P bootstrap")
	flag.IntVar(&options.P2P.ListenPort, "p2pListenPort", options.P2P.ListenPort, "P2P Listen Port")
	flag.StringVar(&options.P2P.Room, "p2pRoom", options.P2P.Room, "P2P Room")

	flag.StringVar(&options.IPC.Subject, "ipcSubject", options.IPC.Subject, "IPC communication subject")
	flag.IntVar(&options.IPC.ChildID, "ipcChildID", options.IPC.ChildID, "IPC child ID")
	flag.IntVar(&options.IPC.ChildProcessCount, "ipcChildProcessCount", options.IPC.ChildProcessCount, "Spawn child process")
	flag.StringVar(&options.IPC.NatsHost, "ipcNatsHost", options.IPC.NatsHost, "IPC NATS host")
	flag.IntVar(&options.IPC.NatsPort, "ipcNatsPort", options.IPC.NatsPort, "IPC nats port")

	flag.Parse()

	options.Load(options.Soroban.Config)

	level, err := log.ParseLevel(options.LogLevel)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)

	logOutput := os.Stderr
	if len(options.LogFile) > 0 && options.LogFile != "-" {
		log.SetFormatter(&log.JSONFormatter{})
		var err error
		if logOutput, err = os.OpenFile(options.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600); err != nil {
			log.WithField("logOutput", logOutput).Fatal("Invalid ConfigFile")
			return
		}
	}
	log.SetOutput(logOutput)

	if len(export) != 0 {
		options.Soroban.WithTor = true
	}
	if !options.Soroban.WithTor && len(options.Soroban.Seed) != 0 {
		log.Fatalf("Can't use seed without tor")
	}
}

func main() {
	// export seed & exit
	if len(export) > 0 && len(options.Soroban.Seed) > 0 {
		data, err := server.ExportHiddenServiceSecret(options.Soroban.Seed)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(export, data, 0600)
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
	prefix = strings.Trim(prefix, " ")

	ctx := context.Background()
	ctx = soroban.WithTorContext(ctx)

	ctx, sorobanServer := server.New(ctx, options)
	if sorobanServer == nil {
		// soroban is in child mode
		// keep the process alive
		log.Info("Soroban started in child mode")
		<-ctx.Done()
		return nil
	}

	err := services.RegisterAll(ctx, sorobanServer)
	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Info("Staring soroban...")
	if options.Soroban.WithTor {
		err = sorobanServer.StartWithTor(ctx, options.Soroban.Hostname, options.Soroban.Port, options.Soroban.Seed)
	} else {
		err = sorobanServer.Start(ctx, options.Soroban.Hostname, options.Soroban.Port)
	}
	if err != nil {
		return err
	}
	defer sorobanServer.Stop(ctx)

	sorobanServer.WaitForStart(ctx)

	if len(sorobanServer.ID()) != 0 {
		log.Infof("Soroban started: http://%s.onion", sorobanServer.ID())
	} else {
		log.Infof("Soroban started: http://%s:%d/", options.Soroban.Hostname, options.Soroban.Port)
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
