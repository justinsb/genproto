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

# Cleanup before regeneration
mkdir -p kubee
find kubee -name generated.proto -delete
find kubee -name *.pb.go -delete

# pushd kubee
# go mod init justinsb.com/kubee
# go work use .
# echo "*.pb.go" >> .gitignore
# echo "generated.proto" >> .gitignore
# popd

pushd kubee/
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative kubee/v1/*.proto
popd

go run . -- ../apimachinery/pkg/runtime
pushd kubee/
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative apimachinery/pkg/runtime/*.proto
popd


# go run . -- ../apimachinery/pkg/api/resource
mkdir -p kubee/apimachinery/pkg/api/resource/
pushd kubee
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative apimachinery/pkg/api/resource/*.proto
popd


go run . -- ../apimachinery/pkg/util/intstr
pushd kubee
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative apimachinery/pkg/util/intstr/*.proto
popd

go run . -- ../apimachinery/pkg/apis/meta/v1
pushd kubee
protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative apimachinery/pkg/apis/meta/v1/*.proto
popd

go run . -- ../api/...
pushd kubee
protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative api/*/*/*.proto
popd