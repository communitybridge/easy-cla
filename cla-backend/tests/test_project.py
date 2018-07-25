"""
Tests having to do with the CLA projects.
"""

import unittest
import uuid
import hug

# Importing to setup proper python path and DB for tests.
from test_cla import CLATestCase
import cla

class ProjectTestCase(CLATestCase):
    """Project test cases."""
    def test_get_project(self):
        """Tests for getting projects."""
        response = hug.test.get(cla.routes, '/v1/project')
        self.assertEqual(response.data, [])
        self.create_project()
        self.create_project()
        response = hug.test.get(cla.routes, '/v1/project')
        self.assertEqual(len(response.data), 2)

    def test_get_project_by_external_id(self):
        """Tests for getting projects by external ID."""
        response = hug.test.get(cla.routes, '/v1/project/external/fake-external-id')
        self.assertEqual(len(response.data), 0)
        project = self.create_project()
        response = hug.test.get(cla.routes, '/v1/project/external/external-id')
        self.assertEqual(len(response.data), 1)
        self.assertEqual(response.data[0]['project_id'], project['project_id'])

    def test_get_project_by_organization_id(self):
        """Tests for getting projects by github organization ID."""
        project = self.create_project()
        organization = self.create_organization(organization_project_id=project['project_id'])
        response = hug.test.get(cla.routes, '/v1/project/' + project['project_id'] + '/organizations')
        self.assertEqual(len(response.data), 1)
        self.assertEqual(response.data[0]['organization_project_id'], project['project_id'])

    def test_get_project(self):
        """Tests for getting individual projects."""
        response = hug.test.get(cla.routes, '/v1/project/' + str(uuid.uuid4()))
        self.assertEqual(response.data, {'errors': {'project_id': 'Project not found'}})
        project = self.create_project()
        response = hug.test.get(cla.routes, '/v1/project/' + project['project_id'])
        self.assertEqual(response.data['project_id'], project['project_id'])

    def test_post_project(self):
        """Tests for creating projects."""
        project = self.create_project(project_name='Name Here')
        response = hug.test.get(cla.routes, '/v1/project/' + project['project_id'])
        self.assertEqual(response.data['project_name'], 'Name Here')

    def test_put_project(self):
        """Tests for updating projects."""
        project = self.create_project()
        response = hug.test.put(cla.routes, '/v1/project',
                                {'project_id': project['project_id'],
                                 'project_name': 'New Name'})
        self.assertEqual(response.data['project_name'], 'New Name')

    def test_delete_project(self):
        """Tests for deleting projects."""
        project = self.create_project()
        response = hug.test.get(cla.routes, '/v1/project')
        self.assertTrue(len(response.data), 1)
        response = hug.test.delete(cla.routes, '/v1/project/' + str(uuid.uuid4()))
        self.assertEqual(response.data, {'errors': {'project_id': 'Project not found'}})
        response = hug.test.delete(cla.routes, '/v1/project/' + project['project_id'])
        self.assertEqual(response.data, {'success': True})
        response = hug.test.get(cla.routes, '/v1/project')
        self.assertEqual(len(response.data), 0)

    def test_post_project_document(self):
        """Tests for creating project documents"""
        project = self.create_project('Name Here')
        project_id = project['project_id']
        self.create_document(project_id)
        self.create_document(project_id)
        doc3 = self.create_document(project_id, 'corporate')
        self.assertTrue(len(doc3['project_individual_documents']) == 2)
        self.assertTrue(len(doc3['project_corporate_documents']) == 1)
        response = hug.test.get(cla.routes,
                                '/v1/project/' + project_id + '/document/individual')
        self.assertEqual(response.data['document_major_version'], 1)
        self.assertEqual(response.data['document_minor_version'], 1)
        self.assertTrue('document_creation_date' in response.data)
        response = hug.test.get(cla.routes,
                                '/v1/project/' + project_id + '/document/corporate')
        self.assertEqual(response.data['document_major_version'], 1)
        self.assertEqual(response.data['document_minor_version'], 0)
        path = '/v1/project/' + project_id + '/document/individual/3/0'
        response = hug.test.delete(cla.routes, path)
        self.assertEqual(response.data, {'errors': {'document': 'Document version not found'}})
        path = '/v1/project/' + project_id + '/document/individual/1/1'
        response = hug.test.delete(cla.routes, path)
        self.assertEqual(response.data, {'success': True})
        response = hug.test.get(cla.routes,
                                '/v1/project/' + project_id + '/document/individual')
        self.assertEqual(response.data['document_major_version'], 1)
        self.assertEqual(response.data['document_minor_version'], 0)
        self.create_document(project_id, 'individual', new_major_version=True)
        response = hug.test.get(cla.routes,
                                '/v1/project/' + project_id + '/document/individual')
        self.assertEqual(response.data['document_major_version'], 2)
        self.assertEqual(response.data['document_minor_version'], 0)
        self.create_document(project_id, 'individual', new_major_version=False)
        response = hug.test.get(cla.routes,
                                '/v1/project/' + project_id + '/document/individual')
        self.assertEqual(response.data['document_major_version'], 2)
        self.assertEqual(response.data['document_minor_version'], 1)
        path = '/v1/project/' + project_id + '/document/corporate/1/0'
        response = hug.test.delete(cla.routes, path)
        self.assertEqual(response.data, {'success': True})
        response = hug.test.get(cla.routes,
                                '/v1/project/' + project_id + '/document/corporate')
        self.assertEqual(response.data,
                         {'errors': {'document': 'No corporate document exists for this project'}})

if __name__ == '__main__':
    unittest.main()
