package server

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"crypto"
	"crypto/ed25519"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal"
	"code.samourai.io/wallet/samourai-soroban/p2p"
	"code.samourai.io/wallet/samourai-soroban/services"

	"github.com/cretz/bine/tor"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	log "github.com/sirupsen/logrus"
)

type Soroban struct {
	p2p       *p2p.P2P
	directory soroban.Directory
	t         *tor.Tor
	onion     *tor.OnionService
	started   chan bool
	rpcServer *rpc.Server
}

func New(ctx context.Context, options soroban.Options) *Soroban {
	var directory soroban.Directory

	switch options.DirectoryType {
	case "memory":
		directory = internal.NewDirectory(options.Domain, internal.DirectoryTypeRedis, options.Directory)
	case "redis":
		directory = internal.NewDirectory(options.Domain, internal.DirectoryTypeRedis, options.Directory)
	case "default":
		directory = internal.DefaultDirectory(options.Domain, options.Directory)
	}
	if directory == nil {
		log.Fatal("Invalid Directory")
	}

	var t *tor.Tor
	if options.WithTor {
		var err error
		t, err = tor.Start(ctx, &tor.StartConf{
			TempDataDirBase: "/tmp",
		})
		if err != nil {
			log.WithError(err).Error("tor.Start error")
			return nil
		}
		t.DeleteDataDirOnClose = true
	}

	rpcServer := rpc.NewServer()

	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	http.Handle("/rpc", rpcServer)

	// StartDirectory p2p messages service and directory
	p2P := &p2p.P2P{OnMessage: make(chan p2p.Message)}

	if len(options.P2P.Bootstrap) > 0 {
		ctx = context.WithValue(ctx, internal.SorobanDirectoryKey, directory)
		ctx = context.WithValue(ctx, internal.SorobanP2PKey, p2P)

		go services.StartP2PDirectory(ctx, options.P2P.Bootstrap, options.P2P.Room)
	}

	return &Soroban{
		p2p:       p2P,
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
	go p.startServer(ctx, fmt.Sprintf("%s:%d", hostname, port), nil)

	return nil
}

func (p *Soroban) StartWithTor(ctx context.Context, port int, seed string) error {
	if p.t == nil {
		return errors.New("Tor not initialized")
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
			LocalPort:   port,
			RemotePorts: []int{80},
			Key:         key,
		})
	if err != nil {
		return err
	}

	// start with listener
	go p.startServer(ctx, "", p.onion)

	return nil
}

func (p *Soroban) startServer(ctx context.Context, addr string, listener net.Listener) {
	p.started <- true
	router := mux.NewRouter()
	router.HandleFunc("/rpc", WrapHandler(p.rpcServer))
	router.HandleFunc("/status", StatusHandler)

	// Create http.Server with returning redis in context
	srv := http.Server{
		Addr: addr, // addr can be empty

		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			ctx = context.WithValue(ctx, internal.SorobanDirectoryKey, p.directory)
			ctx = context.WithValue(ctx, internal.SorobanP2PKey, p.p2p)

			return ctx
		},
		Handler: router,
	}

	var err error
	if listener != nil {
		err = srv.Serve(listener) // use specified http listener
	} else {
		err = srv.ListenAndServe()
	}
	if err != http.ErrServerClosed {
		log.Fatal("Http Server error")
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
