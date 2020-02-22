#!/bin/bash
set -x
set -e

pushd $HOME/go/src/github.com/tinygo-org/tinygo
go install -v
popd

TINYGO=tinygo
#TINYGO=$HOME/go/bin/tinygo

time $TINYGO build -o main.wasm -target wasm -no-debug ./main.go
