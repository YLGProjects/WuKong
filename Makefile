#
#MIT License
#
#Copyright (c) 2025 ylgeeker
#
#Permission is hereby granted, free of charge, to any person obtaining a copy
#of this software and associated documentation files (the "Software"), to deal
#in the Software without restriction, including without limitation the rights
#to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
#copies of the Software, and to permit persons to whom the Software is
#furnished to do so, subject to the following conditions:
#
#copies or substantial portions of the Software.
#
#THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
#AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
#LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
#OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
#SOFTWARE.
#

# base
PROJECT  := WuKong
PROTOC    = protoc
PROTO_DIR = pkg/proto
GEN_DIR   = pkg/proto
BUILD_DIR = build
GO_OS    ?= linux
MODULE    = YLGProjects/WuKong/pkg

# version
BUILDTIME  = $(shell date +%Y-%m-%dT%T%z)
GITTAG     = $(shell git describe --tags --always)
GITHASH    = $(shell git rev-parse --short HEAD)
VERSION   ?= $(GITTAG)-$(shell date +%y.%m.%d)

BUILD_FLAG = "-X '$(MODULE)/version.buildTime=$(BUILDTIME)' \
			 -X '$(MODULE)/version.gitTag=$(GITTAG)' \
			 -X '$(MODULE)/version.gitHash=$(GITHASH)' \
             -X '$(MODULE)/version.version=$(VERSION)' "

# flags
GO_FLAGS = --go_out=$(GEN_DIR) --go_opt=paths=source_relative --go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative

# search .proto file
PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)

# generate go code files
GO_GEN_FILES=$(PROTO_FILES:$(PROTO_DIR)/%.proto=$(GEN_DIR)/%.pb.go)

.PHONY: all proto agent controller transfer clean

# build target

all: proto agent controller transfer

proto: $(GO_GEN_FILES)

agent:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=${GO_OS} GOARCH=amd64 go build -ldflags=$(BUILD_FLAG) -gcflags="all=-trimpath=$(PWD)" \
				-asmflags="all=-trimpath=$(PWD)" -o $(BUILD_DIR)/$@ cmd/agent/*.go

controller:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=${GO_OS} GOARCH=amd64 go build -ldflags=$(BUILD_FLAG) -gcflags="all=-trimpath=$(PWD)" \
				-asmflags="all=-trimpath=$(PWD)" -o $(BUILD_DIR)/$@ cmd/controller/*.go
transfer:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=${GO_OS} GOARCH=amd64 go build -ldflags=$(BUILD_FLAG) -gcflags="all=-trimpath=$(PWD)" \
				-asmflags="all=-trimpath=$(PWD)" -o $(BUILD_DIR)/$@ cmd/transfer/*.go

# build protobuf to go
$(GEN_DIR)/%.pb.go: $(PROTO_DIR)/%.proto
	@mkdir -p $(GEN_DIR)
	$(PROTOC) $(GO_FLAGS) -I$(PROTO_DIR) $<

clean:
	rm -rf $(GEN_DIR)/*.go
	rm -rf $(BUILD_DIR)/*

