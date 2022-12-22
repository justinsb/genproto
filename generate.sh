#!/bin/bash

set -e
set -x

rm -rf generated/

go run . -- ../apimachinery/pkg/runtime
pushd generated
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative k8s.io/apimachinery/pkg/runtime/*.proto
popd


# go run . -- ../apimachinery/pkg/api/resource
mkdir -p generated/k8s.io/apimachinery/pkg/api/resource/
cat  > generated/k8s.io/apimachinery/pkg/api/resource/generated.proto <<EOF
syntax = "proto2";

package k8s.io.apimachinery.pkg.api.resource;

// Package-wide variables from generator "generated".
option go_package = "k8s.io/apimachinery/pkg/api/resource";

message Quantity {
  optional string string = 1;
}
EOF
pushd generated
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative k8s.io/apimachinery/pkg/api/resource/*.proto
popd


go run . -- ../apimachinery/pkg/util/intstr
pushd generated
protoc --experimental_allow_proto3_optional --go_out=. --go_opt=paths=source_relative k8s.io/apimachinery/pkg/util/intstr/*.proto
popd

go run . -- ../apimachinery/pkg/apis/meta/v1
pushd generated
protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative k8s.io/apimachinery/pkg/apis/meta/v1/*.proto
popd

go run . -- ../api/core/v1
pushd generated
protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative k8s.io/api/core/v1/*.proto
popd

go run . -- ../api/apps/v1
pushd generated
protoc --experimental_allow_proto3_optional -I. --go_out=. --go_opt=paths=source_relative k8s.io/api/apps/v1/*.proto
popd