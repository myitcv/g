# Copyright (c) 2016 Paul Jolly <paul@myitcv.org.uk>, all rights reserved.
# Use of this document is governed by a license found in the LICENSE document.

set -eu
set -o pipefail

if [ "${BASH_VERSINFO[0]}" -lt 4  ]
then
  echo "You need at least bash-4.0 to run this script." >&2
  exit 1;
fi

source "${BASH_SOURCE%/*}/commonEnv.bash"
source "${BASH_SOURCE%/*}/requiredVars.bash"
