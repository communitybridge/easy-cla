#! /bin/bash

# This script updates the SalesForce parameters for a given environment. Only
# parameters provided in the list below are updated.

INSTANCE_URL=''
USERNAME=''
PASSWORD=''
SECURITY_TOKEN=''
CONSUMER_KEY=''
CONSUMER_SECRET=''

ENV=''

if [ -z "$ENV" ]; then
    echo "ERROR: missing environment"
    exit 1
fi

if [ -n "$INSTANCE_URL" ]; then
    echo "updating instance url: $INSTANCE_URL"
    aws ssm put-parameter --profile lf-cla --region us-east-1 --name "sf-instance-url-$ENV" --description "SalesForce instance URL" --value "$INSTANCE_URL" --type "String"
fi

if [ -n "$USERNAME" ]; then
    echo "updating username: $USERNAME"
    aws ssm put-parameter --profile lf-cla --region us-east-1 --name "sf-username-$ENV" --description "SalesForce user name" --value "$USERNAME" --type "String"
fi

# The SalesForce API password is the user password concatenated with a security token.
if [ -n "$PASSWORD$SECURITY_TOKEN" ]; then
    echo "updating password: $PASSWORD$SECURITY_TOKEN"
    aws ssm put-parameter --profile lf-cla --region us-east-1 --name "sf-password-$ENV" --description "SalesForce password. Combined user password and secret token" --value "$PASSWORD$SECURITY_TOKEN" --type "String"
fi

if [ -n "$CONSUMER_KEY" ]; then
    echo "updating consumer key: $CONSUMER_KEY"
    aws ssm put-parameter --profile lf-cla --region us-east-1 --name "sf-consumer-key-$ENV" --description "SalesForce Connected App Consumer Key" --value "$CONSUMER_KEY" --type "String"
fi

if [ -n "$CONSUMER_SECRET" ]; then
    echo "updating consumer secret: $CONSUMER_SECRET"
    aws ssm put-parameter --profile lf-cla --region us-east-1 --name "sf-consumer-secret-$ENV" --description "SalesForce Connected App Consumer Secret" --value "$CONSUMER_SECRET" --type "String"
fi
