package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Stats struct {
	sync.RWMutex
	Requests []int64
}

func NewStats() *Stats {
	return &Stats{
		Requests: make([]int64, 0),
	}
}

func (s *Stats) RecordRequest() {
	s.Lock()
	defer s.Unlock()
	now := time.Now().Unix()
	s.Requests = append(s.Requests, now)
}

func (s *Stats) CountRequests(duration time.Duration) int {
	s.RLock()
	defer s.RUnlock()
	now := time.Now().Unix()
	threshold := now - int64(duration.Seconds())
	count := 0
	for _, v := range s.Requests {
		if v >= threshold {
			count++
		}
	}
	return count
}

func (s *Stats) Cleanup(duration time.Duration) {
	s.Lock()
	defer s.Unlock()
	oneHourAgo := time.Now().Add(-1 * duration).Unix()
	thresholdIndex := 0
	for i, v := range s.Requests {
		if v > oneHourAgo {
			thresholdIndex = i
			break
		}
	}
	// Remove older entries
	s.Requests = s.Requests[thresholdIndex:]
}

func (s *Stats) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.RecordRequest()
		next.ServeHTTP(w, r)
	})
}

func (s *Stats) StatsHandler(w http.ResponseWriter, r *http.Request) {
	s.Cleanup(24 * time.Hour)
	response := map[string]int{
		"last_01m": s.CountRequests(time.Minute),
		"last_15m": s.CountRequests(15 * time.Minute),
		"last_30m": s.CountRequests(30 * time.Minute),
		"last_1h":  s.CountRequests(time.Hour),
		"last_2h":  s.CountRequests(2 * time.Hour),
		"last_3h":  s.CountRequests(3 * time.Hour),
		"last_6h":  s.CountRequests(6 * time.Hour),
		"last_12h": s.CountRequests(12 * time.Hour),
		"last_24h": s.CountRequests(24 * time.Hour),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
