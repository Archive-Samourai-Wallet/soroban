package server

import (
	"errors"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/services"
)

func addToDirectory(directory soroban.Directory, args *services.DirectoryEntry) error {
	if args == nil {
		return errors.New("invalid args")
	}
	return directory.Add(args.Name, args.Entry, directory.TimeToLive(args.Mode))
}
