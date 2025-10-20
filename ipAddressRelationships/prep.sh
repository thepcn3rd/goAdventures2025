#!/bin/bash
projectName="ipMap"
bin="$projectName.bin"
exe="$projectName.exe"
if [ ! -e "go.mod" ]; then
	go mod init $projectName
fi

# Make sure you change the below path to a valid path
go env -w GOPATH=`pwd`
go env -w GO111MODULE='auto'

# Install Dependencies
go get github.com/neo4j/neo4j-go-driver/v5

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $bin -ldflags "-w -s" main.go
#GOOS=windows GOARCH=amd64 go build -o $exe -ldflags "-w -s" main.go
