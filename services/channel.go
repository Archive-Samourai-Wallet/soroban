package services

import (
	"fmt"
	"net/http"
)

// Channel struct for json-rpc
type Channel struct{}

// ChannelArgs for json-rpc request
type ChannelArgs struct {
	Identifier string
	Name       string
}

func (t *Channel) Join(r *http.Request, args *ChannelArgs, result *Response) error {
	*result = Response{Result: fmt.Sprintf("%s joined channel %s", args.Identifier, args.Name)}
	return nil
}
