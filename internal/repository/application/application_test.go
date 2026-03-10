/*
Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

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
package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nsl-graph/internal/repository/application"
	cloverdb "nsl-graph/internal/repository/infra/cloverdb/base"
)

func setupTestRepository(t *testing.T) cloverdb.BasicOpsCloverRepository {
	// Create temporary directory for test database
	tempDir := t.TempDir()

	repo, err := cloverdb.NewCloverRepository(tempDir)
	assert.NoError(t, err)

	return repo
}

func TestNetService_AddAndGetBrand(t *testing.T) {
	repo := setupTestRepository(t)

	service := application.NewNetService(repo)

	// Test AddBrand
	err := service.AddBrand("TestBrand")
	assert.NoError(t, err)

	// Test GetBrands
	brands, err := service.GetBrands()
	assert.NoError(t, err)

	found := false
	for _, brand := range brands {
		if brand.Name == "TestBrand" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find brand with name 'TestBrand'")
}
