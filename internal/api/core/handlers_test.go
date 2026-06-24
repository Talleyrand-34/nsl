/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)

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
package core_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"nsl-graph/internal/api/core"
	"nsl-graph/internal/repository/application"
	"nsl-graph/internal/repository/infra/cloverdb/base"
)

func setupTestAPI(t *testing.T) (*mux.Router, application.NetServiceInt) {
	// Create temporary directory for test database
	tempDir := t.TempDir()

	repo, err := basicops.NewCloverRepository(tempDir)
	assert.NoError(t, err)

	service := application.NewNetService(repo)

	r := mux.NewRouter()
	core.RegisterRoutes(r, service)

	return r, service
}

func TestCore_RootEndpoint(t *testing.T) {
	router, _ := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse the JSON response
	var endpoints []string
	err := json.Unmarshal(w.Body.Bytes(), &endpoints)
	assert.NoError(t, err)

	// Check that we have a reasonable number of endpoints
	assert.Greater(t, len(endpoints), 10, "Should have multiple API endpoints listed")

	// Check for some expected endpoints
	endpointStrings := w.Body.String()
	assert.Contains(t, endpointStrings, "GET    /brands")
	assert.Contains(t, endpointStrings, "POST   /brands")
	assert.Contains(t, endpointStrings, "GET    /devices")
	assert.Contains(t, endpointStrings, "GET    /diagram")
}

func TestCore_DiagramEndpoint(t *testing.T) {
	router, _ := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/diagram", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// With no devices the diagram can't be rendered: the endpoint signals this
	// with 404 + a reason so the web <img> falls back to its alt text.
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "No devices")
}

func TestCore_RootEndpointContentStructure(t *testing.T) {
	router, _ := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var endpoints []string
	err := json.Unmarshal(w.Body.Bytes(), &endpoints)
	assert.NoError(t, err)

	// Verify endpoint format (should include HTTP method and path)
	for _, endpoint := range endpoints {
		// Each endpoint should have format like "GET    /path" or "POST   /path"
		assert.Regexp(t, `^(GET|POST|PUT|DELETE)\s+/\w+`, endpoint,
			"Endpoint should have format 'METHOD /path', got: %s", endpoint)
	}
}
