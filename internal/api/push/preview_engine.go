// SPDX-License-Identifier: AGPL-3.0-or-later
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
package push

import (
	"encoding/json"
	"fmt"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/push"
)

// previewPatch runs the engine.Preview against (intent, observed) decoded
// from raw JSON. It returns the rendered patch text + change count.
//
// The intent / observed maps are first marshalled to JSON then unmarshalled
// into configparser.ConfigData so the renderer accepts them. This is the
// cheapest way to honour the handler's "give me a string and I'll figure it
// out" contract without coupling the HTTP body to the engine types.
func previewPatch(req PreviewRequest) (string, int, error) {
	intent, err := mapToConfigData(req.Intent)
	if err != nil {
		return "", 0, fmt.Errorf("intent: %w", err)
	}
	observed, err := mapToConfigData(req.Observed)
	if err != nil {
		return "", 0, fmt.Errorf("observed: %w", err)
	}
	engine := push.NewDefaultEngine()
	text, err := engine.Preview(req.OS, intent, observed)
	if err != nil {
		return "", 0, err
	}
	// Count `@@ … @@` markers as the change count. Cheap & matches the
	// engine's output format. The engine returns "" on no changes.
	if text == "" {
		return "", 0, nil
	}
	count := 0
	for i := 0; i < len(text)-3; i++ {
		if text[i] == '@' && text[i+1] == '@' && text[i+2] == ' ' {
			count++
		}
	}
	return text, count, nil
}

// mapToConfigData rehydrates a generic JSON map into configparser.ConfigData.
// Used by the preview handler to accept an unstructured intent/observed body
// without coupling the HTTP wire shape to the engine types.
func mapToConfigData(in map[string]any) (*configparser.ConfigData, error) {
	if in == nil {
		return &configparser.ConfigData{}, nil
	}
	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var cd configparser.ConfigData
	if err := json.Unmarshal(b, &cd); err != nil {
		return nil, err
	}
	return &cd, nil
}
