#!/bin/sh
/usr/local/go/bin/go mod init builder
mv code main.go
mkdir cache 
GOCACHE=/box/cache GOMAXPROCS=2 /usr/local/go/bin/go build -p 1 -o prog ./main.go