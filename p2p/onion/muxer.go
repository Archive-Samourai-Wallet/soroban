package onion

import (
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/sec/insecure"
	"github.com/libp2p/go-libp2p/core/transport"
	msmux "github.com/libp2p/go-libp2p/p2p/muxer/muxer-multistream"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	csms "github.com/libp2p/go-libp2p/p2p/net/conn-security-multistream"
	tptu "github.com/libp2p/go-libp2p/p2p/net/upgrader"
)

func makeUpgrader(sk crypto.PrivKey, id peer.ID) (transport.Upgrader, error) {
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(insecure.ID, insecure.NewWithIdentity(id, sk))

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)
	u, err := tptu.New(secMuxer, stMuxer, tptu.WithAcceptTimeout(3*time.Minute))
	return u, err
}
