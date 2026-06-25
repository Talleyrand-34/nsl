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
package basicops

import (
	"fmt"

	d "github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"
)

// GetVaultMeta returns the persisted vault metadata blob (wrapped data key + salt),
// or "" if the vault has never been initialized.
func (r BasicOpsCloverRepository) GetVaultMeta() (string, error) {
	doc, err := r.db.FindFirst(q.NewQuery(vaultCollection))
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", nil
	}
	if m, ok := doc.Get("meta").(string); ok {
		return m, nil
	}
	return "", nil
}

// SetVaultMeta persists the vault metadata blob, replacing any existing one (the
// vault holds a single meta document).
func (r BasicOpsCloverRepository) SetVaultMeta(meta string) error {
	if err := r.db.Delete(q.NewQuery(vaultCollection)); err != nil {
		return fmt.Errorf("SetVaultMeta: clearing previous meta: %w", err)
	}
	doc := d.NewDocument()
	doc.Set("meta", meta)
	if _, err := r.db.InsertOne(vaultCollection, doc); err != nil {
		return fmt.Errorf("SetVaultMeta: %w", err)
	}
	return nil
}
