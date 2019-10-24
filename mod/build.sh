#!/bin/bash

APIVERSION=0.x.x
VERSION=$(cat VERSION);
COMMIT=$(git rev-parse --short HEAD);

WRITTENVERSION=$APIVERSION'-'$VERSION'-'$COMMIT

# git diff-index --quiet HEAD --

# if [[ $? != 0 ]]
# then
#   echo "There is uncommitted code, commit first, and build again."
#   exit 1
# fi

sed "s/versionplaceholder/"$WRITTENVERSION"/g" version.template > ./version.go
sed "s/versionplaceholder/"$WRITTENVERSION"/g" module.toml.template > ./module.toml

mkdir bungkus
go build -buildmode=plugin -ldflags="-s -w" -o bungkus/user.so
cp -Rvf LICENSE CHANGELOG  module.toml schema bungkus
mv bungkus user
tar zcvvf user-$WRITTENVERSION.tar.gz user
rm -Rvf user ./module.toml
