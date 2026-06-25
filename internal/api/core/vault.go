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
package core

import (
	"encoding/json"
	"net/http"

	q "nsl-graph/internal/repository/application"
)

func vaultHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// VaultStatusHandler reports whether the credential vault is initialized and unlocked.
func VaultStatusHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vaultHeaders(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		v := service.Vault()
		json.NewEncoder(w).Encode(map[string]bool{
			"initialized": v.Initialized(),
			"unlocked":    v.Unlocked(),
		})
	}
}

// passphraseBody is the shared request shape for init/unlock.
type passphraseBody struct {
	Passphrase string `json:"passphrase"`
}

func decodePassphrase(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req passphraseBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_json", "message": err.Error()})
		return "", false
	}
	if req.Passphrase == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing_passphrase", "message": "passphrase is required"})
		return "", false
	}
	return req.Passphrase, true
}

// VaultInitHandler sets the master passphrase for a fresh vault (leaves it unlocked).
func VaultInitHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vaultHeaders(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		pass, ok := decodePassphrase(w, r)
		if !ok {
			return
		}
		if err := service.Vault().Init(pass); err != nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "init_failed", "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Vault initialized and unlocked"})
	}
}

// VaultUnlockHandler unlocks the vault with the master passphrase.
func VaultUnlockHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vaultHeaders(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		pass, ok := decodePassphrase(w, r)
		if !ok {
			return
		}
		if err := service.Vault().Unlock(pass); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unlock_failed", "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Vault unlocked"})
	}
}

// VaultLockHandler locks the vault, clearing the in-memory key.
func VaultLockHandler(service q.NetServiceInt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vaultHeaders(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		service.Vault().Lock()
		json.NewEncoder(w).Encode(map[string]string{"message": "Vault locked"})
	}
}
