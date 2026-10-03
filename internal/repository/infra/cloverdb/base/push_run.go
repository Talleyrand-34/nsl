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
	"time"

	d "github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"

	"nsl-graph/internal/configparser"
	p "nsl-graph/internal/push"
)

func pushRunToDoc(r p.PushRun) *d.Document {
	doc := d.NewDocument()
	doc.Set("run_id", r.ID)
	doc.Set("device_id", r.DeviceID)
	doc.Set("os", r.OS)
	doc.Set("safety", int(r.Safety))
	doc.Set("renderer", r.Renderer)
	doc.Set("started_at", r.StartedAt.Unix())
	doc.Set("finished_at", r.FinishedAt.Unix())
	doc.Set("backup_path", r.BackupPath)
	doc.Set("snapshot_id", r.SnapshotID)
	doc.Set("backup_relpath", r.BackupRelpath)
	doc.Set("exit_status", r.ExitStatus)
	doc.Set("error_string", r.ErrorString)
	return doc
}

func docToPushRun(doc *d.Document) p.PushRun {
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
	unixStart := getInt64("started_at")
	unixFinish := getInt64("finished_at")
	var started, finished time.Time
	if unixStart > 0 {
		started = time.Unix(unixStart, 0)
	}
	if unixFinish > 0 {
		finished = time.Unix(unixFinish, 0)
	}
	return p.PushRun{
		ID:            getStr("run_id"),
		DeviceID:      getStr("device_id"),
		OS:            getStr("os"),
		Safety:        configparser.SafetyLevel(getInt("safety")),
		Renderer:      getStr("renderer"),
		StartedAt:     started,
		FinishedAt:    finished,
		BackupPath:    getStr("backup_path"),
		SnapshotID:    getStr("snapshot_id"),
		BackupRelpath: getStr("backup_relpath"),
		ExitStatus:    getStr("exit_status"),
		ErrorString:   getStr("error_string"),
	}
}

func (r BasicOpsCloverRepository) AppendPushRun(run p.PushRun) error {
	_, err := r.db.InsertOne(pushRunsCollection, pushRunToDoc(run))
	return err
}

func (r BasicOpsCloverRepository) AllPushRuns() ([]p.PushRun, error) {
	docs, err := r.db.FindAll(
		q.NewQuery(pushRunsCollection).
			Sort(q.SortOption{Field: "started_at", Direction: -1}),
	)
	if err != nil {
		return nil, err
	}
	out := make([]p.PushRun, 0, len(docs))
	for _, doc := range docs {
		out = append(out, docToPushRun(doc))
	}
	return out, nil
}
