package soroban

import (
	"context"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"crypto"
	"crypto/ed25519"

	"github.com/cretz/bine/tor"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

type Soroban struct {
	t         *tor.Tor
	onion     *tor.OnionService
	Ready     chan bool
	rpcServer *rpc.Server
}

func New() *Soroban {
	t, err := tor.Start(nil, nil)
	if err != nil {
		return nil
	}
	t.DeleteDataDirOnClose = true

	rpcServer := rpc.NewServer()

	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
	rpcServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")

	http.Handle("/rpc", rpcServer)

	return &Soroban{
		t:         t,
		Ready:     make(chan bool),
		rpcServer: rpcServer,
	}
}

func (p *Soroban) ID() string {
	return p.onion.ID
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
			LocalPort:   8080,
			RemotePorts: []int{80},
			Key:         key,
		})
	if err != nil {
		return err
	}

	go func() {
		p.Ready <- true
		err := http.Serve(p.onion, nil)
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

func (p *Soroban) Register(receiver interface{}, name string) error {
	return p.rpcServer.RegisterService(receiver, name)
}
