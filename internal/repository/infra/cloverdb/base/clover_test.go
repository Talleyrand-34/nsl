// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package basicops

import (
	"testing"
)

// setupTestCloverRepository creates a test repository using a temporary directory
// Returns the repository, a cleanup function, and an error
func setupTestCloverRepository(t *testing.T) (BasicOpsCloverRepository, func(), error) {
	// Create temporary directory for test database
	tempDir := t.TempDir()

	// Create repository using the proper constructor
	repo, err := NewCloverRepository(tempDir)
	if err != nil {
		return BasicOpsCloverRepository{}, nil, err
	}

	// Cleanup function
	cleanup := func() {
		repo.Close()
	}

	return repo, cleanup, nil
}
