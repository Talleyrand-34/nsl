/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package observ

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

// Init configures the process-wide structured logger.
//
//	level  : "debug" | "info" | "warn" | "error" (default info)
//	format : "text" (human-readable, default) | "json"
//	file   : optional path; when set, logs go to both stdout and the file.
func Init(level, format, file string) error {
	var lvl slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lvl = slog.LevelDebug
	case "", "info":
		lvl = slog.LevelInfo
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		return fmt.Errorf("invalid log level %q (want debug|info|warn|error)", level)
	}

	var w io.Writer = os.Stdout
	if file != "" {
		f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("open log file %q: %w", file, err)
		}
		w = io.MultiWriter(os.Stdout, f)
	}

	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		h = slog.NewJSONHandler(w, opts)
	case "", "text":
		h = slog.NewTextHandler(w, opts)
	default:
		return fmt.Errorf("invalid log format %q (want text|json)", format)
	}
	slog.SetDefault(slog.New(h))
	return nil
}

// logAt logs at a level named by a string (used by Run.Emit so the same level
// vocabulary serves both the event feed and the log).
func logAt(level, msg string, attrs ...any) {
	switch level {
	case "debug":
		slog.Debug(msg, attrs...)
	case "warn", "warning":
		slog.Warn(msg, attrs...)
	case "error":
		slog.Error(msg, attrs...)
	default:
		slog.Info(msg, attrs...)
	}
}

var reqCounter atomic.Uint64

// statusRecorder captures the status code and byte count of a response.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// Flush proxies http.Flusher so streaming handlers keep working through the wrapper.
func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// RequestLogger is mux middleware that logs one structured line per request with
// a request id, method, path, status, size and duration.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := fmt.Sprintf("req-%d", reqCounter.Add(1))
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		// The scan-status endpoint is polled ~1/s by the UI, so log it at debug to
		// avoid drowning the log; server errors are logged at error.
		level := "info"
		switch {
		case rec.status >= 500:
			level = "error"
		case r.URL.Path == "/scan/status":
			level = "debug"
		}
		logAt(level, "http request",
			"req", reqID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"bytes", rec.bytes,
			"dur", time.Since(start).Truncate(time.Microsecond).String(),
			"remote", r.RemoteAddr,
		)
	})
}

// Recover is mux middleware that turns a handler panic into a logged 500 instead
// of crashing the server.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				slog.Error("handler panic",
					"path", r.URL.Path,
					"panic", fmt.Sprintf("%v", p),
					"stack", string(debug.Stack()),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal_error","message":"an unexpected error occurred"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
