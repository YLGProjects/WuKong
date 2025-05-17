#!/bin/sh
# install protobuf compiler
apt install -y protobuf-compiler

# install grpc tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
