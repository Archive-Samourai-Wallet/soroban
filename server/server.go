package server

import (
	"context"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"time"

	"crypto"
	"crypto/ed25519"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal"

	"github.com/cretz/bine/tor"
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
	case "redis":
		directory = internal.NewDirectory(options.Domain, internal.DirectoryTypeRedis, options.Directory)
	case "default":
		directory = internal.DefaultDirectory(options.Domain, options.Directory)
	}
	if directory == nil {
		log.Fatalf("Invalid Directory")
	}

	t, err := tor.Start(ctx, &tor.StartConf{
		TempDataDirBase: "/tmp",
	})
	if err != nil {
		log.Printf("tor.Start error: %s", err)
		return nil
	}
	t.DeleteDataDirOnClose = true

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
	return p.onion.ID
}

// Register json-rpc service
func (p *Soroban) Register(name string, receiver soroban.Service) error {
	return p.rpcServer.RegisterService(receiver, name)
}

func (p *Soroban) Start(seed string) error {
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
			LocalPort:   4242,
			RemotePorts: []int{80},
			Key:         key,
		})
	if err != nil {
		return err
	}

	go func() {
		p.started <- true
		// Create http.Server with returning redis in context
		srv := http.Server{
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				return context.WithValue(ctx, internal.SorobanDirectoryKey, p.directory)
			},
			Handler: WrapHandler(p.rpcServer),
		}
		err := srv.Serve(p.onion)
		if err != http.ErrServerClosed {
			log.Fatalf("Http Server error")
		}
	}()

	return nil
}

func (p *Soroban) Stop() {
	err := p.onion.Close()
	if err != nil {
		log.Printf("Fails to Close tor")
	}
	err = p.t.Close()
	if err != nil {
		log.Printf("Fails to Close tor")
	}
}

func (p *Soroban) WaitForStart() {
	<-p.started
}
