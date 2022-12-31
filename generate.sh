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

go run . -- ../apimachinery/pkg/runtime
pushd kubee/
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative apimachinery/pkg/runtime/*.proto
popd


# go run . -- ../apimachinery/pkg/api/resource
mkdir -p kubee/apimachinery/pkg/api/resource/
cat  > kubee/apimachinery/pkg/api/resource/generated.proto <<EOF
syntax = "proto2";

package k8s.io.apimachinery.pkg.api.resource;

// Package-wide variables from generator "generated".
option go_package = "k8s.io/apimachinery/pkg/api/resource";

message Quantity {
  optional string string = 1;
}
EOF
pushd kubee
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative apimachinery/pkg/api/resource/*.proto
popd


go run . -- ../apimachinery/pkg/util/intstr
pushd kubee
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative apimachinery/pkg/util/intstr/*.proto
popd

go run . -- ../apimachinery/pkg/apis/meta/v1
cat  > kubee/apimachinery/pkg/apis/meta/v1/custom.proto <<EOF
syntax = "proto3";

package k8s.io.apimachinery.pkg.apis.meta.v1;

option go_package = "justinsb.com/kubee/apimachinery/pkg/apis/meta/v1";

// Timestamp is a struct that is equivalent to Time, but intended for
// protobuf marshalling/unmarshalling. It is generated into a serialization
// that matches Time. Do not use in Go structs.
message Time {
  // Represents seconds of UTC time since Unix epoch
  // 1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to
  // 9999-12-31T23:59:59Z inclusive.
  optional int64 seconds = 1;

  // Non-negative fractions of a second at nanosecond resolution. Negative
  // second values with fractions must still have non-negative nanos values
  // that count forward in time. Must be from 0 to 999,999,999
  // inclusive. This field may be limited in precision depending on context.
  optional int32 nanos = 2;
}
EOF
pushd kubee
protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative apimachinery/pkg/apis/meta/v1/*.proto
popd

go run . -- ../api/...
pushd kubee
protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative api/core/v1/*.proto
popd

pushd kubee
protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative api/apps/v1/*.proto
popd