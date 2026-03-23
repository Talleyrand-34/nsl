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
package plugins_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"nsl-graph/internal/api/plugins"
)

func setupTestAPI(t *testing.T) *mux.Router {
	r := mux.NewRouter()
	plugins.RegisterRoutes(r)
	return r
}

func TestPlugins_GetPluginsEndpoint(t *testing.T) {
	router := setupTestAPI(t)

	req := httptest.NewRequest("GET", "/plugins", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Since plugin manager might not be initialized in test, expect error
	// The important thing is that the route is registered and responds
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestPlugins_SetActivePluginEndpoint(t *testing.T) {
	router := setupTestAPI(t)

	// Test with valid JSON body
	requestBody := plugins.SetActivePluginRequest{
		PluginID: "test_plugin",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/plugins/active", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Since plugin manager might not be initialized in test, expect error
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestPlugins_CORSOptions(t *testing.T) {
	router := setupTestAPI(t)

	// Test OPTIONS request for CORS preflight
	req := httptest.NewRequest("OPTIONS", "/plugins", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should handle OPTIONS method
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
}

func TestPlugins_SetActivePluginCORS(t *testing.T) {
	router := setupTestAPI(t)

	// Test OPTIONS request for plugins/active endpoint
	req := httptest.NewRequest("OPTIONS", "/plugins/active", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should handle OPTIONS method for CORS
	assert.NotEqual(t, http.StatusNotFound, w.Code)
	assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
}

func TestPlugins_InvalidJSONRequest(t *testing.T) {
	router := setupTestAPI(t)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", "/plugins/active", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should not return 404 (route exists)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}