#!/bin/bash
projectName="calcEntropyDirectory"
bin="$projectName.bin"
exe="$projectName.exe"
if [ ! -e "go.mod" ]; then
	go mod init $projectName
fi

# Make sure you change the below path to a valid path
go env -w GOPATH=`pwd`
go env -w GO111MODULE='auto'

# Install Dependencies
# The dependencies do require golang version 1.24
go get github.com/thepcn3rd/goAdvsCommonFunctions

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $bin -ldflags "-w -s" main.go
GOOS=windows GOARCH=amd64 go build -o $exe -ldflags "-w -s" main.go
