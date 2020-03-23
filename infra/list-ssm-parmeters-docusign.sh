#!/usr/bin/env bash
# Copyright The Linux Foundation and each contributor to CommunityBridge.
# SPDX-License-Identifier: MIT
set -o nounset -o pipefail
declare -r SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

#------------------------------------------------------------------------------
# Load helper scripts
#------------------------------------------------------------------------------
if [[ -f "${SCRIPT_DIR}/colors.sh" ]]; then
  source "${SCRIPT_DIR}/colors.sh"
else
  echo "Unable to load script: ${SCRIPT_DIR}/colors.sh"
  exit 1
fi

if [[ -f "${SCRIPT_DIR}/logger.sh" ]]; then
  source "${SCRIPT_DIR}/logger.sh"
else
  echo "Unable to load script: ${SCRIPT_DIR}/logger.sh"
  exit 1
fi

if [[ -f "${SCRIPT_DIR}/utils.sh" ]]; then
  source "${SCRIPT_DIR}/utils.sh"
else
  echo "Unable to load script: ${SCRIPT_DIR}/utils.sh"
  exit 1
fi

#------------------------------------------------------------------------------
# Check command line arguments
#------------------------------------------------------------------------------
if [[ $# -eq 0 ]]; then
  echo "Missing environment parameter. Expecting one of: 'dev', 'staging', or 'prod'."
  echo "usage:   $0 [environment]"
  echo "example: $0 dev"
  echo "example: $0 staging"
  echo "example: $0 prod"
  exit 1
fi

declare -r env="${1}"
if [[ "${env}" == 'dev' || "${env}" == 'staging' || "${env}" == 'prod' ]]; then
  echo "Using environment '${env}'..."
else
  echo "Environment parameter does not match expected values. Expecting one of: 'dev', 'staging', or 'prod'."
  echo "usage:   $0 [environment]"
  echo "example: $0 dev"
  echo "example: $0 staging"
  echo "example: $0 prod"
  exit 1
fi

#------------------------------------------------------------------------------
# Common variables
#------------------------------------------------------------------------------
declare -r region="us-east-1"
declare -r profile="lfproduct-${env}"
declare -a parameters=("cla-docusign-root-url-${env}"
  "cla-docusign-username-${env}"
  "cla-docusign-password-${env}"
  "cla-docusign-integrator-key-${env}"
  )

#------------------------------------------------------------------------------
# Show parameter function
#------------------------------------------------------------------------------
function show-parameter() {
  aws --profile "${profile}" ssm get-parameter --name "${1}" | jq -r '.Parameter | .Value'
}

#------------------------------------------------------------------------------
# Main function
#------------------------------------------------------------------------------
function main() {
  for parameter in "${parameters[@]}"; do
    log "Feching parameter: ${_Y}${parameter}${_W}..."
    show-parameter "${parameter}"
  done
}

#------------------------------------------------------------------------------
# Call the main function
#------------------------------------------------------------------------------
main
