#!/usr/bin/env bash

# Copyright The Linux Foundation and each contributor to CommunityBridge.
# SPDX-License-Identifier=MIT

# In a separate terminal, you can then locally invoke the function using cURL:
# curl -XPOST "http://localhost:8080/2015-03-31/functions/function/invocations" -d '{"payload":"hello world!"}'

podman run \
  --rm \
  -it \
  --name easycla-python \
  -p 8080:8080 \
  -e STAGE="${STAGE}" \
  -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
  -e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
  -e AWS_SESSION_TOKEN="${AWS_SESSION_TOKEN}" \
  -e REGION="us-east-1" \
  -e DYNAMODB_AWS_REGION="us-east-1" \
  -e GH_APP_WEBHOOK_SECRET="${GH_APP_WEBHOOK_SECRET}" \
  -e GH_APP_ID="${GH_APP_ID}" \
  -e GH_OAUTH_CLIENT_ID="${GH_OAUTH_CLIENT_ID}" \
  -e GH_OAUTH_SECRET="${GH_OAUTH_SECRET}" \
  -e GITHUB_OAUTH_TOKEN="${GITHUB_OAUTH_TOKEN}" \
  -e GITHUB_APP_WEBHOOK_SECRET="${GITHUB_APP_WEBHOOK_SECRET}" \
  -e GH_STATUS_CTX_NAME="EasyCLA" \
  -e AUTH0_DOMAIN="${AUTH0_DOMAIN}" \
  -e AUTH0_CLIENT_ID="${AUTH0_CLIENT_ID}" \
  -e AUTH0_USERNAME_CLAIM="${AUTH0_USERNAME_CLAIM}" \
  -e AUTH0_ALGORITHM="${AUTH0_ALGORITHM}" \
  -e SF_INSTANCE_URL="${SF_INSTANCE_URL}" \
  -e SF_CLIENT_ID="${SF_CLIENT_ID}" \
  -e SF_CLIENT_SECRET="${SF_CLIENT_SECRET}" \
  -e SF_USERNAME="${SF_USERNAME}" \
  -e SF_PASSWORD="${SF_PASSWORD}" \
  -e DOCRAPTOR_API_KEY="${DOCRAPTOR_API_KEY}" \
  -e DOCUSIGN_ROOT_URL="${DOCUSIGN_ROOT_URL}" \
  -e DOCUSIGN_USERNAME="${DOCUSIGN_USERNAME}" \
  -e DOCUSIGN_PASSWORD="${DOCUSIGN_PASSWORD}" \
  -e DOCUSIGN_AUTH_SERVER="${DOCUSIGN_AUTH_SERVER}" \
  -e CLA_API_BASE="${CLA_API_BASE}" \
  -e CLA_CONTRIBUTOR_BASE="${CLA_CONTRIBUTOR_BASE}" \
  -e CLA_CONTRIBUTOR_V2_BASE="${CLA_CONTRIBUTOR_V2_BASE}" \
  -e CLA_CORPORATE_BASE="${CLA_CORPORATE_BASE}" \
  -e CLA_CORPORATE_V2_BASE="${CLA_CORPORATE_V2_BASE}" \
  -e CLA_LANDING_PAGE="${CLA_LANDING_PAGE}" \
  -e CLA_SIGNATURE_FILES_BUCKET="${CLA_SIGNATURE_FILES_BUCKET}" \
  -e CLA_BUCKET_LOGO_URL="${CLA_BUCKET_LOGO_URL}" \
  -e SES_SENDER_EMAIL_ADDRESS="${SES_SENDER_EMAIL_ADDRESS}" \
  -e SMTP_SENDER_EMAIL_ADDRESS="${SMTP_SENDER_EMAIL_ADDRESS}" \
  -e LF_GROUP_CLIENT_ID="${LF_GROUP_CLIENT_ID}" \
  -e LF_GROUP_CLIENT_SECRET="${LF_GROUP_CLIENT_SECRET}" \
  -e LF_GROUP_REFRESH_TOKEN="${LF_GROUP_REFRESH_TOKEN}" \
  -e LF_GROUP_CLIENT_URL="${LF_GROUP_CLIENT_URL}" \
  -e SNS_EVENT_TOPIC_ARN="${SNS_EVENT_TOPIC_ARN}" \
  -e PLATFORM_AUTH0_URL="${PLATFORM_AUTH0_URL}" \
  -e PLATFORM_AUTH0_CLIENT_ID="${PLATFORM_AUTH0_CLIENT_ID}" \
  -e PLATFORM_AUTH0_CLIENT_SECRET="${PLATFORM_AUTH0_CLIENT_SECRET}" \
  -e PLATFORM_AUTH0_AUDIENCE="${PLATFORM_AUTH0_AUDIENCE}" \
  -e PLATFORM_GATEWAY_URL="${PLATFORM_GATEWAY_URL}" \
  -e PLATFORM_MAINTAINERS="${PLATFORM_MAINTAINERS}" \
  easycla-python:latest
