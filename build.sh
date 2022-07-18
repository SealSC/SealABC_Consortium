#!/bin/bash

plat_linux=$1

if [[ $plat_linux -eq "1" ]]; then
GOOS=linux GOARCH=amd64 go build SealABC.go
else
    go build SealABC.go
fi
