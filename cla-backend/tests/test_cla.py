"""
Base test case for CLA system.
"""
# Ensure the CLA app is in our path for tests.
import os
import sys
TESTDIR = os.path.abspath(os.path.dirname(__file__))
sys.path.insert(0, TESTDIR + '/..')

# Ensure no debug messages from the boto package (and others) don't appear.
import logging
logging.disable(logging.ERROR)

import unittest
import uuid
import hug
import cla

# For now, force DynamoDB for testing (should probably be SQlite :memory: in the future).
cla.conf['DATABASE'] = 'DynamoDB'
cla.conf['DATABASE_HOST'] = 'http://dynamodb:8000'
# Use a mock signing service for testing.
cla.conf['SIGNING_SERVICE'] = 'MockDocuSign'
cla.conf['BASE_URL'] = 'http://cla-system.com'
# Use a mock email service for testing.
cla.conf['EMAIL_SERVICE'] = 'MockSMTP'
# Use exclusively local storage for tests.
cla.conf['STORAGE_SERVICE'] = 'LocalStorage'
cla.conf['GITHUB_APP_PRIVATE_KEY_PATH'] = '../' + cla.conf['GITHUB_APP_PRIVATE_KEY_PATH']

from cla import routes
from cla.utils import create_database, delete_database

class CLATestCase(unittest.TestCase):
    def setUp(self):
        """Setup method executes before every test."""
        create_database()

    def create_user(self,
                    user_email='user@email.com',
                    user_name='User Name',
                    user_company_id=None,
                    user_external_id=None,
                    user_github_id=12345):
        """Helper method to create a user."""
        data = {'user_email': user_email,
                'user_name': user_name,
                'user_github_id': user_github_id}
        if user_company_id is not None:
            data['user_company_id'] = user_company_id
        response = hug.test.post(routes, '/v1/user', data)
        return response.data

    def create_signature(self,
                         signature_project_id,
                         signature_reference_id,
                         signature_reference_type,
                         signature_type='cla',
                         signature_user_ccla_company_id=None,
                         signature_signed=True,
                         signature_approved=True,
                         signature_return_url='http://test-github.com/user/repo/1',
                         signature_sign_url='http://link-to-signature.com/sign-here'):
        """Helper method to create an signatures."""
        data = {'signature_project_id': signature_project_id,
                'signature_reference_id': signature_reference_id,
                'signature_reference_type': signature_reference_type,
                'signature_type': signature_type,
                'signature_user_ccla_company_id': signature_user_ccla_company_id,
                'signature_signed': signature_signed,
                'signature_approved': signature_approved,
                'signature_return_url': signature_return_url,
                'signature_sign_url': signature_sign_url}
        response = hug.test.post(routes, '/v1/signature', data)
        return response.data

    def create_repository(self, # pylint: disable=too-many-arguments
                          repository_project_id,
                          repository_name='Repo Name',
                          repository_type='mock_github',
                          repository_url='https://some-github-url.com/repo-name',
                          repository_external_id=1):
        """Helper method to create a repository."""
        data = {'repository_project_id': repository_project_id,
                'repository_name': repository_name,
                'repository_type': repository_type,
                'repository_url': repository_url,
                'repository_external_id': repository_external_id}
        response = hug.test.post(routes, '/v1/repository', data)
        return response.data

    def create_company(self,
                       company_manager_id=None,
                       company_name='Org Name',
                       company_whitelist=['whitelisted@safe.org'],
                       company_whitelist_patterns=['safe.org']):
        """Helper method to create companys."""
        data = {'company_manager_id': company_manager_id,
                'company_name': company_name,
                'company_whitelist': company_whitelist,
                'company_whitelist_patterns': company_whitelist_patterns}
        response = hug.test.post(routes, '/v1/company', data)
        return response.data

    def create_github_organization(self, project_id, organization_installation_id=1,
                                   organization_name='GitHub Org Name'):
        """Helper method to create a GitHub organization."""
        data = {'organization_installation_id': organization_installation_id,
                'organization_name': organization_name,
                'organization_project_id': project_id}
        response = hug.test.post(routes, '/v1/github/organizations', data)
        return response.data

    def create_document(self, project_id, document_type='individual',
                        document_name='doc_name.pdf',
                        document_content_type='url+pdf',
                        document_content='http://url.com/document.pdf',
                        document_preamble='Test Preamble',
                        document_legal_entity_name='Legal Entity Name Inc.',
                        new_major_version=False):
        """Helper method to create a document."""
        data = {'document_name': document_name,
                'document_content_type': document_content_type,
                'document_content': document_content,
                'document_preamble': document_preamble,
                'document_legal_entity_name': document_legal_entity_name,
                'new_major_version': new_major_version}
        response = hug.test.post(routes,
                                 '/v1/project/' + project_id +
                                 '/document/' + document_type, data)
        return response.data

    def create_project(self, project_external_id='external-id', project_name='Project Name',
                       project_icla_enabled=True, project_ccla_enabled=True,
                       project_ccla_requires_icla_signature=True):
        """Helper method to create a project."""
        data = {'project_external_id': project_external_id,
                'project_icla_enabled': project_icla_enabled,
                'project_ccla_enabled': project_ccla_enabled,
                'project_ccla_requires_icla_signature': project_ccla_requires_icla_signature,
                'project_name': project_name}
        response = hug.test.post(routes, '/v1/project', data)
        return response.data

    def create_organization(self, organization_name='organization-name',
                            organization_project_id=None,
                            organization_installation_id=None):
        """Helper method to create a GitHub organization."""
        data = {'organization_name': organization_name}
        if organization_project_id is not None:
            data['organization_project_id'] = organization_project_id
        if organization_installation_id is not None:
            data['organization_installation_id'] = organization_installation_id
        response = hug.test.post(routes, '/v1/github/organizations', data)
        return response.data

    def tearDown(self):
        """Tear down method executes after each test."""
        delete_database()
