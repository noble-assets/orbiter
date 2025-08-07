#!/bin/bash

# NOTE: this file should be executed from the project root.

cd ./proto || exit 1
buf generate --template buf.gen.gogo.yaml
buf generate --template buf.gen.pulsar.yaml
cd ..

cp -r orbiter.dev/* ./
cp -r api/noble/orbiter/* api/
find api/ -type f -name "*.go" -exec sed -i 's|orbiter.dev/api/noble/orbiter|orbiter.dev/api|g' {} +

rm -rf orbiter.dev
rm -rf api/noble
rm -rf noble
