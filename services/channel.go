package services

import (
	"fmt"
	"net/http"

	soroban "code.samourai.io/wallet/samourai-soroban"
)

// Channel struct for json-rpc
type Channel struct{}

// ChannelArgs for json-rpc request
type ChannelArgs struct {
	Identifier string
	Name       string
}

func (t *Channel) Join(r *http.Request, args *ChannelArgs, result *Response) error {
	redis := soroban.RedisFromContext(r.Context())
	if redis == nil {
		fmt.Println("Redis not found")
		return nil
	}
	*result = Response{Result: fmt.Sprintf("%s joined channel %s", args.Identifier, args.Name)}
	return nil
}
