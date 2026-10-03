// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package observ

import "testing"

func TestRunLifecycleAndCursor(t *testing.T) {
	reg := NewRegistry(10)
	run := reg.NewRun("run", "snmp 10.0.0.0/30")
	if run.State() != "running" {
		t.Fatalf("new run should be running, got %q", run.State())
	}

	// NewRun emits a "run started" event (seq 1).
	run.Emit("info", "probing", "ip", "10.0.0.1")
	run.Progress(1, 2)
	run.Emit("info", "device discovered", "ip", "10.0.0.1")

	// Cursor from 0 returns all events so far.
	snap := run.Snapshot(0)
	if snap.State != "running" || snap.Done != 1 || snap.Total != 2 {
		t.Fatalf("unexpected snapshot: %+v", snap)
	}
	if len(snap.Events) < 3 {
		t.Fatalf("expected >=3 events, got %d", len(snap.Events))
	}
	last := snap.LastSeq

	// Cursor at last seq returns no new events.
	if got := run.Snapshot(last); len(got.Events) != 0 {
		t.Errorf("expected no new events after cursor %d, got %d", last, len(got.Events))
	}

	// Finish stores the result and exposes it only when completed.
	run.Finish(map[string]any{"devices": []string{"a"}})
	done := run.Snapshot(last)
	if done.State != "completed" {
		t.Fatalf("expected completed, got %q", done.State)
	}
	if done.Result == nil {
		t.Error("completed snapshot should include the result")
	}

	// Lookup by id.
	if got, ok := reg.Get(run.ID); !ok || got != run {
		t.Error("registry should return the run by id")
	}
}

func TestRegistryEviction(t *testing.T) {
	reg := NewRegistry(2)
	a := reg.NewRun("run", "a")
	b := reg.NewRun("run", "b")
	c := reg.NewRun("run", "c") // evicts a

	if _, ok := reg.Get(a.ID); ok {
		t.Error("oldest run should have been evicted")
	}
	if _, ok := reg.Get(b.ID); !ok {
		t.Error("b should still be present")
	}
	if _, ok := reg.Get(c.ID); !ok {
		t.Error("c should be present")
	}
}

func TestFailRecordsError(t *testing.T) {
	reg := NewRegistry(2)
	run := reg.NewRun("run", "x")
	run.Fail(errStub("boom"))
	snap := run.Snapshot(0)
	if snap.State != "failed" || snap.Error != "boom" {
		t.Fatalf("expected failed/boom, got state=%q err=%q", snap.State, snap.Error)
	}
}

func TestNilEmitterAndDiscard(t *testing.T) {
	var run *Run             // nil
	run.Emit("info", "noop") // must not panic
	run.Progress(1, 2)       // must not panic
	Discard.Emit("info", "noop")
	Discard.Progress(0, 0)
}

type errStub string

func (e errStub) Error() string { return string(e) }
