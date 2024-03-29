#!/usr/bin/env bash

set -eou pipefail

# syn-protobuf.sh is a bash script to sync the protobuf
# files using ibc-proto-compiler. This script will checkout
# the protobuf files from the git versions specified in
# proto/src/prost/COSMOS_SDK_COMMIT and
# proto/src/prost/IBC_GO_TAG. If you want to sync
# the protobuf files to a newer version, modify the
# corresponding of those 2 files by specifying the commit ID
# that you wish to checkout from.

# This script should be run from the root directory of ibc-proto-rs.

# We can specify where to clone the git repositories
# for cosmos-sdk and ibc-go. By default they are cloned
# to /tmp/cosmos-sdk.git and /tmp/ibc-go.git.
# We can override this to existing directories
# that already have a clone of the repositories,
# so that there is no need to clone the entire
# repositories over and over again every time
# the script is called.

CACHE_PATH="${XDG_CACHE_HOME:-$HOME/.cache}"
#COSMOS_SDK_GIT="${COSMOS_SDK_GIT:-$CACHE_PATH/cosmos/cosmos-sdk.git}"
IBC_GO_GIT="${IBC_GO_GIT:-$CACHE_PATH/ibc-go.git}"
# IBC_GRANDPA_GIT="${IBC_GRANDPA_GIT:-$CACHE_PATH/ibc_grandpa.git}"
IBC_GRANDPA_DIR=${PWD}
echo "IBC_GRANDPA_DIR: $IBC_GRANDPA_DIR"
# COSMOS_SDK_COMMIT="$(cat src/COSMOS_SDK_COMMIT)"
IBC_GO_TAG="$(cat $IBC_GRANDPA_DIR/scripts/IBC_GO_TAG)"
# IBC_GRANDPA_COMMIT="$(cat src/IBC_GRANDPA_COMMIT)"

# echo "COSMOS_SDK_COMMIT: $COSMOS_SDK_COMMIT"
echo "IBC_GO_TAG: $IBC_GO_TAG"
# echo "IBC_GRANDPA_COMMIT: $IBC_GRANDPA_COMMIT"

# Use either --sdk-commit flag for commit ID,
# or --sdk-tag for git tag. Because we can't modify
# proto-compiler to have smart detection on that.

# if [[ "$COSMOS_SDK_COMMIT" =~ ^[a-zA-Z0-9]{40}$ ]]
# then
#     SDK_COMMIT_OPTION="--sdk-commit"
# else
#     SDK_COMMIT_OPTION="--sdk-tag"
# fi

# If the git directories does not exist, clone them as
# bare git repositories so that no local modification
# can be done there.

# if [[ ! -e "$COSMOS_SDK_GIT" ]]
# then
#     echo "Cloning cosmos-sdk source code to as bare git repository to $COSMOS_SDK_GIT"
#     git clone --mirror https://github.com/cosmos/cosmos-sdk.git "$COSMOS_SDK_GIT"
# else
#     echo "Using existing cosmos-sdk bare git repository at $COSMOS_SDK_GIT"
# fi

if [[ ! -e "$IBC_GO_GIT" ]]
then
    echo "Cloning ibc-go source code to as bare git repository to $IBC_GO_GIT"
    git clone --mirror https://github.com/cosmos/ibc-go.git "$IBC_GO_GIT"
else
    echo "Using existing ibc-go bare git repository at $IBC_GO_GIT"
fi

# if [[ ! -e "$IBC_GRANDPA_GIT" ]]
# then
#     echo "Cloning ibc-go source code to as bare git repository to $IBC_GRANDPA_GIT"
#     git clone --mirror https://github.com/octopus-network/ibc-go.git "$IBC_GRANDPA_GIT"
# else
#     echo "Using existing ibc-go bare git repository at $IBC_GRANDPA_GIT"
# fi


# Update the repositories using git fetch. This is so that
# we keep local copies of the repositories up to sync first.
# pushd "$COSMOS_SDK_GIT"
# git fetch
# popd

pushd "$IBC_GO_GIT"
git fetch
popd

# pushd "$IBC_GRANDPA_GIT"
# git fetch
# popd


# Create a new temporary directory to check out the
# actual source files from the bare git repositories.
# This is so that we do not accidentally use an unclean
# local copy of the source files to generate the protobuf.
# COSMOS_SDK_DIR=$(mktemp -d /tmp/cosmos-sdk-XXXXXXXX)

# pushd "$COSMOS_SDK_DIR"
# git clone "$COSMOS_SDK_GIT" .
# git checkout "$COSMOS_SDK_COMMIT"

# We have to name the commit as a branch because
# proto-compiler uses the branch name as the commit
# output. Otherwise it will just output HEAD
# git checkout -b "$COSMOS_SDK_COMMIT"

# cd proto
# bufbuild mod update
# bufbuild export -v -o ../proto-include
# popd

IBC_GO_DIR=$(mktemp -d /tmp/ibc-go-XXXXXXXX)
echo "IBC_GO_DIR: $IBC_GO_DIR"
pushd "$IBC_GO_DIR"
git clone "$IBC_GO_GIT" .
git checkout "$IBC_GO_TAG"
git checkout -b "$IBC_GO_TAG"

# cp protocgen.sh
mv $IBC_GO_DIR/scripts/protocgen.sh $IBC_GO_DIR/scripts/protocgen.sh.bak
cp $IBC_GRANDPA_DIR/scripts/protocgen.sh $IBC_GO_DIR/scripts/

# cp grandpa protocol buffer into ibc-go dir
cp -Rv $IBC_GRANDPA_DIR/proto/ibc/lightclients/grandpa/ $IBC_GO_DIR/proto/ibc/lightclients

# compile protocol buffer
make proto-all

# cp grandpa lc rust code into $IBC_GRANDPA_DIR
cp $IBC_GO_DIR/github.com/octopus-network/ics10-grandpa-go/grandpa/grandpa.pb.go $IBC_GRANDPA_DIR/grandpa

# cd proto
# bufbuild export -v -o ../proto-include
# popd

# # cp grandpa pb into ibc pb dir for compile
# IBC_GRANDPA_DIR=$(mktemp -d /tmp/ibc-grandpa-XXXXXXXX)

# pushd "$IBC_GRANDPA_DIR"
# git clone "$IBC_GRANDPA_GIT" .
# git checkout "$IBC_GRANDPA_COMMIT"
# git checkout -b "$IBC_GRANDPA_COMMIT"

# cp grandpa protocol buffer into ibc-go dir
# cp -Rv ./proto/ibc/lightclients/grandpa/ $IBC_GO_DIR/proto-include/ibc/lightclients
# popd

# Remove the existing generated protobuf files
# so that the newly generated code does not
# contain removed files.

#rm -rf src/prost
#mkdir -p src/prost

# cd tools/proto-compiler

# cargo build --locked

# Run the proto-compiler twice,
# once for std version with --build-tonic set to true
# and once for no-std version with --build-tonic set to false

# OUT=$(mktemp -d /tmp/prost-XXXXXXXX)

# # cargo run --locked -- compile \
# #   --sdk "$COSMOS_SDK_DIR/proto-include" \
# #   --ibc "$IBC_GO_DIR/proto-include" \
# #   --out ../../src/prost
# cargo run --locked -- compile \
#   --sdk "$COSMOS_SDK_DIR/proto-include" \
#   --ibc "$IBC_GO_DIR/proto-include" \
#   --out $OUT

# # cp grandpa lc rust code into ibc-proto-rs/src/prost
# cp $OUT/ibc.lightclients.grandpa.v1.rs ../../src/prost

# Remove the temporary checkouts of the repositories
# rm -rf "$COSMOS_SDK_DIR"
rm -rf "$IBC_GO_DIR"
# rm -rf "$IBC_GRANDPA_DIR"
# rm -rf "$OUT"