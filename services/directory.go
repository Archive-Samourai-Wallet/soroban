package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/confidential"
	"code.samourai.io/wallet/samourai-soroban/internal"
	"code.samourai.io/wallet/samourai-soroban/ipc"
	"code.samourai.io/wallet/samourai-soroban/p2p"

	log "github.com/sirupsen/logrus"
)

// DirectoryEntries for json-rpc request
type DirectoryEntries struct {
	Name      string
	Limit     int
	PublicKey string
	Algorithm string
	Signature string
	Timestamp int64
}

// DirectoryEntriesResponse for json-rpc response
type DirectoryEntriesResponse struct {
	Name    string
	Entries []string
}

// DirectoryEntry for json-rpc request
type DirectoryEntry struct {
	Name      string
	Entry     string
	Mode      string
	PublicKey string
	Algorithm string
	Signature string
	Timestamp int64
}

// Directory struct for json-rpc
type Directory struct{}

func (t *Directory) List(r *http.Request, args *DirectoryEntries, result *DirectoryEntriesResponse) error {
	directory := internal.DirectoryFromContext(r.Context())
	if directory == nil {
		log.Error("Directory not found")
		return nil
	}

	info := confidential.GetConfidentialInfo(args.Name, args.PublicKey)
	// check signature if key is confidential, list is not allowed for anonymous
	if info.Confidential {
		err := args.VerifySignature(info)
		if err != nil {
			log.WithError(err).Error("Failed to verifySignature")
			return nil
		}
	}

	entries, err := directory.List(args.Name)
	if err != nil {
		log.WithError(err).Error("Failed to list directory")
		return nil
	}

	if args.Limit > 0 && args.Limit < len(entries) {
		rand.Shuffle(len(entries), func(i, j int) {
			entries[i], entries[j] = entries[j], entries[i]
		})
		entries = entries[:args.Limit]
	}

	log.Tracef("List: %s (%d)", args.Name, len(entries))

	if entries == nil {
		entries = make([]string, 0)
	}
	*result = DirectoryEntriesResponse{
		Name:    args.Name,
		Entries: entries,
	}
	return nil
}

func addToDirectory(directory soroban.Directory, args *DirectoryEntry) error {
	if args == nil {
		return errors.New("invalid args")
	}
	return directory.Add(args.Name, args.Entry, directory.TimeToLive(args.Mode))
}

func (t *Directory) Add(r *http.Request, args *DirectoryEntry, result *Response) error {
	ctx := r.Context()
	directory := internal.DirectoryFromContext(ctx)
	if directory == nil {
		log.Error("Directory not found")
		return nil
	}

	info := confidential.GetConfidentialInfo(args.Name, args.PublicKey)
	// check signature if key is readonly, add is not allowed for anonymous
	if info.ReadOnly {
		err := args.VerifySignature(info)
		if err != nil {
			log.WithError(err).Error("Failed to verifySignature")
			*result = Response{
				Status: "error",
			}
			return nil
		}
	}

	log.Debugf("Add: %s %s", args.Name, args.Entry)

	err := addToDirectory(directory, args)
	if err != nil {
		log.WithError(err).Error("Failed to Add entry")
		*result = Response{
			Status: "error",
		}
		return nil
	}

	if client := internal.IPCFromContext(ctx); client != nil {
		log.Debug("Forward Message message to IPC client")
		message, err := p2p.NewMessage("Directory.Add", &args)
		if err != nil {
			log.WithError(err).Error("failed to marshal p2P message.")
			*result = Response{
				Status: "error",
			}
			return nil
		}

		data, err := json.Marshal(message)
		if err != nil {
			log.WithError(err).Error("failed to marshal p2p message")
			*result = Response{
				Status: "error",
			}
			return nil
		}
		resp, err := client.Request(ipc.Message{
			Type:    ipc.MessageTypeIPC,
			Payload: string(data),
		}, "down")
		if err != nil {
			log.WithError(err).Error("IPC requext failed")
			*result = Response{
				Status: "error",
			}
			return nil
		}
		if resp.Message != "success" {
			log.WithField("Message", resp.Message).Warning("IPC Message failed")
		}
		log.WithField("Message", resp.Message).Debug("IPC Message sent")
	}

	if p2P := internal.P2PFromContext(ctx); p2P != nil {

		err := p2P.PublishJson(ctx, "Directory.Add", args)
		if err != nil {
			// non fatal error
			log.Printf("p2P - Failed to PublishJson. %s\n", err)
		}

		*result = Response{
			Status: "success",
		}
	} else {
		log.Println("p2P - P2P not found")
		return nil
	}

	return nil
}

func removeFromDirectory(directory soroban.Directory, args *DirectoryEntry) error {
	if args == nil {
		return errors.New("invalid args")
	}
	return directory.Remove(args.Name, args.Entry)
}

func (t *Directory) Remove(r *http.Request, args *DirectoryEntry, result *Response) error {
	ctx := r.Context()
	directory := internal.DirectoryFromContext(ctx)
	if directory == nil {
		log.Error("Directory not found")
		return nil
	}

	info := confidential.GetConfidentialInfo(args.Name, args.PublicKey)
	// check signature if key is readonly, remove is not allowed for anonymous
	if info.ReadOnly {
		err := args.VerifySignature(info)
		if err != nil {
			log.WithError(err).Error("Failed to verifySignature")
			return nil
		}
	}

	p2P := internal.P2PFromContext(ctx)
	if p2P == nil {
		log.Println("p2P - P2P not found")
		return nil
	}

	log.Debugf("Remove: %s %s", args.Name, args.Entry)

	status := "success"
	err := removeFromDirectory(directory, args)
	if err != nil {
		status = "error"
		log.WithError(err).Error("Failed to Remove directory")
	}

	err = p2P.PublishJson(ctx, "Directory.Remove", args)
	if err != nil {
		// non fatal error
		log.Printf("p2P - Failed to PublishJson. %s\n", err)
	}

	*result = Response{
		Status: status,
	}
	return nil
}

func timeInRange(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func (p *DirectoryEntries) VerifySignature(info confidential.ConfidentialEntry) error {
	if len(info.Prefix) == 0 || len(info.Algorithm) == 0 || len(info.PublicKey) == 0 {
		return nil
	}

	now := time.Now().UTC()
	timestamp := time.Unix(0, p.Timestamp).UTC()
	log.WithField("Timestamp", timestamp).Warning("VerifySignature")
	delta := 24 * time.Hour

	if p.PublicKey != info.PublicKey {
		return errors.New("PublicKey not allowed")
	}

	if !timeInRange(now.Add(-delta), now.Add(delta), timestamp) {
		return errors.New("timestamp not in time range")
	}

	message := fmt.Sprintf("%v.%v", p.Name, p.Timestamp)
	return confidential.VerifySignature(info, p.PublicKey, message, p.Algorithm, p.Signature)
}

func (p *DirectoryEntry) VerifySignature(info confidential.ConfidentialEntry) error {
	if len(info.Prefix) == 0 || len(info.Algorithm) == 0 || len(info.PublicKey) == 0 {
		return nil
	}

	if p.PublicKey != info.PublicKey {
		return errors.New("PublicKey not allowed")
	}

	now := time.Now().UTC()
	timestamp := time.Unix(0, p.Timestamp).UTC()
	delta := 24 * time.Hour
	if !timeInRange(now.Add(-delta), now.Add(delta), timestamp) {
		return errors.New("timestamp not in time range")
	}
	message := fmt.Sprintf("%s.%d.%s", p.Name, p.Timestamp, p.Entry)
	return confidential.VerifySignature(info, p.PublicKey, message, p.Algorithm, p.Signature)
}
