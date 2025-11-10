#!/bin/bash -eux

# Build the application
pushd pull_request
  make build
  cp build/dis-redirect-proxy Dockerfile.concourse ../build
popd
