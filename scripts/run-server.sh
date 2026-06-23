#!/bin/sh
# scripts/run-server.sh — (re)start the nsl-graph API server + PHP frontend.
# Rerunning first stops any previous instances (by port and by name), rebuilds,
# and relaunches, so you always run the latest binary.
#
# Usage:   ./scripts/run-server.sh [db]
#   db         database dir/file (default: test.db; or set DB=...)
#   API_PORT   API server port   (default: 8081)
#   WEB_PORT   PHP frontend port (default: 8091)

cd "$(dirname "$0")/.." || exit 1

DB="${1:-${DB:-test.db}}"
API_PORT="${API_PORT:-8081}"
WEB_PORT="${WEB_PORT:-8091}"

echo "Stopping previous instances..."
# Free the ports — robust against go-run wrappers / stale binaries.
for port in "$API_PORT" "$WEB_PORT"; do
    pids=$(lsof -ti "tcp:$port" 2>/dev/null)
    [ -n "$pids" ] && kill $pids 2>/dev/null
done
# Belt-and-braces: stop anything matching by name too.
pkill -f "nsl-graph .*server" 2>/dev/null
pkill -f "main.go server" 2>/dev/null
pkill -f "php -S .*:$WEB_PORT" 2>/dev/null
sleep 1

echo "Building nsl-graph..."
go build -o nsl-graph main.go || exit 1

echo "Starting API on :$API_PORT (db=$DB) and web on :$WEB_PORT ..."
./nsl-graph -s "$DB" server --port "$API_PORT" &
php -S "localhost:$WEB_PORT" -t frontend/php &

sleep 1
echo
echo "API:  http://localhost:$API_PORT"
echo "Web:  http://localhost:$WEB_PORT/main.php   (Scan connections: /connections.php)"
echo "DB:   $DB"
echo "Re-run this script to restart; it will kill these instances first."
