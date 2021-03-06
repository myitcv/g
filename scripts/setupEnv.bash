# Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
# Use of this document is governed by a license found in the LICENSE document.

only_run_on_ci_server

export PROTOBUF_VERSION="3.0.0-beta-2"

export CI_CACHE_DIR=~/cache
export CI_DEPENDENCIES_DIR=$CI_CACHE_DIR/depedencies

export PROTOBUF_INSTALL_DIR=$CI_DEPENDENCIES_DIR/protobuf/$PROTOBUF_VERSION
export PROTOBUF_INCLUDE=$PROTOBUF_INSTALL_DIR

mkdir -p $CI_CACHE_DIR
mkdir -p $CI_DEPENDENCIES_DIR
mkdir -p $PROTOBUF_INSTALL_DIR
