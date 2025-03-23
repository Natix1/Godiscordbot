#!/bin/bash

go mod tidy
mkdir dist
rm -r ./dist/*
go build -o app
mv app dist
echo "Done"