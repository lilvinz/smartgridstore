#!/bin/bash

pushd ../../../btrdb-server/btrdbd
go build -v
ver=$(./btrdbd -version)
popd

docker build -t btrdb/dev-postgres:${ver} .
docker push btrdb/dev-postgres:${ver}
docker tag btrdb/dev-postgres:${ver} btrdb/dev-postgres:latest
docker push btrdb/dev-postgres:latest
