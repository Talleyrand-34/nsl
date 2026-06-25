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
package devices_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"nsl-graph/internal/api/devices"
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
	devices.RegisterRoutes(r, service)

	return r, service
}

func TestDevices_BrandsEndpoint(t *testing.T) {
	router, _ := setupTestAPI(t)

	// Test GET /brands endpoint
	req := httptest.NewRequest("GET", "/brands", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 OK with empty brands list
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDevices_ModelTypesEndpoint(t *testing.T) {
	router, _ := setupTestAPI(t)

	// Test GET /modeltypes endpoint
	req := httptest.NewRequest("GET", "/modeltypes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 OK with empty device classes list
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDevices_ModelsEndpoint(t *testing.T) {
	router, _ := setupTestAPI(t)

	// Test GET /models endpoint
	req := httptest.NewRequest("GET", "/models", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 OK with empty models list
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDevices_DevicesEndpoint(t *testing.T) {
	router, _ := setupTestAPI(t)

	// Test GET /devices endpoint
	req := httptest.NewRequest("GET", "/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 OK with empty devices list
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDevices_ModelPortsEndpoint(t *testing.T) {
	router, _ := setupTestAPI(t)

	// Test GET /modelports endpoint
	req := httptest.NewRequest("GET", "/modelports", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 OK with empty model ports list
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDevices_DevicePortsEndpoint(t *testing.T) {
	router, _ := setupTestAPI(t)

	// Test GET /deviceports endpoint
	req := httptest.NewRequest("GET", "/deviceports", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 200 OK with empty device ports list
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDevices_RoutesRegistration(t *testing.T) {
	router, _ := setupTestAPI(t)

	// Test that all expected routes are registered by checking route matching
	testRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/brands"},
		{"POST", "/brands"},
		{"PUT", "/brands"},
		{"DELETE", "/brands"},
		{"GET", "/modeltypes"},
		{"POST", "/modeltypes"},
		{"PUT", "/modeltypes"},
		{"DELETE", "/modeltypes"},
		{"GET", "/devices"},
		{"POST", "/devices"},
		{"PUT", "/devices"},
		{"DELETE", "/devices"},
		{"GET", "/modelports"},
		{"POST", "/modelports"},
		{"PUT", "/modelports"},
		{"DELETE", "/modelports"},
		{"GET", "/deviceports"},
		{"POST", "/deviceports"},
		{"PUT", "/deviceports"},
		{"DELETE", "/deviceports"},
		{"GET", "/allports/all"},
		{"GET", "/allports/device"},
	}

	for _, route := range testRoutes {
		req := httptest.NewRequest(route.method, route.path, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not return 404 (route not found)
		assert.NotEqual(t, http.StatusNotFound, w.Code,
			"Route %s %s should be registered", route.method, route.path)
	}
}
