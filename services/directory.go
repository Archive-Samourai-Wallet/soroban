package services

import (
	"net/http"

	"code.samourai.io/wallet/samourai-soroban/internal"

	log "github.com/sirupsen/logrus"
)

// DirectoryEntries for json-rpc request
type DirectoryEntries struct {
	Name    string
	Entries []string
}

// DirectoryEntry for json-rpc request
type DirectoryEntry struct {
	Name  string
	Entry string
	Mode  string
}

// Directory struct for json-rpc
type Directory struct{}

func (t *Directory) List(r *http.Request, args *DirectoryEntries, result *DirectoryEntries) error {
	directory := internal.DirectoryFromContext(r.Context())
	if directory == nil {
		log.Error("Directory not found")
		return nil
	}

	entries, err := directory.List(args.Name)
	if err != nil {
		log.WithError(err).Error("Failed to list directory")
	}
	log.Debugf("List: %s (%d)", args.Name, len(entries))

	if entries == nil {
		entries = make([]string, 0)
	}
	*result = DirectoryEntries{
		Name:    args.Name,
		Entries: entries,
	}
	return nil
}

func (t *Directory) Add(r *http.Request, args *DirectoryEntry, result *Response) error {
	directory := internal.DirectoryFromContext(r.Context())
	if directory == nil {
		log.Error("Directory not found")
		return nil
	}

	log.Debugf("Add: %s %s", args.Name, args.Entry)

	status := "success"
	err := directory.Add(args.Name, args.Entry, directory.TimeToLive(args.Mode))
	if err != nil {
		status = "error"
		log.WithError(err).Error("Failed to Add entry")
	}

	*result = Response{
		Status: status,
	}
	return nil
}

func (t *Directory) Remove(r *http.Request, args *DirectoryEntry, result *Response) error {
	directory := internal.DirectoryFromContext(r.Context())
	if directory == nil {
		log.Error("Directory not found")
		return nil
	}

	log.Debugf("Remove: %s %s", args.Name, args.Entry)

	status := "success"
	err := directory.Remove(args.Name, args.Entry)
	if err != nil {
		status = "error"
		log.WithError(err).Error("Failed to Remove directory")
	}

	*result = Response{
		Status: status,
	}
	return nil
}
