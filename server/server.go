package server

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"crypto"
	"crypto/ed25519"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal"

	"github.com/cretz/bine/tor"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

type Soroban struct {
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
		log.Fatalf("Invalid Directory")
	}

	var t *tor.Tor
	if options.WithTor {
		var err error
		t, err = tor.Start(ctx, &tor.StartConf{
			TempDataDirBase: "/tmp",
		})
		if err != nil {
			log.Printf("tor.Start error: %s", err)
			return nil
		}
		t.DeleteDataDirOnClose = true
	}

	rpcServer := rpc.NewServer()

	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	http.Handle("/rpc", rpcServer)

	return &Soroban{
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
func (p *Soroban) Register(name string, receiver soroban.Service) error {
	return p.rpcServer.RegisterService(receiver, name)
}

func (p *Soroban) Start(hostname string, port int) error {
	// start without listener
	go p.startServer(fmt.Sprintf("%s:%d", hostname, port), nil)

	return nil
}

func (p *Soroban) StartWithTor(port int, seed string) error {
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

	// Wait at most a few minutes to publish the service
	listenCtx, listenCancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	go p.startServer("", p.onion)

	return nil
}

func (p *Soroban) startServer(addr string, listener net.Listener) {
	p.started <- true
	router := mux.NewRouter()
	router.HandleFunc("/rpc", WrapHandler(p.rpcServer))
	router.HandleFunc("/status", StatusHandler)

	// Create http.Server with returning redis in context
	srv := http.Server{
		Addr: addr, // addr can be empty

		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, internal.SorobanDirectoryKey, p.directory)
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
		log.Fatalf("Http Server error")
	}
}

func (p *Soroban) Stop() {
	if p.onion == nil {
		return
	}
	err := p.onion.Close()
	if err != nil {
		log.Printf("Fails to Close tor")
	}
	if p.t == nil {
		return
	}
	err = p.t.Close()
	if err != nil {
		log.Printf("Fails to Close tor")
	}
}

func (p *Soroban) WaitForStart() {
	<-p.started
}
