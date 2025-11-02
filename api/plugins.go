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
package api

import (
	"encoding/json"
	"net/http"

	"nsl-graph/internal/repository/plugins"
)

// PluginInfo represents metadata about a plugin
type PluginInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// getPluginsHandler returns a list of all available connection sorter plugins
func getPluginsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if plugins.GlobalPluginManager == nil {
			http.Error(w, "Plugin manager not initialized", http.StatusInternalServerError)
			return
		}

		// Get list of available sorters
		sorterIDs := plugins.GlobalPluginManager.ListAvailableSorters()
		activeSorter := plugins.GlobalPluginManager.GetActiveSorter()

		var pluginList []PluginInfo
		for _, id := range sorterIDs {
			sorter := plugins.GlobalPluginManager.GetSorter(id)
			if sorter != nil {
				pluginList = append(pluginList, PluginInfo{
					ID:          sorter.ID(),
					Name:        sorter.Name(),
					Description: sorter.Description(),
					IsActive:    activeSorter != nil && sorter.ID() == activeSorter.ID(),
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pluginList)
	}
}

// SetActivePluginRequest represents the request body for changing the active plugin
type SetActivePluginRequest struct {
	PluginID string `json:"plugin_id"`
}

// setActivePluginHandler changes the active connection sorter plugin
func setActivePluginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if plugins.GlobalPluginManager == nil {
			http.Error(w, "Plugin manager not initialized", http.StatusInternalServerError)
			return
		}

		// Parse request body
		var req SetActivePluginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate plugin ID
		if req.PluginID == "" {
			http.Error(w, "plugin_id is required", http.StatusBadRequest)
			return
		}

		// Set active sorter
		if err := plugins.GlobalPluginManager.SetActiveSorter(req.PluginID); err != nil {
			http.Error(w, "Failed to set active plugin: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"plugin_id": req.PluginID,
			"message":   "Active plugin updated successfully",
		})
	}
}
