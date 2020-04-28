# Copyright The Linux Foundation and each contributor to CommunityBridge.
# SPDX-License-Identifier: MIT

import json
import os
from http import HTTPStatus
from typing import Optional

import requests

import cla
import cla.auth
from cla.models.dynamo_models import UserPermissions

stage = os.environ.get('STAGE', '')

sf_instance_url = os.environ.get('SF_INSTANCE_URL', '')
sf_client_id = os.environ.get('SF_CLIENT_ID', '')
sf_client_secret = os.environ.get('SF_CLIENT_SECRET', '')
sf_username = os.environ.get('SF_USERNAME', '')
sf_password = os.environ.get('SF_PASSWORD', '')



cla_logo_url = os.environ.get('CLA_BUCKET_LOGO_URL', '')

platform_gateway_url = os.environ.get('PLATFORM_GATEWAY_URL', '')
auth0_url = os.environ.get('PLATFORM_AUTH0_URL')
platform_client_id = os.environ.get('PLATFORM_AUTH0_CLIENT_ID')
platform_client_secret = os.environ.get('PLATFORM_AUTH0_CLIENT_SECRET')
platform_audience = os.environ.get('PLATFORM_GATEWAY_URL')



def format_response(status_code, headers, body):
    """
    Helper function: Generic response formatter
    """
    response = {
        'statusCode': status_code,
        'headers': headers,
        'body': body
    }
    return response


def format_json_cors_response(status_code, body):
    """
    Helper function: Formats json responses with cors headers
    """
    body = json.dumps(body)
    cors_headers = {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Credentials': True,
        'Access-Control-Allow-Headers': 'Content-Type, Authorization'
    }
    response = format_response(status_code, cors_headers, body)
    return response


def get_sf_oauth_access() -> Optional[dict]:
    data = {
        'grant_type': 'password',
        'client_id': sf_client_id,
        'client_secret': sf_client_secret,
        'username': sf_username,
        'password': sf_password
    }

    token_url = f'https://{sf_instance_url}/services/oauth2/token'
    r = requests.post(token_url, data=data)
    if r.status_code == requests.codes.ok:
        cla.log.debug(f'Lookup succeeded fetching oauth token by username: {sf_username}, client_id: {sf_client_id}.')
        return r.json()
    else:
        cla.log.warning(f'Error fetching oauth token by username: {sf_username}, client_id: {sf_client_id}. '
                        f'Error: {r.status_code} - {r.text}')
        return None


def get_projects(event, context):
    """
    Gets list of all projects from Salesforce
    """

    cla.log.debug('event: {}'.format(event))

    try:
        auth_user = cla.auth.authenticate_user(event.get('headers'))
    except cla.auth.AuthError as e:
        cla.log.error('Authorization error: {}'.format(e))
        return format_json_cors_response(401, 'Error parsing Bearer token')
    except Exception as e:
        cla.log.error('Unknown authorization error: {}'.format(e))
        return format_json_cors_response(401, 'Error parsing Bearer token')
    

    # Get project access list for user
    user_permissions = UserPermissions()
    try:
        user_permissions.load(auth_user.username)
    except Exception as e:
        cla.log.error('Error invalid username: {}. error: {}'.format(auth_user.username, e))
        return format_json_cors_response(400, 'Error invalid username')

    user_permissions = user_permissions.to_dict()

    authorized_projects = user_permissions.get('projects')
    if authorized_projects is None:
        cla.log.error('Error user not authorized to access projects: {}'.format(user_permissions))
        return format_json_cors_response(403, 'Error user not authorized to access projects')

    project_list = ','.join([id for id in authorized_projects])

    access_token, code = get_access_token()

    if code != HTTPStatus.OK:
        cla.log.error('Authentication failure')
        return format_json_cors_response(code, 'Authentication failure')

    headers = {
        'Authorization': f'bearer {access_token}',
        'accept': 'application/json'
    }

    query_url = f'{platform_gateway_url}/project-service/v1/projects/search?id={project_list}'
    resp = requests.get(query_url, headers=headers)

    response = json.loads(resp.text)
    status_code = resp.status_code
    if status_code != HTTPStatus.OK:
        cla.log.error('Error retrieving projects: %s', response[0].get('message'))
        return format_json_cors_response(status_code, 'Error retrieving projects')
    records = response.get('records')

    projects = []
    for project in records:
        logo_url = None
        project_id = project.get('Id')
        if project_id:
            logo_url = '{}/{}.png'.format(cla_logo_url, project_id)

        projects.append({
            'name': project.get('Name'),
            'id': project_id,
            'description': project.get('Description__c'),
            'logoUrl': logo_url
        })

    return format_json_cors_response(status_code, projects)

def get_access_token():
    """
    Get token access token for platform service
    """
    auth0_payload = {
        'grant_type': 'client_credentials',
        'client_id': platform_client_id,
        'client_secret': platform_client_secret,
        'audience': platform_audience
    }

    headers = {
        'content-type': 'application/x-www-form-urlencoded',
        'accept': 'application/json'
    }
    
    access_token = ''
    try:
        # cla.log.debug(f'Sending POST to {auth0_url} with payload: {auth0_payload}')
        cla.log.debug(f'Sending POST to {auth0_url}')
        resp = requests.post(auth0_url, data=auth0_payload, headers=headers)
        status_code = resp.status_code
        if status_code != HTTPStatus.OK:
            cla.log.error('Forbidden: %s', resp.raise_for_status())
        json_data = json.loads(resp.text)
        access_token = json_data["access_token"]
        return access_token, status_code
    except requests.exceptions.HTTPError as err:
        msg = f'Could not get auth token, error: {err}'
        cla.log.warning(msg)
        return None, err.response.status_code

def get_project(event, context):
    """
    Given project id, gets project details from Salesforce
    """

    cla.log.info('event: {}'.format(event))

    project_id = event.get('queryStringParameters').get('id')
    if project_id is None:
        return format_json_cors_response(400, 'Missing project ID')

    try:
        auth_user = cla.auth.authenticate_user(event.get('headers'))
    except cla.auth.AuthError as e:
        cla.log.error('Authorization error: {}'.format(e))
        return format_json_cors_response(401, 'Error parsing Bearer token')
    except Exception as e:
        cla.log.error('Unknown authorization error: {}'.format(e))
        return format_json_cors_response(401, 'Error parsing Bearer token')

    # Get project access list for user
    user_permissions = UserPermissions()
    try:
        user_permissions.load(auth_user.username)
    except:
        cla.log.error('Error invalid username: {}'.format(auth_user.username))
        return format_json_cors_response(400, 'Error invalid username')

    user_permissions = user_permissions.to_dict()

    authorized_projects = user_permissions.get('projects')
    if authorized_projects is None:
        cla.log.error('Error user not authorized to access projects: {}'.format(user_permissions))
        return format_json_cors_response(403, 'Error user not authorized to access projects')

    if project_id not in authorized_projects:
        cla.log.error('Error user not authorized')
        return format_json_cors_response(403, 'Error user not authorized')

    token, code = get_access_token()

    if code != HTTPStatus.OK:
        cla.log.error('Authentication failure')
        return format_json_cors_response(code, 'Authentication failure')

    headers = {
        'Authorization': 'Bearer {}'.format(token)
    }

    url = f'{platform_gateway_url}/project-service/v1/projects/search?id={project_id}'

    cla.log.info('Using Project service to get project info..')
    resp = requests.get(url, headers=headers)

    response = resp.json()
    status_code = resp.status_code
    if status_code != HTTPStatus.OK:
        cla.log.error('Error retrieving project: %s', response[0].get('message'))
        return format_json_cors_response(status_code, 'Error retrieving project')

    logo_url = None
    if response.get('id'):
        logo_url = '{}/{}.png'.format(cla_logo_url, response.get('id'))

    project = {
        'name': response.get('Name'),
        'id': response.get('Id'),
        'description': response.get('Description__c'),
        'logoUrl': logo_url
    }
    return format_json_cors_response(status_code, project)
