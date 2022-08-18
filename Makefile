# Makefile for building CoreDNS
GITCOMMIT:=$(shell git describe --dirty --always)
BINARY:=coredns
SYSTEM:=
CHECKS:=check
BUILDOPTS:=-v
GOPATH?=$(HOME)/go
GO=$(GOPATH)/bin/go1.17
MAKEPWD:=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))
CGO_ENABLED:=0

.PHONY: all
all: coredns

.PHONY: coredns
coredns: $(CHECKS)
	CGO_ENABLED=$(CGO_ENABLED) $(SYSTEM) $(GO) build $(BUILDOPTS) -ldflags="-s -w -X github.com/coredns/coredns/coremain.GitCommit=$(GITCOMMIT)" -o $(BINARY)

.PHONY: check
check: core/plugin/zplugin.go core/dnsserver/zdirectives.go

core/plugin/zplugin.go core/dnsserver/zdirectives.go: plugin.cfg
	$(GO) generate coredns.go
	$(GO) get

.PHONY: gen
gen:
	$(GO) generate coredns.go
	$(GO) get

.PHONY: pb
pb:
	$(MAKE) -C pb

.PHONY: clean
clean:
	$(GO) clean
	rm -f coredns
