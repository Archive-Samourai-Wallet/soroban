package services

import (
	"log"
	"net/http"

	"code.samourai.io/wallet/samourai-soroban/internal"
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
		log.Println("Directory not found")
		return nil
	}

	entries, err := directory.List(args.Name)
	if err != nil {
		log.Printf("Failed to list directory. %s", err)
	}
	log.Printf("List: %s (%d)", args.Name, len(entries))

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
		log.Println("Directory not found")
		return nil
	}

	log.Printf("Add: %s %s", args.Name, args.Entry)

	status := "success"
	err := directory.Add(args.Name, args.Entry, directory.TimeToLive(args.Mode))
	if err != nil {
		status = "error"
		log.Printf("Failed to Add entry. %s", err)
	}

	*result = Response{
		Status: status,
	}
	return nil
}

func (t *Directory) Remove(r *http.Request, args *DirectoryEntry, result *Response) error {
	directory := internal.DirectoryFromContext(r.Context())
	if directory == nil {
		log.Println("Directory not found")
		return nil
	}

	log.Printf("Remove: %s %s", args.Name, args.Entry)

	status := "success"
	err := directory.Remove(args.Name, args.Entry)
	if err != nil {
		status = "error"
		log.Printf("Failed to Remove directory. %s", err)
	}

	*result = Response{
		Status: status,
	}
	return nil
}
