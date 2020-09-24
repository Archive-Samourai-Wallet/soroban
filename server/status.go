package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	soroban "code.samourai.io/wallet/samourai-soroban"
	"code.samourai.io/wallet/samourai-soroban/internal"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	directory := internal.DirectoryFromContext(r.Context())

	fullStatus, err := directory.Status()
	if err != nil {
		http.Error(w, "Directory status error", http.StatusInternalServerError)
	}

	// remove private informations
	status := soroban.StatusInfo{
		CPU:      fullStatus.CPU,
		Clients:  fullStatus.Clients,
		Keyspace: fullStatus.Keyspace,
		Memory:   fullStatus.Memory,
		Stats:    fullStatus.Stats,
	}

	// filter informations
	filtersQuery := r.URL.Query().Get("filters")
	if len(filtersQuery) == 0 {
		filtersQuery = "default"
	}

	var result soroban.StatusInfo
	filters := strings.Split(filtersQuery, ",")
	if len(filters) == 0 {
		filters = []string{"default"}
	}

	for _, filter := range filters {
		switch filter {
		case "default":
			result = soroban.StatusInfo{
				CPU:      status.CPU,
				Clients:  status.Clients,
				Keyspace: status.Keyspace,
			}

		case "cpu":
			result.CPU = status.CPU
		case "clients":
			result.Clients = status.Clients
		case "keyspace":
			result.Keyspace = status.Keyspace

		case "memory":
			result.Memory = status.Memory
		case "stats":
			result.Stats = status.Stats

		case "*":
			result = status

		case "debug_all":
			result = fullStatus
		}
	}

	data, err := json.Marshal(&result)
	if err != nil {
		http.Error(w, "Directory status error", http.StatusInternalServerError)
	}

	// prepare response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, string(data))
}
