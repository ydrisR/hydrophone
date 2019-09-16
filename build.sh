#!/bin/sh -eu

rm -rf dist
mkdir dist

echo "Build hydrophone"
go build -o dist/hydrophone hydrophone.go
cp env.sh dist/
cp start.sh dist/

echo "Push email templates"
rsync -av --progress templates dist/ --exclude '*.go'
