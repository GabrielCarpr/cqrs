#!/bin/sh

set -ex

URI="postgresql://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

migrate -path /var/migrations -database $URI "$@"
