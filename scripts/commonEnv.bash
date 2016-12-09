# Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
# Use of this document is governed by a license found in the LICENSE document.

set -e
shopt -s globstar
shopt -s extglob

ensure_clean()
{
  local dir
  dir=$(pwd)
  if [ $# -eq 1 ]
  then
    dir=$1
  fi

  pushd $dir > /dev/null

  local output
  output=$(git status --porcelain)

  if [ -z "$output"  ]
  then
    echo "Git is clean"
  else
    >&2 echo -e "Git is not clean. The following files should have been committed:\n\n$output"
    exit 1
  fi

  popd > /dev/null
}

export -f ensure_clean

only_run_on_ci_server()
{
  if [ $(running_on_ci_server) != "yes" ]
  then
    echo "This script can ONLY be run on the CI server"
    exit 1
  fi
}

export -f only_run_on_ci_server

running_on_ci_server()
{
  set +u
  local res
  if [ "$TRAVIS" == "true" ]
  then
    res=yes
  else
    res=no
  fi
  set -u
  echo $res
}

export -f running_on_ci_server
