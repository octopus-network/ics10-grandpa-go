#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"
cd proto

# set golang proxy
go env -w GOPROXY=https://goproxy.cn,direct
# print env 
go env

buf generate --template buf.gen.gogo.yaml $file

cd ..

# move proto files to the right places
# cp -r github.com/cosmos/ibc-go/v*/modules/* modules/
# rm -rf github.com

# go mod tidy
