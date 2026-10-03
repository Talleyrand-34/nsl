// push_run_test.go: TDD for Clover-backed PushRun persistence.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package basicops

import (
	"testing"
	"time"

	"nsl-graph/internal/configparser"
	p "nsl-graph/internal/push"
)

func TestCloverPushRun_AppendAndAll(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setupTestCloverRepository: %v", err)
	}
	defer cleanup()

	run := p.PushRun{
		ID:       "push-001",
		DeviceID: "R1",
		OS:       "openwrt",
		Safety:   configparser.SafetyApply,
		Renderer: "OpenWrtRenderer",
		StartedAt: time.Now(),
		FinishedAt: time.Now().Add(2 * time.Second),
		BackupPath: "/tmp/backup.gz",
		SnapshotID: "snap-001",
		BackupRelpath: "R1/push-001.gz",
		ExitStatus: "success",
	}
	if err := repo.AppendPushRun(run); err != nil {
		t.Fatalf("AppendPushRun: %v", err)
	}

	runs, err := repo.AllPushRuns()
	if err != nil {
		t.Fatalf("AllPushRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].ID != run.ID {
		t.Errorf("ID: got %q, want %q", runs[0].ID, run.ID)
	}
	if runs[0].SnapshotID != run.SnapshotID {
		t.Errorf("SnapshotID: got %q, want %q", runs[0].SnapshotID, run.SnapshotID)
	}
	if runs[0].BackupRelpath != run.BackupRelpath {
		t.Errorf("BackupRelpath: got %q, want %q", runs[0].BackupRelpath, run.BackupRelpath)
	}
	if runs[0].ExitStatus != "success" {
		t.Errorf("ExitStatus: got %q, want success", runs[0].ExitStatus)
	}
}

func TestCloverPushRun_MultipleRuns(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setupTestCloverRepository: %v", err)
	}
	defer cleanup()

	for i := 1; i <= 3; i++ {
		repo.AppendPushRun(p.PushRun{
			ID:       "push-001",
			DeviceID: "R1",
			OS:       "openwrt",
			Safety:   configparser.SafetyDryRun,
			Renderer: "OpenWrtRenderer",
			StartedAt: time.Now().Add(time.Duration(i) * time.Minute),
			ExitStatus: "success",
		})
	}

	runs, _ := repo.AllPushRuns()
	if len(runs) != 3 {
		t.Errorf("expected 3 runs, got %d", len(runs))
	}
	// Newest first
	if !runs[0].StartedAt.After(runs[2].StartedAt) {
		t.Error("runs should be sorted newest-first")
	}
}
