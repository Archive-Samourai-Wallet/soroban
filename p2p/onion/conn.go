package onion

import (
	"encoding/base32"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"

	tpt "github.com/libp2p/go-libp2p/core/transport"
	ma "github.com/multiformats/go-multiaddr"
)

// OnionConn implement's go-libp2p-transport's Conn interface
type OnionConn struct {
	net.Conn
	transport tpt.Transport
	laddr     ma.Multiaddr
	raddr     ma.Multiaddr
}

func (c *OnionConn) ReadByte() (byte, error) {
	r, ok := c.Conn.(io.Reader)
	if !ok {
		return 0, errors.New("invalid byte reader")
	}
	var b [1]byte
	_, err := r.Read(b[:])
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// Transport returns the OnionTransport associated
// with this OnionConn
func (c *OnionConn) Transport() tpt.Transport {
	return c.transport
}

// LocalMultiaddr returns the local multiaddr for this connection
func (c *OnionConn) LocalMultiaddr() ma.Multiaddr {
	return c.laddr
}

// RemoteMultiaddr returns the remote multiaddr for this connection
func (c *OnionConn) RemoteMultiaddr() ma.Multiaddr {
	return c.raddr
}

// isValidOnionMultiAddr is used to validate that a multiaddr
// is representing a Tor onion service
func isValidOnionMultiAddr(a ma.Multiaddr) bool {
	if len(a.Protocols()) != 1 {
		return false
	}

	// check for correct network type
	if a.Protocols()[0].Name != "onion3" {
		return false
	}

	// split into onion address and port
	var (
		addr string
		err  error
	)
	addr, err = a.ValueForProtocol(ma.P_ONION3)
	if err != nil {
		addr, err = a.ValueForProtocol(ma.P_ONION)
		if err != nil {
			return false
		}
	}
	split := strings.Split(addr, ":")
	if len(split) != 2 {
		return false
	}

	_, err = base32.StdEncoding.DecodeString(strings.ToUpper(split[0]))
	if err != nil {
		return false
	}

	// onion port number
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return false
	}
	if port >= 65536 || port < 1024 {
		return false
	}

	return true
}
