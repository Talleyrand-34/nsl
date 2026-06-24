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

// Package observ provides observability for the API: a configured slog logger,
// HTTP middleware, and an in-memory registry of "runs" (long operations such as
// scans) that buffer granular events for the UI to poll while also writing them
// to the structured log. The same Emit call feeds both the server log and the
// live status feed, so they never diverge.
package observ

import (
	"log/slog"
	"sync"
	"time"
)

// maxEventsPerRun caps the per-run event buffer so a huge scan can't grow memory
// without bound; the oldest events are dropped (their seq numbers are preserved
// so a polling client can detect the gap).
const maxEventsPerRun = 2000

// Event is one granular thing that happened during a run.
type Event struct {
	Seq    int            `json:"seq"`
	Time   time.Time      `json:"time"`
	Level  string         `json:"level"`
	Msg    string         `json:"msg"`
	Fields map[string]any `json:"fields,omitempty"`
}

// Run is a tracked long-running operation (e.g. a scan). It is safe for
// concurrent use: the worker goroutine emits while pollers read snapshots.
type Run struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`  // "run" | "network" | "host" | "host-ssh" | "connections"
	Title string `json:"title"` // human summary, e.g. "snmp 10.0.2.0/26"

	mu      sync.Mutex
	state   string // "running" | "completed" | "failed"
	start   time.Time
	end     time.Time
	done    int
	total   int
	seq     int
	events  []Event
	dropped int // number of events evicted from the front of the buffer
	result  any
	err     string
}

// Emitter is the minimal surface the service layer needs to report progress; a
// *Run satisfies it. A nil Emitter is a no-op, so callers (e.g. the CLI) that
// don't track a run can pass nil.
type Emitter interface {
	Emit(level, msg string, fields ...any)
	Progress(done, total int)
}

// discard is a no-op Emitter for callers (e.g. the CLI) that don't track a run.
type discard struct{}

func (discard) Emit(string, string, ...any) {}
func (discard) Progress(int, int)           {}

// Discard is the shared no-op Emitter. Service code normalizes a nil Emitter to
// this so emit calls are always safe.
var Discard Emitter = discard{}

// RunStatus is the JSON snapshot returned to a polling client.
type RunStatus struct {
	ScanID  string  `json:"scan_id"`
	Kind    string  `json:"kind"`
	Title   string  `json:"title"`
	State   string  `json:"state"`
	Done    int     `json:"done"`
	Total   int     `json:"total"`
	Elapsed string  `json:"elapsed"`
	LastSeq int     `json:"last_seq"`
	Dropped int     `json:"dropped,omitempty"`
	Events  []Event `json:"events"`
	Result  any     `json:"result,omitempty"`
	Error   string  `json:"error,omitempty"`
}

// fieldsToMap turns a variadic key,value,... list into a map and a slog-friendly
// attr slice. Non-string or dangling keys are ignored defensively.
func fieldsToAttrs(fields []any) (map[string]any, []any) {
	if len(fields) == 0 {
		return nil, nil
	}
	m := make(map[string]any, len(fields)/2)
	attrs := make([]any, 0, len(fields))
	for i := 0; i+1 < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		m[key] = fields[i+1]
		attrs = append(attrs, slog.Any(key, fields[i+1]))
	}
	return m, attrs
}

// Emit records an event on the run (capped buffer) and writes the same line to
// the structured log, tagged with the run id. level is "debug"|"info"|"warn"|"error".
func (r *Run) Emit(level, msg string, fields ...any) {
	if r == nil {
		return
	}
	m, attrs := fieldsToAttrs(fields)

	r.mu.Lock()
	r.seq++
	ev := Event{Seq: r.seq, Time: time.Now(), Level: level, Msg: msg, Fields: m}
	r.events = append(r.events, ev)
	if len(r.events) > maxEventsPerRun {
		drop := len(r.events) - maxEventsPerRun
		r.events = r.events[drop:]
		r.dropped += drop
	}
	r.mu.Unlock()

	logAt(level, msg, append([]any{slog.String("run", r.ID)}, attrs...)...)
}

// State returns the run's current state ("running"|"completed"|"failed").
func (r *Run) State() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state
}

// Progress updates the run's done/total counters.
func (r *Run) Progress(done, total int) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.done, r.total = done, total
	r.mu.Unlock()
}

// Finish marks the run completed and stores its result payload.
func (r *Run) Finish(result any) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.state = "completed"
	r.end = time.Now()
	r.result = result
	elapsed := r.end.Sub(r.start)
	r.mu.Unlock()
	r.Emit("info", "run completed", "elapsed", elapsed.String())
}

// Fail marks the run failed and records the error.
func (r *Run) Fail(err error) {
	if r == nil {
		return
	}
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	r.mu.Lock()
	r.state = "failed"
	r.end = time.Now()
	r.err = msg
	r.mu.Unlock()
	r.Emit("error", "run failed", "error", msg)
}

// Snapshot returns the run's status including only events with Seq > sinceSeq
// (the polling cursor). The result is included only once the run is completed.
func (r *Run) Snapshot(sinceSeq int) RunStatus {
	r.mu.Lock()
	defer r.mu.Unlock()

	fresh := []Event{}
	for _, e := range r.events {
		if e.Seq > sinceSeq {
			fresh = append(fresh, e)
		}
	}
	end := r.end
	if r.state == "running" {
		end = time.Now()
	}
	st := RunStatus{
		ScanID:  r.ID,
		Kind:    r.Kind,
		Title:   r.Title,
		State:   r.state,
		Done:    r.done,
		Total:   r.total,
		Elapsed: end.Sub(r.start).Truncate(time.Millisecond).String(),
		LastSeq: r.seq,
		Dropped: r.dropped,
		Events:  fresh,
	}
	if r.state == "completed" {
		st.Result = r.result
	}
	if r.state == "failed" {
		st.Error = r.err
	}
	return st
}

// Registry is a bounded, concurrency-safe store of recent runs.
type Registry struct {
	mu    sync.Mutex
	runs  map[string]*Run
	order []string // insertion order, oldest first, for LRU-style eviction
	max   int
	seq   uint64 // monotonic counter for unique ids
}

// NewRegistry creates a registry retaining at most max recent runs.
func NewRegistry(max int) *Registry {
	if max < 1 {
		max = 1
	}
	return &Registry{runs: make(map[string]*Run), max: max}
}

// NewRun creates and registers a running Run, evicting the oldest if over capacity.
func (reg *Registry) NewRun(kind, title string) *Run {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	reg.seq++
	id := "scan-" + time.Now().Format("20060102T150405") + "-" + itoa(reg.seq)
	run := &Run{ID: id, Kind: kind, Title: title, state: "running", start: time.Now()}
	reg.runs[id] = run
	reg.order = append(reg.order, id)
	for len(reg.order) > reg.max {
		oldest := reg.order[0]
		reg.order = reg.order[1:]
		delete(reg.runs, oldest)
	}
	run.Emit("info", "run started", "kind", kind, "title", title)
	return run
}

// Get returns a run by id.
func (reg *Registry) Get(id string) (*Run, bool) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	r, ok := reg.runs[id]
	return r, ok
}

// Runs is the process-wide registry used by the API handlers.
var Runs = NewRegistry(50)

// itoa avoids importing strconv for a single small conversion.
func itoa(n uint64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
