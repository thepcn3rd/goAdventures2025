#!/bin/bash
projectName="displayCert"
bin="$projectName.bin"
exe="$projectName.exe"
if [ ! -e "go.mod" ]; then
	go mod init $projectName
fi

go env -w GOPATH=`pwd`
go env -w GO111MODULE='auto'

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $bin -ldflags "-w -s" main.go 
#GOOS=windows GOARCH=amd64 go build -o $exe -ldflags "-w -s" main.go

