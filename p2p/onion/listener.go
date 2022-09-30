package onion

import (
	"net"

	tpt "github.com/libp2p/go-libp2p/core/transport"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

// OnionListener implements go-libp2p-transport's Listener interface
type OnionListener struct {
	net.Conn
	laddr     ma.Multiaddr
	raddr     ma.Multiaddr
	listener  net.Listener
	transport tpt.Transport
	Upgrader  tpt.Upgrader
}

// Accept blocks until a connection is received returning
// go-libp2p-transport's Conn interface or an error if
// something went wrong
func (l *OnionListener) Accept() (manet.Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	raddr, err := manet.FromNetAddr(conn.RemoteAddr())
	if err != nil {
		return nil, err
	}
	onionConn := OnionConn{
		Conn:      conn,
		transport: l.transport,
		laddr:     l.laddr,
		raddr:     raddr,
	}

	return &onionConn, nil
}

func (l *OnionListener) LocalMultiaddr() ma.Multiaddr {
	return l.laddr
}

func (l *OnionListener) RemoteMultiaddr() ma.Multiaddr {
	return l.raddr
}

// Close shuts down the listener
func (l *OnionListener) Close() error {
	return l.listener.Close()
}

// Addr returns the net.Addr interface which represents
// the local multiaddr we are listening on
func (l *OnionListener) Addr() net.Addr {
	netaddr, _ := manet.ToNetAddr(l.laddr)
	return netaddr
}

// Multiaddr returns the local multiaddr we are listening on
func (l *OnionListener) Multiaddr() ma.Multiaddr {
	return l.laddr
}
