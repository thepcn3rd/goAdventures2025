#!/bin/bash
projectName="toolsAPI"
bin="$projectName.bin"
exe="$projectName.exe"
if [ ! -e "go.mod" ]; then
	go mod init $projectName
fi

# Make sure you change the below path to a valid path
go env -w GOPATH="YOUR PATH HERE/toolsAPI"
go env -w GO111MODULE='auto'


GOOS=linux GOARCH=amd64 go build -o $bin -ldflags "-w -s" main.go
#GOOS=windows GOARCH=amd64 go build -o $exe -ldflags "-w -s" main.go
