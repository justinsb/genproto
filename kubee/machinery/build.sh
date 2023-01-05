#!/bin/bash

set -e
set -x

pushd ../..

mkdir -p ../bin
pushd ../bin
export PATH=`pwd`:$PATH
popd

pushd ../protobuf-go
go build -o ../bin/protoc-gen-go ./cmd/protoc-gen-go
popd

popd

protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative apis/*/*/*.proto
