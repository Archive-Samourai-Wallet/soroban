package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type ContextKey string
type ListenerType string

const (
	ListenerTypeKey = ContextKey("soroban-listener")

	IPv4Listener ListenerType = "IPv4"
	TorListener  ListenerType = "Tor"
)

type Stats struct {
	sync.RWMutex
	IPv4 []int64
	Tor  []int64
}

func NewStats() *Stats {
	return &Stats{
		IPv4: make([]int64, 0),
		Tor:  make([]int64, 0),
	}
}

func (s *Stats) RecordRequest(listenerType ListenerType) {
	s.Lock()
	defer s.Unlock()
	now := time.Now().Unix()

	switch listenerType {
	case IPv4Listener:
		s.IPv4 = append(s.IPv4, now)

	case TorListener:
		s.Tor = append(s.Tor, now)
	}
}

func (s *Stats) CountRequests(listenerType ListenerType, duration time.Duration) int {
	s.RLock()
	defer s.RUnlock()
	now := time.Now().Unix()
	threshold := now - int64(duration.Seconds())
	count := 0

	requests := []int64{}
	switch listenerType {
	case IPv4Listener:
		requests = s.IPv4

	case TorListener:
		requests = s.Tor
	}

	for _, v := range requests {
		if v >= threshold {
			count++
		}
	}
	return count
}

func (s *Stats) Cleanup(duration time.Duration) {
	s.Lock()
	defer s.Unlock()
	startDate := time.Now().Add(-1 * duration).Unix()
	thresholdIndex := 0

	// Remove older entries
	for i, v := range s.IPv4 {
		if v > startDate {
			thresholdIndex = i
			break
		}
	}
	s.IPv4 = s.IPv4[thresholdIndex:]

	for i, v := range s.Tor {
		if v > startDate {
			thresholdIndex = i
			break
		}
	}
	s.Tor = s.Tor[thresholdIndex:]
}

func (s *Stats) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listenerType := r.Context().Value(ListenerTypeKey).(ListenerType)
		s.RecordRequest(listenerType)
		next.ServeHTTP(w, r)
	})
}

func (s *Stats) StatsHandler(w http.ResponseWriter, r *http.Request) {
	s.Cleanup(24 * time.Hour)

	ipv4 := map[string]int{
		"last_01m": s.CountRequests(IPv4Listener, time.Minute),
		"last_15m": s.CountRequests(IPv4Listener, 15*time.Minute),
		"last_30m": s.CountRequests(IPv4Listener, 30*time.Minute),
		"last_1h":  s.CountRequests(IPv4Listener, time.Hour),
		"last_2h":  s.CountRequests(IPv4Listener, 2*time.Hour),
		"last_3h":  s.CountRequests(IPv4Listener, 3*time.Hour),
		"last_6h":  s.CountRequests(IPv4Listener, 6*time.Hour),
		"last_12h": s.CountRequests(IPv4Listener, 12*time.Hour),
		"last_24h": s.CountRequests(IPv4Listener, 24*time.Hour),
	}

	tor := map[string]int{
		"last_01m": s.CountRequests(TorListener, time.Minute),
		"last_15m": s.CountRequests(TorListener, 15*time.Minute),
		"last_30m": s.CountRequests(TorListener, 30*time.Minute),
		"last_1h":  s.CountRequests(TorListener, time.Hour),
		"last_2h":  s.CountRequests(TorListener, 2*time.Hour),
		"last_3h":  s.CountRequests(TorListener, 3*time.Hour),
		"last_6h":  s.CountRequests(TorListener, 6*time.Hour),
		"last_12h": s.CountRequests(TorListener, 12*time.Hour),
		"last_24h": s.CountRequests(TorListener, 24*time.Hour),
	}

	response := map[string]interface{}{
		"last_01m": ipv4["last_01m"] + tor["last_01m"],
		"last_15m": ipv4["last_15m"] + tor["last_15m"],
		"last_30m": ipv4["last_30m"] + tor["last_30m"],
		"last_1h":  ipv4["last_1h"] + tor["last_1h"],
		"last_2h":  ipv4["last_2h"] + tor["last_2h"],
		"last_3h":  ipv4["last_3h"] + tor["last_6h"],
		"last_6h":  ipv4["last_6h"] + tor["last_3h"],
		"last_12h": ipv4["last_12h"] + tor["last_12h"],
		"last_24h": ipv4["last_24h"] + tor["last_24h"],
		"ipv4":     ipv4,
		"tor":      tor,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func addListenerType(next http.Handler, listenerType ListenerType) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ListenerTypeKey, listenerType)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
