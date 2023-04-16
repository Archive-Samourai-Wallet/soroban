package p2p

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/p2p/onion"

	"github.com/cretz/bine/tor"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"

	log "github.com/sirupsen/logrus"
)

func initTorP2P(ctx context.Context, p2pSeed string, listenPort int) ([]libp2p.Option, error) {
	extraArgs := []string{
		// "--DNSPort", "2121",
	}

	var privateKey ed25519.PrivateKey
	var priv crypto.PrivKey
	if len(p2pSeed) > 0 {
		if p2pSeed == "auto" {
			_, pri, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return nil, err
			}

			p2pSeed = hex.EncodeToString(pri.Seed())
		}
		data, err := hex.DecodeString(p2pSeed)
		if err != nil {
			return nil, err
		}
		priv, err = crypto.UnmarshalSecp256k1PrivateKey(data)
		if err != nil {
			return nil, err
		}
		privateKey = ed25519.NewKeyFromSeed(data)
	}

	// Create the embedded Tor client.
	torClient, err := tor.Start(ctx, &tor.StartConf{
		DebugWriter:     io.Discard,
		TempDataDirBase: "/tmp",
		ExtraArgs:       extraArgs,
	})
	if err != nil {
		log.WithError(err).Error("Failed to tor.Start")
		return nil, err
	}
	torClient.DeleteDataDirOnClose = true
	// wait for network ready
	log.Info("Waiting for p2p tor network")
	soroban.AddTorClient(ctx, torClient)
	torClient.EnableNetwork(ctx, true)

	// Create the onion service.
	onionService, err := torClient.Listen(ctx, &tor.ListenConf{
		RemotePorts: []int{listenPort},
		LocalPort:   listenPort,
		Version3:    true,
		Key:         privateKey,
	})
	if err != nil {
		log.WithError(err).Error("Failed to torClient.Listen")
		return nil, err
	}

	// Override the default lip2p DNS resolver. We need this because libp2p address may contain a
	// DNS hostname that will be resolved before dialing. If we do not configure the resolver to
	// use Tor we will blow any anonymity we gained by using Tor.
	//
	// Note you must enter the DNS resolver address that was used when creating the Tor client.
	resolver := madns.DefaultResolver // Noop
	madns.DefaultResolver = resolver  //onion.NewTorResover("localhost:2121")

	dialOnlyOnion := true

	// Create the libp2p transport option.
	// Create address option.
	onionAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/onion3/%s:%d", onionService.ID, listenPort))
	if err != nil {
		log.WithError(err).Error("Failed to NewMultiaddr onion3")
		return nil, err
	}

	// Create the dialer.
	//
	// IMPORTANT: If you are genuinely trying to anonymize your IP you will need to route
	// any non-libp2p traffic through this dialer as well. For example, any HTTP requests
	// you make MUST go through this dialer.
	dialer, err := torClient.Dialer(ctx, nil)
	if err != nil {
		log.WithError(err).Error("Failed to torClient.Dialer")
		return nil, err
	}

	return []libp2p.Option{
		libp2p.Identity(priv),

		libp2p.ListenAddrs(onionAddr),
		libp2p.Transport(onion.NewOnionTransportC(priv, dialer, onionService, dialOnlyOnion)),
		libp2p.DefaultMuxers,
		libp2p.DefaultPeerstore,

		libp2p.NoSecurity,
		libp2p.WithDialTimeout(5 * time.Minute),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(),
	}, nil
}
