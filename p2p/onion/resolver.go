package onion

import (
	"context"
	"net"
	"time"

	madns "github.com/multiformats/go-multiaddr-dns"
)

// NewTorResover returns a no madns.Resolver that will resolve
// IP addresses over Tor.
//
// TODO: This does not seem to work for TXT records. Look into if
// Tor can resolve TXT records.
func NewTorResover(proxy string) *madns.Resolver {
	netResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Minute,
			}
			return d.DialContext(ctx, network, proxy)
		},
	}
	r, _ := madns.NewResolver(madns.WithDefaultResolver(netResolver))
	return r
}
