/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	"nsl-graph/cmd"
	_ "nsl-graph/cmd/add"
	_ "nsl-graph/cmd/compare"
	_ "nsl-graph/cmd/delete"
	_ "nsl-graph/cmd/diagram"
	_ "nsl-graph/cmd/diff"
	_ "nsl-graph/cmd/export"
	_ "nsl-graph/cmd/print"
	_ "nsl-graph/cmd/push"
	_ "nsl-graph/cmd/root"
	_ "nsl-graph/cmd/scan"
	_ "nsl-graph/cmd/update"
)

func main() {
	cmd.Execute()
}
