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

	e "nsl-graph/internal/repository/entities"
)

// scanprofileDeviceToDoc + docToScanprofileDevice map ProfileDevice rows
// to/from Clover documents. The primary key in storage is
// (profile_name, host) — uniqueness enforced at the application layer
// (see AddProfileDevice) because clover's query language does not
// expose compound-key constraints.
func scanprofileDeviceToDoc(d_ e.ProfileDevice) *d.Document {
	doc := d.NewDocument()
	doc.Set("profile_name", d_.ProfileName)
	doc.Set("host", d_.Host)
	doc.Set("ssh_profile_name", d_.SSHProfileName)
	doc.Set("ssh_config_text", d_.SSHConfigText)
	doc.Set("ssh_key_filename", d_.SSHKeyFilename)
	doc.Set("ssh_key", d_.SSHKey) // server-side uploaded content (no encryption here; transient)
	doc.Set("port", d_.Port)
	return doc
}

func docToScanprofileDevice(doc *d.Document) e.ProfileDevice {
	getStr := func(k string) string {
		if v, ok := doc.Get(k).(string); ok {
			return v
		}
		return ""
	}
	getInt := func(k string) int {
		switch v := doc.Get(k).(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
		return 0
	}
	return e.ProfileDevice{
		ID:             doc.ObjectId(),
		ProfileName:    getStr("profile_name"),
		Host:           getStr("host"),
		SSHProfileName: getStr("ssh_profile_name"),
		SSHConfigText:  getStr("ssh_config_text"),
		SSHKeyFilename: getStr("ssh_key_filename"),
		SSHKey:         getStr("ssh_key"),
		Port:           getInt("port"),
	}
}

// AddProfileDevice inserts a new (profile_name, host) row. Returns an
// error if the host is already in the profile; callers should treat that
// as 409 Conflict upstream.
func (r BasicOpsCloverRepository) AddProfileDevice(d_ e.ProfileDevice) error {
	if d_.ProfileName == "" {
		return fmt.Errorf("profile_name is required")
	}
	if d_.Host == "" {
		return fmt.Errorf("host is required")
	}
	exists, err := r.db.Exists(q.NewQuery(scanProfileDevicesCollection).
		Where(q.Field("profile_name").Eq(d_.ProfileName).
			And(q.Field("host").Eq(d_.Host))))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("device %q is already in profile %q", d_.Host, d_.ProfileName)
	}
	_, err = r.db.InsertOne(scanProfileDevicesCollection, scanprofileDeviceToDoc(d_))
	return err
}

// GetProfileDevices returns every row attached to the given profile, in
// insertion order (Clover's natural document order is insertion order).
func (r BasicOpsCloverRepository) GetProfileDevices(profileName string) ([]e.ProfileDevice, error) {
	docs, err := r.db.FindAll(q.NewQuery(scanProfileDevicesCollection).
		Where(q.Field("profile_name").Eq(profileName)))
	if err != nil {
		return []e.ProfileDevice{}, err
	}
	result := make([]e.ProfileDevice, 0, len(docs))
	for _, doc := range docs {
		result = append(result, docToScanprofileDevice(doc))
	}
	return result, nil
}

// DeleteProfileDevice removes a single (profile_name, host) row.
func (r BasicOpsCloverRepository) DeleteProfileDevice(profileName, host string) error {
	return r.db.Delete(q.NewQuery(scanProfileDevicesCollection).
		Where(q.Field("profile_name").Eq(profileName).
			And(q.Field("host").Eq(host))))
}

// DeleteAllProfileDevices removes every row attached to a profile. Used
// when the profile itself is deleted or when the operator replaces the
// device list wholesale.
func (r BasicOpsCloverRepository) DeleteAllProfileDevices(profileName string) error {
	return r.db.Delete(q.NewQuery(scanProfileDevicesCollection).
		Where(q.Field("profile_name").Eq(profileName)))
}