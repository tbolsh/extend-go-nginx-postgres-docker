#!/bin/sh
#$(aws ecr get-login --no-include-email --region us-west-2)

go mod tidy
if [ $? -ne 0 ]; then
    exit
fi

gofmt -s -w src/*.go
# src/*/*.go

go vet src/extend-api-service.go
if [ $? -ne 0 ]; then
    exit
fi

golint src/*.go
if [ $? -ne 0 ]; then
    exit
fi

go test -v github.com/tbolsh/extend-go-nginx-postgres-docker/genericjson
if [ $? -ne 0 ]; then
    exit
fi

OOS=linux GOARCH=amd64 go build -o extend-api-service src/extend-api-service.go
if [ $? -ne 0 ]; then
    exit
fi

echo `cat src/version;date +".%y%m%d.%H%M"`| tr -d ' ' > version
docker build --compress -t extend-api-service:latest .
if [ $? -ne 0 ]; then
    exit
fi

#docker tag c20-service:latest 486916610627.dkr.ecr.us-west-2.amazonaws.com/c20-service:latest
#docker push 486916610627.dkr.ecr.us-west-2.amazonaws.com/c20-service:latest

docker-compose build
docker-compose up -d
