package onion

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cretz/bine/tor"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"golang.org/x/net/proxy"

	tpt "github.com/libp2p/go-libp2p/core/transport"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/whyrusleeping/mafmt"
)

// OnionTransport implements go-libp2p-transport's Transport interface
type OnionTransport struct {
	sk            crypto.PrivKey
	service       *tor.OnionService
	dialer        proxy.Dialer
	dialOnlyOnion bool
	laddr         ma.Multiaddr

	// Connection upgrader for upgrading insecure stream connections to
	// secure multiplex connections.
	Upgrader tpt.Upgrader
}

var _ tpt.Transport = &OnionTransport{}

// NewOnionTransport creates a new OnionTransport
func NewOnionTransport(sk crypto.PrivKey, dialer proxy.Dialer, service *tor.OnionService, dialOnlyOnion bool, upgrader tpt.Upgrader) (*OnionTransport, error) {
	o := OnionTransport{
		sk:            sk,
		dialer:        dialer,
		service:       service,
		dialOnlyOnion: dialOnlyOnion,
		Upgrader:      upgrader,
	}
	return &o, nil
}

// OnionTransportC is a type alias for OnionTransport constructors, for use
// with libp2p.New
type OnionTransportC func(upgrader tpt.Upgrader) (tpt.Transport, error)

// NewOnionTransportC is a convenience function that returns a function
// suitable for passing into libp2p.Transport for host configuration
func NewOnionTransportC(sk crypto.PrivKey, dialer proxy.Dialer, service *tor.OnionService, dialOnlyOnion bool) OnionTransportC {
	return func(upgrader tpt.Upgrader) (tpt.Transport, error) {
		return NewOnionTransport(sk, dialer, service, dialOnlyOnion, upgrader)
	}
}

// Dial dials a remote peer. It should try to reuse local listener
// addresses if possible but it may choose not to.
func (t *OnionTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (tpt.CapableConn, error) {
	netaddr, err := manet.ToNetAddr(raddr)
	var onionAddress string
	if err != nil {
		onionAddress, err = raddr.ValueForProtocol(ma.P_ONION3)
		if err != nil {
			onionAddress, err = raddr.ValueForProtocol(ma.P_ONION)
			if err != nil {
				return nil, err
			}
		}
	}
	onionConn := OnionConn{
		transport: tpt.Transport(t),
		laddr:     t.laddr,
		raddr:     raddr,
	}
	if len(onionAddress) > 0 {
		split := strings.Split(onionAddress, ":")
		onionConn.Conn, err = t.dialer.Dial("tcp4", split[0]+".onion:"+split[1])
	} else {
		onionConn.Conn, err = t.dialer.Dial(netaddr.Network(), netaddr.String())
	}
	if err != nil {
		return nil, err
	}

	u, err := makeUpgrader(t.sk, p)
	if err != nil {
		return nil, err
	}
	return u.Upgrade(ctx, onionConn.transport, &onionConn, network.DirOutbound, p, network.NullScope)
}

// Listen listens on the passed multiaddr.
func (t *OnionTransport) Listen(laddr ma.Multiaddr) (tpt.Listener, error) {
	// convert to net.Addr
	var (
		netaddr string
		err     error
	)
	netaddr, err = laddr.ValueForProtocol(ma.P_ONION3)
	if err != nil {
		netaddr, err = laddr.ValueForProtocol(ma.P_ONION)
		if err != nil {
			return nil, err
		}
	}

	// retreive onion service virtport
	addr := strings.Split(netaddr, ":")
	if len(addr) != 2 {
		return nil, fmt.Errorf("failed to parse onion address")
	}

	listener := OnionListener{
		laddr:     laddr,
		Upgrader:  t.Upgrader,
		transport: t,
	}

	if addr[0] != t.service.ID {
		return nil, errors.New("incorrect onion address")
	}

	listener.listener = t.service.LocalListener
	t.laddr = laddr

	ul := t.Upgrader.UpgradeListener(t, &listener)

	return ul, nil
}

// CanDial returns true if this transport knows how to dial the given
// multiaddr.
//
// Returning true does not guarantee that dialing this multiaddr will
// succeed. This function should *only* be used to preemptively filter
// out addresses that we can't dial.
func (t *OnionTransport) CanDial(a ma.Multiaddr) bool {
	if t.dialOnlyOnion {
		// only dial out on onion addresses
		return isValidOnionMultiAddr(a)
	} else {
		return isValidOnionMultiAddr(a) || mafmt.TCP.Matches(a)
	}
}

// Protocols returns the list of terminal protocols this transport can dial.
func (t *OnionTransport) Protocols() []int {
	if t.dialOnlyOnion {
		return []int{ma.P_ONION, ma.P_ONION3}
	} else {
		return []int{ma.P_ONION, ma.P_ONION3, ma.P_TCP}
	}
}

// Proxy always returns false for the onion transport.
func (t *OnionTransport) Proxy() bool {
	return false
}
