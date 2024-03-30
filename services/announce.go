package services

import (
	"context"
	"net/http"
	"time"

	"code.samourai.io/wallet/samourai-soroban/internal"
	log "github.com/sirupsen/logrus"
)

func StartAnnounce(ctx context.Context, announceKey string, nodeURLs ...string) {
	directory := internal.DirectoryFromContext(ctx)
	if directory == nil {
		log.Error("directory not found in context")
		return
	}

	p2P := internal.P2PFromContext(ctx)
	if p2P == nil {
		log.Error("P2P not found in context")
		return
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	dir := new(Directory)

	for {
		select {
		case <-ctx.Done():
			log.Info("Exiting Announce Loop")
			return
		case <-ticker.C:
			for _, nodeURL := range nodeURLs {
				log.WithField("nodeURL", nodeURL).Info("Announce")

				directoryEntry := DirectoryEntry{
					Name:  announceKey,
					Entry: nodeURL,
					Mode:  "short",
				}

				req, err := http.NewRequestWithContext(ctx, "POST", "", nil)
				if err != nil {
					log.WithError(err).Error("failed to create request")
					break
				}

				var resp Response
				err = dir.Add(req, &directoryEntry, &resp)
				if err != nil {
					log.WithError(err).Error("Failed to announce to directory")
				}
			}
		}
	}
}
