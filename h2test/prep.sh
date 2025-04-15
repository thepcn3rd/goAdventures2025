#!/bin/bash
projectName="h2test"
bin="$projectName.bin"
exe="$projectName.exe"
if [ ! -e "go.mod" ]; then
	go mod init $projectName
fi

go env -w GOPATH="/home/thepcn3rd/go/workspaces/2025/h2test"
go env -w GO111MODULE='auto'

# Install Dependencies
go get github.com/thepcn3rd/goAdvsCommonFunctions
go get golang.org/x/net/http2


GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $bin -ldflags "-w -s" main.go 
#GOOS=windows GOARCH=amd64 go build -o $exe -ldflags "-w -s" main.go

