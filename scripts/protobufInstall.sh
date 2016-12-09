#!/usr/bin/env bash

# Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
# Use of this document is governed by a license found in the LICENSE document.

source "${BASH_SOURCE%/*}/common.bash"

only_run_on_ci_server

DOWNLOAD_URL=https://github.com/google/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip

pushd $PROTOBUF_INSTALL_DIR > /dev/null

if [ ! -e protoc ]
then
  t=`mktemp`
  curl -sL $DOWNLOAD_URL > $t
  unzip $t
  rm $t
fi

popd > /dev/null
