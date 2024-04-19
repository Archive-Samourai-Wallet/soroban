package server

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"crypto"
	"crypto/ed25519"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/confidential"
	"code.samourai.io/wallet/samourai-soroban/internal"
	"code.samourai.io/wallet/samourai-soroban/ipc"
	"code.samourai.io/wallet/samourai-soroban/p2p"
	"code.samourai.io/wallet/samourai-soroban/services"

	"github.com/cretz/bine/tor"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	gjson "github.com/gorilla/rpc/json"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

type Soroban struct {
	p2p       *p2p.P2P
	ipc       *ipc.IPCService
	directory soroban.Directory
	t         *tor.Tor
	onion     *tor.OnionService
	started   chan bool
	rpcServer *rpc.Server
}

func New(ctx context.Context, options soroban.Options) (context.Context, *Soroban) {
	var directory soroban.Directory

	if len(options.Soroban.Confidential) > 0 {
		go confidential.ConfigWatcher(ctx, options.Soroban.Confidential)
	}

	switch options.Soroban.DirectoryType {
	case "memory":
		directory = internal.NewDirectory(options.Soroban.Domain, internal.DirectoryTypeMemory)
	case "default":
		directory = internal.DefaultDirectory(options.Soroban.Domain)
	}
	if directory == nil {
		log.Fatal("Invalid Directory")
	}

	startIPCService := options.IPC.ChildProcessCount > 0 && options.IPC.ChildID == 0
	startMainSoroban := startIPCService || (options.IPC.ChildProcessCount == 0 && options.IPC.ChildID == 0)
	startP2PDirectory := len(options.P2P.Bootstrap) > 0 && (options.IPC.ChildProcessCount == 0 || options.IPC.ChildID > 0)

	ipcMode := "peer"
	if !startMainSoroban {
		ipcMode = "child"
	}

	ctx = context.WithValue(ctx, internal.SorobanDirectoryKey, directory)
	ctx = context.WithValue(ctx, internal.SorobanP2PKey, &p2p.P2P{OnMessage: make(chan p2p.Message)})
	ctx = context.WithValue(ctx, internal.SorobanIPCKey, ipc.New(ctx, ipc.IPCOptions{
		Mode:     ipcMode,
		Subject:  options.IPC.Subject,
		NatsHost: options.IPC.NatsHost,
		NatsPort: options.IPC.NatsPort,
	}))

	log.WithFields(log.Fields{
		"ipcMode":               ipcMode,
		"startMainSoroban":      startMainSoroban,
		"startIPCService":       startIPCService,
		"startP2PDirectory":     startP2PDirectory,
		"IPC.ChildID":           options.IPC.ChildID,
		"IPC.ChildProcessCount": options.IPC.ChildProcessCount,
	}).Debug("IPC info")

	if startIPCService {
		// start IPC directory
		log.Info("Start IPC Server")
		ready := make(chan struct{})

		go services.StartIPCService(ctx, ready)
		<-ready
		log.Info("Starting child process")

		// spawn p2p child process
		for i := 0; i < options.IPC.ChildProcessCount; i++ {
			startChildSoroban(ctx, options, i+1)
			<-time.After(5 * time.Second)
		}
	}

	// Both IPC Server & client need to connect to IPC server (bi-directionnal communcation)

	if client := internal.IPCFromContext(ctx); client != nil {
		client.Connect(ctx)
	}

	if startP2PDirectory {
		log.Warning("Start P2P Directory")

		if options.IPC.ChildID > 0 {
			log.Warning("IPC Client ListenFromServer requests")
			if client := internal.IPCFromContext(ctx); client != nil {
				go client.ListenFromServer(ctx, options.IPC.Subject, func(ctx context.Context, message ipc.Message) (ipc.Message, error) {
					switch message.Type {
					case ipc.MessageTypeIPC:
						log.Debug("IPC Message recieved from server")

						var p2pMessage p2p.Message

						err := unmarshalString(message.Payload, &p2pMessage)
						if err != nil {
							log.WithError(err).Error("Failed to Unmarshal IPC message")
							return ipc.Message{
								Type:    message.Type,
								Message: "error",
							}, nil
						}

						// forward message to p2p network
						p2P := internal.P2PFromContext(ctx)
						var args services.DirectoryEntry
						err = unmarshalData(p2pMessage.Payload, &args)
						if err != nil {
							log.WithError(err).Error("Failed to Unmarshal IPC message")
							return ipc.Message{
								Type:    message.Type,
								Message: "error",
							}, nil
						}

						log.WithField("p2pMessage", fmt.Sprintf("%s: %s", p2pMessage.Context, string(p2pMessage.Payload))).Debug("Publish Message to p2p")

						err = p2P.PublishJson(ctx, p2pMessage.Context, &args)
						if err != nil {
							log.WithError(err).Error("Failed to Publish P2P message")
							return ipc.Message{
								Type:    message.Type,
								Message: "error",
							}, nil
						}
						return ipc.Message{
							Type:    message.Type,
							Message: "success",
						}, nil
					default:
						return ipc.Message{
							Type:    message.Type,
							Message: "unknown",
						}, nil
					}

				})
			} else {
				log.Fatal("IPC Client not found in context")
			}
		}

		ready := make(chan struct{})
		go services.StartP2PDirectory(ctx, options.P2P.Seed, options.P2P.Bootstrap, options.P2P.Hostname, options.P2P.ListenPort, options.P2P.Room, ready)
		<-ready
		log.Info("P2PDirectory service started")
	}

	if !startMainSoroban {
		// soroban is in child mode
		return ctx, nil
	}

	// start soroban service
	var t *tor.Tor
	if options.Soroban.WithTor {
		var err error
		t, err = tor.Start(ctx, &tor.StartConf{
			DebugWriter:     io.Discard,
			TempDataDirBase: "/tmp",
		})
		if err != nil {
			log.WithError(err).Error("tor.Start error")
			return ctx, nil
		}
		t.DeleteDataDirOnClose = true
		soroban.AddTorClient(ctx, t)
		// wait for network ready
		log.Info("Waiting for soroban tor network")
		t.EnableNetwork(ctx, true)
	}

	rpcServer := rpc.NewServer()

	rpcServer.RegisterCodec(gjson.NewCodec(), "application/json")
	rpcServer.RegisterCodec(gjson.NewCodec(), "application/json;charset=UTF-8")

	http.Handle("/rpc", rpcServer)

	return ctx, &Soroban{
		p2p:       internal.P2PFromContext(ctx),
		ipc:       internal.IPCFromContext(ctx),
		t:         t,
		started:   make(chan bool),
		rpcServer: rpcServer,
		directory: directory,
	}
}

/// Soroban interface

func (p *Soroban) ID() string {
	if p.onion == nil {
		return ""
	}
	return p.onion.ID
}

// Register json-rpc service
func (p *Soroban) Register(ctx context.Context, name string, receiver soroban.Service) error {
	return p.rpcServer.RegisterService(receiver, name)
}

func (p *Soroban) Start(ctx context.Context, hostname string, port int) error {
	// start without listener
	go p.startServer(hostname, port, nil)

	return nil
}

func (p *Soroban) StartWithTor(ctx context.Context, hostname string, port int, seed string) error {
	if p.t == nil {
		return errors.New("tor not initialized")
	}
	var key crypto.PrivateKey
	if len(seed) > 0 {
		str, err := hex.DecodeString(seed)
		if err != nil {
			return err
		}
		key = ed25519.NewKeyFromSeed(str)
	}

	// Wait at most a few minutes to publish the tor hidden service
	listenCtx, listenCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer listenCancel()

	var err error
	p.onion, err = p.t.Listen(listenCtx,
		&tor.ListenConf{
			RemotePorts: []int{80},
			Key:         key,
		})
	if err != nil {
		return err
	}

	// start with listener
	go p.startServer(hostname, port, p.onion)

	return nil
}

func (p *Soroban) startServer(hostname string, port int, listener net.Listener) {
	p.started <- true
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Use your allowed origin here
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	stats := NewStats()
	rpcHandler := WrapHandler(stats.Middleware((p.rpcServer)))

	router := mux.NewRouter()
	router.HandleFunc("/rpc", rpcHandler)
	router.HandleFunc("/stats", stats.StatsHandler)
	router.HandleFunc("/status", StatusHandler)

	mainHandler := c.Handler(router)

	if listener != nil {
		go func() {
			torServer := p.createHttpServer("", mainHandler, TorListener)
			err := torServer.Serve(listener)
			if err != http.ErrServerClosed {
				log.WithError(err).Error("Tor Http Server exited")
			}
		}()
	}

	addr := fmt.Sprintf("%s:%d", hostname, port)
	ipv4Server := p.createHttpServer(addr, mainHandler, IPv4Listener)
	err := ipv4Server.ListenAndServe()
	if err != http.ErrServerClosed {
		log.WithError(err).Error("IPv4 Http Server exited")
	}
}

func (p *Soroban) Stop(ctx context.Context) {
	if p.onion == nil {
		return
	}
	err := p.onion.Close()
	if err != nil {
		log.WithError(err).Error("Fails to Close tor")
	}
	if p.t == nil {
		return
	}
	err = p.t.Close()
	if err != nil {
		log.WithError(err).Error("Fails to Close tor")
	}
}

func (p *Soroban) WaitForStart(ctx context.Context) {
	<-p.started
}

func (p *Soroban) createHttpServer(addr string, handler http.Handler, listenerType ListenerType) http.Server {

	return http.Server{
		Addr: addr,

		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			ctx = context.WithValue(ctx, internal.SorobanDirectoryKey, p.directory)
			if p.p2p != nil {
				ctx = context.WithValue(ctx, internal.SorobanP2PKey, p.p2p)
			}
			if p.ipc != nil {
				ctx = context.WithValue(ctx, internal.SorobanIPCKey, p.ipc)
			}

			return ctx
		},
		Handler: addListenerType(handler, listenerType),
	}
}
