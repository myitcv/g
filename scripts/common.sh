set -eu
set -o pipefail

if [ "${BASH_VERSINFO[0]}" -lt 4  ]
then
  echo "You need at least bash-4.0 to run this script." >&2
  exit 1;
fi

source "${BASH_SOURCE%/*}/commonEnv.sh"
source "${BASH_SOURCE%/*}/requiredVars.sh"
