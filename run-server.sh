#!/bin/sh

./nsl-graph server --port 8080 &
php -S localhost:8090 -t  frontend/php&
