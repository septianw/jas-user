#!/bin/bash

mkdir bungkus
go build -buildmode=plugin -ldflags="-s -w" -o bungkus/user.so
cp -Rvf LICENSE CHANGELOG  module.toml schema bungkus
mv bungkus user
tar zcvvf user.tar.gz user
rm -Rvf user
