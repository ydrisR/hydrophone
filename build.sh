#!/bin/sh -eu

rm -rf dist
mkdir dist

echo "Run dep ensure"
$GOPATH/bin/dep ensure
$GOPATH/bin/dep check

# generate version number
if [ -z ${TRAVIS_TAG+x} ]; then 
    VERSION_BASE=$(git describe --abbrev=0 --tags 2> /dev/null || echo 'dblp.0.0.0')
else 
    VERSION_BASE=${TRAVIS_TAG}
fi
VERSION_SHORT_COMMIT=$(git rev-parse --short HEAD)
VERSION_FULL_COMMIT=$(git rev-parse HEAD)

GO_COMMON_PATH="github.com/tidepool-org/hydrophone/vendor/github.com/tidepool-org/go-common"
	
echo "Build hydrophone $VERSION_BASE"
go build -ldflags "-X $GO_COMMON_PATH/clients/version.ReleaseNumber=$VERSION_BASE \
    -X $GO_COMMON_PATH/clients/version.FullCommit=$VERSION_FULL_COMMIT \
    -X $GO_COMMON_PATH/clients/version.ShortCommit=$VERSION_SHORT_COMMIT" \
    -o dist/hydrophone hydrophone.go

cp env.sh dist/
cp start.sh dist/

echo "Push email templates"
rsync -av --progress templates dist/ --exclude '*.go'