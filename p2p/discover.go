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

	err := advertize(ctx, routingDiscovery, rendezvous, 3)
	if err != nil {
		log.Panic("Advertise failed, giving up")
	}

	// advertize daemon
	go func() {
		for {
			select {
			case <-time.After(5 * time.Minute):
				err := advertize(ctx, routingDiscovery, rendezvous, 10)
				if err != nil {
					log.Panic("Advertise failed, giving up")
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	ready <- struct{}{}

	ticker := time.NewTicker(45 * time.Second)
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

			newPeersCount := 0
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
					newPeersCount++
				}
			}
			if newPeersCount > 0 {
				log.Printf("Connected to new peers %d", newPeersCount)
			}
		}
	}
}

func advertize(ctx context.Context, routingDiscovery *routing.RoutingDiscovery, rendezvous string, retry int) error {
	var err error
	for i := 0; i < retry; i++ {
		_, err := routingDiscovery.Advertise(ctx, rendezvous, discovery.TTL(15*time.Minute))
		if err != nil {
			log.Printf("failed to Advertise. %s", err)

			// wait 1 minute to retry
			<-time.After(time.Minute)
			continue
		}

		log.Printf("Advertise Complete")

		break
	}
	return err
}
