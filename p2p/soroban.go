package p2p

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"

	log "github.com/sirupsen/logrus"
)

// P2P for distributed soroban
type P2P struct {
	topic     *pubsub.Topic
	OnMessage chan Message
}

func (p *P2P) Valid() bool {
	return p.topic != nil
}

func (p *P2P) Start(ctx context.Context, listenPort int, bootstrap, room string) error {

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
	if err != nil {
		return err
	}

	go Discover(ctx, host, dht, room)

	topic, err := gossipSub.Join(room)
	if err != nil {
		return err
	}

	p.topic = topic

	// subscribe to topic
	subscriber, err := topic.Subscribe()
	if err != nil {
		return err
	}
	go p.subscribe(ctx, subscriber, host.ID())

	return nil
}

// start subsriber to topic
func (p *P2P) subscribe(ctx context.Context, subscriber *pubsub.Subscription, hostID peer.ID) {
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
			log.Debug("Skip unkown message")
			continue
		}

		p.OnMessage <- message
	}
}

// Publish to topic
func (p *P2P) Publish(ctx context.Context, msg string) error {
	if len(msg) == 0 {
		return errors.New("failed to publish empty message")
	}
	if p.topic == nil {
		return nil
	}
	p.topic.Publish(ctx, []byte(msg))
	return nil
}

// Publish to topic
func (p *P2P) PublishJson(ctx context.Context, context string, payload interface{}) error {
	message, err := NewMessage(context, payload)
	if err != nil {
		return err
	}

	data, err := message.ToBytes()
	if err != nil {
		return err
	}

	return p.Publish(ctx, string(data))
}
