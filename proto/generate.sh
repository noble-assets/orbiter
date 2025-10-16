#!/bin/bash

# NOTE: this file should be executed from the project root.

rm -rf api/
cd ./proto || exit 1
buf generate --template buf.gen.gogo.yaml
buf generate --template buf.gen.pulsar.yaml
cd ..

cp -r github.com/noble-assets/orbiter/* ./
cp -r api/noble/orbiter/* api/
find api/ -type f -name "*.go" -exec sed -i 's|github.com/noble-assets/orbiter/api/noble/orbiter|github.com/noble-assets/orbiter/api|g' {} +

rm -rf github.com/noble-assets/orbiter
rm -rf api/noble
rm -rf noble
