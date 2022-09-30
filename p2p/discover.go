package p2p

import (
	"context"
	"log"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

func Discover(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string, ready chan struct{}) {
	var routingDiscovery = routing.NewRoutingDiscovery(dht)

	_, err := routingDiscovery.Advertise(ctx, rendezvous, discovery.TTL(15*time.Minute))
	if err != nil {
		log.Printf("failed to Advertise. %s", err)
		return
	}

	ready <- struct{}{}

	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			peers, err := routingDiscovery.FindPeers(ctx, rendezvous)
			if err != nil {
				log.Printf("failed to FindPeers. %s", err)
				continue
			}

			for p := range peers {
				if p.ID == h.ID() {
					continue
				}
				if h.Network().Connectedness(p.ID) != network.Connected {
					_, err = h.Network().DialPeer(ctx, p.ID)
					// fmt.Printf("Connected to peer %s\n", p.ID.Pretty())
					if err != nil {
						continue
					}
				}
			}
		}
	}
}
