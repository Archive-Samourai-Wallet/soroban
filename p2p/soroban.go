package p2p

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

// P2P for distributed soroban
type P2P struct {
	topic     *pubsub.Topic
	OnMessage chan Message
}

func (t *P2P) Start(ctx context.Context, listenPort int, bootstrap, room string) error {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		return err
	}

	gossipSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return err
	}

	addrs := []multiaddr.Multiaddr{}
	addr, err := multiaddr.NewMultiaddr(bootstrap)
	if err != nil {
		return err
	}
	addrs = append(addrs, addr)
	dht, err := NewDHT(ctx, host, addrs...)

	go Discover(ctx, host, dht, room)

	topic, err := gossipSub.Join(room)
	if err != nil {
		return err
	}

	t.topic = topic

	// subscribe to topic
	subscriber, err := topic.Subscribe()
	if err != nil {
		return err
	}
	go t.subscribe(ctx, subscriber, host.ID())

	return nil
}

// start subsriber to topic
func (t *P2P) subscribe(ctx context.Context, subscriber *pubsub.Subscription, hostID peer.ID) {
	for {
		msg, err := subscriber.Next(ctx)
		if err != nil {
			log.Printf("failed to get next message")
			<-time.After(time.Second)
			continue
		}

		// only consider messages delivered by other peers
		if msg.ReceivedFrom == hostID {
			continue
		}

		message, err := MessageFromBytes(msg.Data)
		if err != nil {
			log.Printf("failed to get MessageFromBytes")
			<-time.After(time.Second)
			continue
		}

		t.OnMessage <- message
	}
}

// Publish to topic
func (t *P2P) Publish(ctx context.Context, msg string) error {
	if len(msg) == 0 {
		return errors.New("failed to publish empty message")
	}
	t.topic.Publish(ctx, []byte(msg))
	return nil
}

// Publish to topic
func (t *P2P) PublishJson(ctx context.Context, context string, payload interface{}) error {
	message, err := NewMessage(context, payload)
	if err != nil {
		return err
	}

	data, err := message.ToBytes()
	if err != nil {
		return err
	}

	return t.Publish(ctx, string(data))
}
