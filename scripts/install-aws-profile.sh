#!/usr/bin/env bash
set -e

# Get the profile name from our project-vars.yml file
PROFILE=$(grep -A3 'profile:' project-vars.yml | tail -n1 | awk '{ print $2}')

if [ -z "${CI}" ]; then
  echo "Install AWS profile should only be run in a containerized CI environment"
  exit 0
fi

echo "Installing Profile ${PROFILE}"
mkdir -p ~/.aws
printf "[${PROFILE}]\naws_access_key_id=${AWS_ACCESS_KEY_ID}\naws_secret_access_key=${AWS_SECRET_ACCESS_KEY}" > ~/.aws/credentials
