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
	"time"

	d "github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"

	p "nsl-graph/internal/push"
)

func snapshotToDoc(s p.ConfigSnapshot) *d.Document {
	doc := d.NewDocument()
	doc.Set("snapshot_id", s.ID)
	doc.Set("device_id", s.DeviceID)
	doc.Set("os", s.OS)
	doc.Set("captured_at", s.CapturedAt.UnixNano())
	doc.Set("captured_by_run", s.CapturedByRun)
	doc.Set("parser_version", s.ParserVersion)
	doc.Set("sha256_raw", s.Sha256Raw)
	doc.Set("interfaces_count", s.Interfaces)
	doc.Set("raw_ext", s.RawExt)
	doc.Set("backup_relpath", s.BackupRelpath)
	return doc
}

func docToSnapshot(doc *d.Document) p.ConfigSnapshot {
	getStr := func(k string) string {
		if v, ok := doc.Get(k).(string); ok {
			return v
		}
		return ""
	}
	getInt := func(k string) int {
		switch n := doc.Get(k).(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		}
		return 0
	}
	getInt64 := func(k string) int64 {
		switch n := doc.Get(k).(type) {
		case int64:
			return n
		case int:
			return int64(n)
		case float64:
			return int64(n)
		}
		return 0
	}
	unixNano := getInt64("captured_at")
	var captured time.Time
	if unixNano > 0 {
		captured = time.Unix(0, unixNano)
	}
	return p.ConfigSnapshot{
		ID:             getStr("snapshot_id"),
		DeviceID:       getStr("device_id"),
		OS:             getStr("os"),
		CapturedAt:     captured,
		CapturedByRun:  getStr("captured_by_run"),
		ParserVersion:  getStr("parser_version"),
		Sha256Raw:      getStr("sha256_raw"),
		Interfaces:     getInt("interfaces_count"),
		RawExt:         getStr("raw_ext"),
		BackupRelpath: getStr("backup_relpath"),
	}
}

// SaveSnapshot inserts a new ConfigSnapshot. The ID must be unique.
func (r BasicOpsCloverRepository) SaveSnapshot(snap p.ConfigSnapshot) (string, error) {
	_, err := r.db.InsertOne(configSnapshotsCollection, snapshotToDoc(snap))
	return snap.ID, err
}

// GetSnapshot returns the snapshot with the given snapshot_id, or an error if not found.
func (r BasicOpsCloverRepository) GetSnapshot(id string) (p.ConfigSnapshot, error) {
	doc, err := r.db.FindFirst(q.NewQuery(configSnapshotsCollection).Where(q.Field("snapshot_id").Eq(id)))
	if err != nil || doc == nil {
		return p.ConfigSnapshot{}, fmt.Errorf("snapshot not found: %s", id)
	}
	return docToSnapshot(doc), nil
}

// ListSnapshots returns the newest `limit` snapshots for a device.
// If limit <= 0, returns all.
func (r BasicOpsCloverRepository) ListSnapshots(deviceID string, limit int) ([]p.ConfigSnapshot, error) {
	docs, err := r.db.FindAll(
		q.NewQuery(configSnapshotsCollection).
			Where(q.Field("device_id").Eq(deviceID)).
			Sort(q.SortOption{Field: "captured_at", Direction: -1}),
	)
	if err != nil {
		return nil, err
	}
	out := make([]p.ConfigSnapshot, 0, len(docs))
	for _, doc := range docs {
		out = append(out, docToSnapshot(doc))
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// LatestSnapshot returns the most recent snapshot for a device.
func (r BasicOpsCloverRepository) LatestSnapshot(deviceID string) (p.ConfigSnapshot, error) {
	docs, err := r.db.FindAll(
		q.NewQuery(configSnapshotsCollection).
			Where(q.Field("device_id").Eq(deviceID)).
			Sort(q.SortOption{Field: "captured_at", Direction: -1}).
			Limit(1),
	)
	if err != nil {
		return p.ConfigSnapshot{}, err
	}
	if len(docs) == 0 {
		return p.ConfigSnapshot{}, nil
	}
	return docToSnapshot(docs[0]), nil
}
