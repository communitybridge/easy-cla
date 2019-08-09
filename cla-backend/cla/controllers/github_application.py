# Copyright The Linux Foundation and each contributor to CommunityBridge.
# SPDX-License-Identifier: MIT

import os
import time

from github import GithubIntegration, Github
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

        cla.log.debug('github installation_id: {}, app id: {}, private key: {}'.
                      format(self.installation_id, self.app_id, self.private_key[:40]))

        try:
            integration = GithubCLAIntegration(self.app_id, self.private_key)
            auth = integration.get_access_token(self.installation_id)

            # cla.log.debug('github access token: {}'.format(auth))

            self.token = auth.token
            self.api_object = Github(self.token)
        except Exception as e:
            cla.log.warning('Error connecting to Github to fetch the access token using app_id: {}, installation id: '
                            '{}, error: {}'.format(self.app_id, self.installation_id, e))
            raise e

        cla.log.info("Initializing Github Application")


class GithubCLAIntegration(GithubIntegration):
    """Custom GithubIntegration using python-jose instead of pyjwt for token creation."""

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
        cla.log.debug('github jwt: {}'.format(gh_jwt))

        return gh_jwt
