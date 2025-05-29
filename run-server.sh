#!/bin/sh

# ./nsl-graph server --port 8080 &
go run main.go server --port 8081 &
php -S localhost:8091 -t  frontend/php&
