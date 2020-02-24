# Copyright The Linux Foundation and each contributor to CommunityBridge.
# SPDX-License-Identifier: MIT

import os
import time
import requests
from requests.exceptions import RequestException

from github import BadCredentialsException, UnknownObjectException, GithubException, GithubIntegration, Github
from jose import jwt

import cla


class GitHubInstallation(object):

    @property
    def app_id(self):
        return os.environ['GH_APP_ID']

    @property
    def private_key(self):
        return os.environ['GH_APP_PRIVATE_SECRET']

    @property
    def repos(self):
        return self.api_object.get_installation(self.installation_id).get_repos()

    def __init__(self, installation_id):
        self.installation_id = installation_id

        cla.log.debug('Initializing github application - installation_id: {}, app id: {}, private key'
                      ' (minus header): {}...'.
                      format(self.installation_id, self.app_id, self.private_key[32:38]))

        try:
            integration = GithubCLAIntegration(self.app_id, self.private_key)
            auth = integration.get_access_token(self.installation_id)
            self.token = auth.token
            self.api_object = Github(self.token)
        except BadCredentialsException as e:
            cla.log.warning('BadCredentialsException connecting to Github using app_id: {}, installation id: '
                            '{}, error: {}'.format(self.app_id, self.installation_id, e))
            raise e
        except UnknownObjectException as e:
            cla.log.warning('UnknownObjectException connecting to Github using app_id: {}, installation id: '
                            '{}, error: {}'.format(self.app_id, self.installation_id, e))
            raise e
        except GithubException as e:
            cla.log.warning('GithubException connecting to Github using app_id: {}, installation id: '
                            '{}, error: {}'.format(self.app_id, self.installation_id, e))
            raise e
        except Exception as e:
            cla.log.warning('Error connecting to Github to fetch the access token using app_id: {}, installation id: '
                            '{}, error: {}'.format(self.app_id, self.installation_id, e))
            raise e

    def create_check_run(self, repository_name, data):
        """
        Function that creates a check run for unsigned users
        """
        try:
            url = 'https://api.github.com/repos/{}/check-runs'.format(repository_name)
            requests.post(
                url,
                data=data,
                headers={
                    'Content-Type': 'application/json',
                    'Authorization': 'token %s' % self.token,
                    'Accept': 'application/vnd.github.antiope-preview+json'
                }
            )

        except RequestException as err:
            cla.log.debug(err)


class GithubCLAIntegration(GithubIntegration):
    """
    Custom GithubIntegration using python-jose instead of pyjwt for token creation.
    """
    def create_jwt(self):
        """
        Overloaded to use python-jose instead of pyjwt.
        Couldn't get it working with pyjwt.
        """
        now = int(time.time())
        payload = {
            "iat": now,
            "exp": now + 60,
            "iss": self.integration_id
        }
        gh_jwt = jwt.encode(payload, self.private_key, 'RS256')
        #cla.log.debug('github jwt: {}'.format(gh_jwt))
        return gh_jwt
