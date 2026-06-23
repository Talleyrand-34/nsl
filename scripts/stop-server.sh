#!/bin/sh
# scripts/stop-server.sh — stop the nsl-graph API server + PHP frontend.
#
# Usage:   ./scripts/stop-server.sh
#   API_PORT   API server port   (default: 8081)
#   WEB_PORT   PHP frontend port (default: 8091)

API_PORT="${API_PORT:-8081}"
WEB_PORT="${WEB_PORT:-8091}"

echo "Stopping nsl-graph server (:$API_PORT) and PHP frontend (:$WEB_PORT)..."
for port in "$API_PORT" "$WEB_PORT"; do
    pids=$(lsof -ti "tcp:$port" 2>/dev/null)
    if [ -n "$pids" ]; then
        kill $pids 2>/dev/null && echo "  killed pid(s) on :$port: $pids"
    fi
done
# Belt-and-braces: stop anything matching by name too.
pkill -f "nsl-graph .*server" 2>/dev/null
pkill -f "main.go server" 2>/dev/null
pkill -f "php -S .*:$WEB_PORT" 2>/dev/null
echo "Done."
