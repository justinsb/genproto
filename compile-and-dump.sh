#!/bin/bash

set -e
set -x

mkdir -p ../bin
pushd ../bin
export PATH=`pwd`:$PATH
popd

pushd ../protobuf-go
go build -o ../bin/protoc-gen-go ./cmd/protoc-gen-go
popd



pushd kubee
protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative api/apps/v1/*.proto
# protoc --experimental_allow_proto3_optional -I. -oapps.desc api/apps/v1/*.proto
# protoc --decode=google.protobuf.FileDescriptorSet google/protobuf/descriptor.proto < apps.desc

protoc --experimental_allow_proto3_optional -I. -o/dev/stdout api/apps/v1/*.proto | protoc --decode=google.protobuf.FileDescriptorSet google/protobuf/descriptor.proto

popd