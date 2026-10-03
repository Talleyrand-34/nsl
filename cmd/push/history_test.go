// history_test.go: TDD for push history.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package cmd_push

import (
	"testing"

	"nsl-graph/internal/push"
)

type fakeSnapStore struct {
	snaps []push.ConfigSnapshot
}

func (f *fakeSnapStore) Save(snap push.ConfigSnapshot) (string, error) { return snap.ID, nil }
func (f *fakeSnapStore) Get(id string) (push.ConfigSnapshot, error) {
	for _, s := range f.snaps {
		if s.ID == id {
			return s, nil
		}
	}
	return push.ConfigSnapshot{}, nil
}
func (f *fakeSnapStore) List(deviceID string, limit int) ([]push.ConfigSnapshot, error) {
	return f.snaps, nil
}
func (f *fakeSnapStore) Latest(deviceID string) (push.ConfigSnapshot, error) {
	return push.ConfigSnapshot{}, nil
}

func TestHistoryCmd_ListsSnaps(t *testing.T) {
	store := &push.Store{
		Runs: push.NewMemoryRunStore(),
		Snap: &fakeSnapStore{
			snaps: []push.ConfigSnapshot{
				{ID: "snap-001", DeviceID: "R1", OS: "openwrt"},
				{ID: "snap-002", DeviceID: "R1", OS: "openwrt"},
			},
		},
	}
	historyStore = store

	cmd := HistoryCmd
	cmd.SetArgs([]string{"--device", "R1"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
}
