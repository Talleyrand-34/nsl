When error on getid due to multiple options/name collsion offer possible ids


Model correctly json

Internet DONE
  - Multiple connections over a port
  - special case for internet multiple connections over a port


Vlans
  - colour code simple
  - multiple lines one vlan per line


Allow multiple ips


VLAN on port if there is only one vlan
  vlan untagged

Button add all deviceports on a device
or select ports to add


<!-- Select backend ip/port -->

Device type hierarchy


Forbid vlan configuration of certain device types

Diagram
  Plugin for order configuration
  Solve colour in vlan port all red


## Observations (code review)

Cleanup / SQL -> CloverDB migration leftovers  [DONE]
  - [x] entities/datastruct.go comments updated (CloverDB, not SQL)
  - [x] removed go-sqlite3 + modernc.org/sqlite from go.mod (go mod tidy)
  - [x] removed dead SQLite code path in cmd/utils/utils.go and -b sqlite option
  - [x] removed broken SQLC `generate`/`db-generate` Makefile targets

Docs consistency
  - README references ./run-server.sh but Makefile already has run-fullstack; confirm the script exists or update README

Project hardening
  - project.md future lines (config generation, discovery beacons, config auditor) -> use for TFG "future work"


