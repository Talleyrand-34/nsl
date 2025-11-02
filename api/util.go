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
	"fmt"

	q "nsl-graph/internal/repository/application"
	infra "nsl-graph/internal/repository/infra/cloverdb/base"
)

func serviceConnection(path string) (q.NetServiceInt, error) {
	repository, err := infra.NewCloverRepository(path)
	if err != nil {
		return nil, fmt.Errorf("Error creating repository: %w", err)
	}
	service := q.NewNetService(repository)
	return service, nil
}
