package httpapi

import (
	"log"
	"net/http"
	"time"
)

type statusRecorder struct {
	status int
	writer http.ResponseWriter
}

func (r *statusRecorder) Header() http.Header {
	return r.writer.Header()
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.writer.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.writer.Write(data)
}

func RequestLogger(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{writer: w}
		start := time.Now()
		next.ServeHTTP(rec, r)
		duration := time.Since(start)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}

		logger.Printf(
			"http request method=%s path=%s status=%d duration=%s ip=%s",
			r.Method,
			r.URL.Path,
			status,
			duration.String(),
			r.RemoteAddr,
		)
	})
}
