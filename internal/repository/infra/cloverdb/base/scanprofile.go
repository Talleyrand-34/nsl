// SPDX-License-Identifier: AGPL-3.0-or-later
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
package basicops

import (
	"fmt"

	d "github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"

	e "nsl-graph/internal/repository/entities"
)

func scanProfileToDoc(p e.ScanProfile) *d.Document {
	doc := d.NewDocument()
	doc.Set("name", p.Name)
	doc.Set("kind", p.Kind)
	doc.Set("host", p.Host)
	doc.Set("snmp_community", p.SNMPCommunity)
	doc.Set("snmp_version", p.SNMPVersion)
	doc.Set("snmp_port", p.SNMPPort)
	doc.Set("timeout_sec", p.TimeoutSec)
	doc.Set("scan_source", p.ScanSource)
	doc.Set("config_source", p.ConfigSource)
	doc.Set("config_file", p.ConfigFile)
	doc.Set("os_type", p.OsType)
	doc.Set("ssh_user", p.SSHUser)
	doc.Set("ssh_password", p.SSHPassword) // already an encrypted blob (or "")
	doc.Set("ssh_key_file", p.SSHKeyFile)
	doc.Set("ssh_key", p.SSHKey) // encrypted PEM key content (or "")
	doc.Set("ssh_port", p.SSHPort)
	doc.Set("discrepancy_action", p.DiscrepancyAction)
	doc.Set("merge_configs", p.MergeConfigs)
	doc.Set("config_timeout", p.ConfigTimeout)
	doc.Set("vlan_accuracy", p.VLANAccuracy)
	return doc
}

func docToScanProfile(doc *d.Document) e.ScanProfile {
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
	getBool := func(k string) bool {
		if v, ok := doc.Get(k).(bool); ok {
			return v
		}
		return false
	}
	kind := getStr("kind")
	if kind == "" {
		kind = "device" // back-compat: profiles stored before kinds existed
	}
	return e.ScanProfile{
		ID:                doc.ObjectId(),
		Name:              getStr("name"),
		Kind:              kind,
		Host:              getStr("host"),
		SNMPCommunity:     getStr("snmp_community"),
		SNMPVersion:       getStr("snmp_version"),
		SNMPPort:          getInt("snmp_port"),
		TimeoutSec:        getInt("timeout_sec"),
		ScanSource:        getStr("scan_source"),
		ConfigSource:      getStr("config_source"),
		ConfigFile:        getStr("config_file"),
		OsType:            getStr("os_type"),
		SSHUser:           getStr("ssh_user"),
		SSHPassword:       getStr("ssh_password"),
		SSHKeyFile:        getStr("ssh_key_file"),
		SSHKey:            getStr("ssh_key"),
		SSHPort:           getInt("ssh_port"),
		DiscrepancyAction: getStr("discrepancy_action"),
		MergeConfigs:      getBool("merge_configs"),
		ConfigTimeout:     getInt("config_timeout"),
		VLANAccuracy:      getInt("vlan_accuracy"),
	}
}

// AddScanProfile inserts a new scan profile (name must be unique).
func (r BasicOpsCloverRepository) AddScanProfile(p e.ScanProfile) error {
	exists, err := r.db.Exists(q.NewQuery(scanProfilesCollection).Where(q.Field("name").Eq(p.Name)))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("scan profile %q already exists", p.Name)
	}
	_, err = r.db.InsertOne(scanProfilesCollection, scanProfileToDoc(p))
	return err
}

// GetScanProfiles returns all scan profiles.
func (r BasicOpsCloverRepository) GetScanProfiles() ([]e.ScanProfile, error) {
	docs, err := r.db.FindAll(q.NewQuery(scanProfilesCollection))
	if err != nil {
		return []e.ScanProfile{}, err
	}
	result := make([]e.ScanProfile, 0, len(docs))
	for _, doc := range docs {
		result = append(result, docToScanProfile(doc))
	}
	return result, nil
}

// GetScanProfileByName returns the profile with the given name, or nil if none.
func (r BasicOpsCloverRepository) GetScanProfileByName(name string) (*e.ScanProfile, error) {
	doc, err := r.db.FindFirst(q.NewQuery(scanProfilesCollection).Where(q.Field("name").Eq(name)))
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	p := docToScanProfile(doc)
	return &p, nil
}

// GetScanProfileByHost returns the profile whose host matches, or nil if none.
func (r BasicOpsCloverRepository) GetScanProfileByHost(host string) (*e.ScanProfile, error) {
	doc, err := r.db.FindFirst(q.NewQuery(scanProfilesCollection).Where(q.Field("host").Eq(host)))
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	p := docToScanProfile(doc)
	return &p, nil
}

// UpdateScanProfile overwrites the stored profile identified by name.
func (r BasicOpsCloverRepository) UpdateScanProfile(p e.ScanProfile) error {
	updates := map[string]interface{}{
		"kind":               p.Kind,
		"host":               p.Host,
		"snmp_community":     p.SNMPCommunity,
		"snmp_version":       p.SNMPVersion,
		"snmp_port":          p.SNMPPort,
		"timeout_sec":        p.TimeoutSec,
		"scan_source":        p.ScanSource,
		"config_source":      p.ConfigSource,
		"config_file":        p.ConfigFile,
		"os_type":            p.OsType,
		"ssh_user":           p.SSHUser,
		"ssh_password":       p.SSHPassword,
		"ssh_key_file":       p.SSHKeyFile,
		"ssh_key":            p.SSHKey,
		"ssh_port":           p.SSHPort,
		"discrepancy_action": p.DiscrepancyAction,
		"merge_configs":      p.MergeConfigs,
		"config_timeout":     p.ConfigTimeout,
		"vlan_accuracy":      p.VLANAccuracy,
	}
	return r.db.Update(q.NewQuery(scanProfilesCollection).Where(q.Field("name").Eq(p.Name)), updates)
}

// DeleteScanProfile removes a profile by name.
func (r BasicOpsCloverRepository) DeleteScanProfile(name string) error {
	return r.db.Delete(q.NewQuery(scanProfilesCollection).Where(q.Field("name").Eq(name)))
}
