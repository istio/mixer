#!/bin/bash

echo "Installing or updating linters"
go get -u gopkg.in/alecthomas/gometalinter.v1
go get -u github.com/bazelbuild/buildifier/buildifier
go get -u github.com/3rf/codecoroner
gometalinter.v1 --install --vendored-linters >/dev/null

echo Done installing linters
