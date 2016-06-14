#!/usr/bin/env bash

source "${BASH_SOURCE%/*}/common.sh"

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
