package server

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type LogRecord struct {
	http.ResponseWriter
	status int
}

func (r *LogRecord) Write(p []byte) (int, error) {
	return r.ResponseWriter.Write(p)
}

func (r *LogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func WrapHandler(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		record := &LogRecord{
			ResponseWriter: w,
		}

		f.ServeHTTP(record, r)

		switch record.status {
		case http.StatusBadRequest:
			fallthrough
		case http.StatusNotFound:
			log.Error("Bad Request %+v\n", *r)
		}
	}
}
