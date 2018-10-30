#!/bin/bash
set -e
chown postgres:postgres /var/lib/postgresql/data
chown postgres:postgres /var/lib/postgresql/backups
docker-entrypoint.sh "$@"
