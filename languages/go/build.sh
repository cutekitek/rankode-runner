#!/bin/sh
echo "start building"
/usr/local/go/bin/go mod init builder
mv code main.go
mkdir cache 
echo "end building"
GOCACHE=/box/cache GOMAXPROCS=2 /usr/local/go/bin/go build -p 1 -o runnable ./main.go