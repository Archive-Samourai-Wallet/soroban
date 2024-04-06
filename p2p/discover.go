package p2p

import (
	"context"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"

	log "github.com/sirupsen/logrus"
)

func Discover(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string, ready chan struct{}) {
	var routingDiscovery = routing.NewRoutingDiscovery(dht)

	err := advertize(ctx, routingDiscovery, rendezvous, 3)
	if err != nil {
		log.WithError(err).Panic("Advertise failed, giving up")
	}

	// advertize daemon
	go func() {
		for {
			select {
			case <-time.After(5 * time.Minute):
				err := advertize(ctx, routingDiscovery, rendezvous, 10)
				if err != nil {
					log.WithError(err).Error("Advertise failed, retrying")
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
				log.WithError(err).Error("failed to FindPeers")
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
						log.WithError(err).WithField("PeerID", p.ID).Warning("failed to connect to peer")
						continue
					}
					newPeersCount++
				}
			}
			if newPeersCount > 0 {
				log.WithField("Count", newPeersCount).Info("Connected to new peers")
			}
		}
	}
}

func advertize(ctx context.Context, routingDiscovery *routing.RoutingDiscovery, rendezvous string, retry int) error {
	var err error
	for i := 0; i < retry; i++ {
		_, err := routingDiscovery.Advertise(ctx, rendezvous, discovery.TTL(15*time.Minute))
		if err != nil {
			log.WithError(err).Error("failed to Advertise")

			// wait 1 minute to retry
			<-time.After(time.Minute)
			continue
		}

		log.Info("Advertise Complete")

		break
	}
	return err
}
