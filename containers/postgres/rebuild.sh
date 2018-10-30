#!/bin/bash

docker build -t btrdb/dev-postgres:latest .
docker push btrdb/dev-postgres:latest
