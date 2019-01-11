"""
Controller related to repository operations.
"""

import uuid
import cla.hug_types
import os
import cla 
from cla.utils import get_gerrit_instance
from cla.models import DoesNotExist
from cla.controllers.lf_group import LFGroup

lf_group_client_url = os.environ.get('LF_GROUP_CLIENT_URL', '')
lf_group_client_id = os.environ.get('LF_GROUP_CLIENT_ID', '')
lf_group_client_secret = os.environ.get('LF_GROUP_CLIENT_SECRET', '')
lf_group_refresh_token = os.environ.get('LF_GROUP_REFRESH_TOKEN', '')
lf_group = LFGroup(lf_group_client_url, lf_group_client_id, lf_group_client_secret, lf_group_refresh_token)

def get_gerrit_by_project_id(project_id):
    gerrit = get_gerrit_instance()
    try:
        gerrit = gerrit.get_gerrit_by_project_id(project_id)
    except DoesNotExist as err:
        return {'errors': {'a gerrit instance does not exist with the given project ID. ': str(err)}}

    if gerrit is None:
        return []

    return [gerrit.to_dict()]


def create_gerrit(project_id, 
                    gerrit_name, 
                    gerrit_url, 
                    group_id_icla,
                    group_id_ccla):
    """
    Creates a gerrit instance and returns the newly created gerrit object dict format.

    :param gerrit_project_id: The project ID of the gerrit instance
    :type gerrit_project_id: string
    :param gerrit_name: The new gerrit instance name
    :type gerrit_name: string
    :param gerrit_url: The new Gerrit URL.
    :type gerrit_url: string
    :param group_id_icla: The id of the LDAP group for ICLA. 
    :type group_id_icla: string
    :param group_id_ccla: The id of the LDAP group for CCLA. 
    :type group_id_ccla: string
    """
    
    gerrit = get_gerrit_instance()
    gerrit.set_gerrit_id(str(uuid.uuid4()))
    gerrit.set_project_id(str(project_id))
    gerrit.set_gerrit_url(gerrit_url)
    gerrit.set_gerrit_name(gerrit_name)
    
    #check if LDAP group exists     
    # returns 'error' if the LDAP group does not exist
    ldap_group_icla = lf_group.get_group(group_id_icla)
    if ldap_group_icla.get('error') is not None:
        return {'error_icla': 'The specified LDAP group for ICLA does not exist. '}
    gerrit.set_group_name_icla(ldap_group_icla.get('title'))
    gerrit.set_group_id_icla(str(group_id_icla))

    ldap_group_ccla = lf_group.get_group(group_id_ccla)
    if ldap_group_ccla.get('error') is not None:
        return {'error_ccla': 'The specified LDAP group for CCLA does not exist. '}
    gerrit.set_group_name_ccla(ldap_group_ccla.get('title'))
    gerrit.set_group_id_ccla(str(group_id_ccla))

    gerrit.save()
    return [gerrit.to_dict()]

def delete_gerrit(gerrit_id):
    """
    Deletes a gerrit instance

    :param gerrit_id: The ID of the gerrit instance.
    """
    gerrit = get_gerrit_instance()
    try:
        gerrit.load(str(gerrit_id))
    except DoesNotExist as err:
        return {'errors': {'gerrit_id': str(err)}}
    gerrit.delete()
    return {'success': True}


def get_agreement_html(project_id, contract_type):
    contributor_base_url = cla.conf['CONTRIBUTOR_BASE_URL']
    return """
        <html>
            <a href="https://{contributor_base_url}/#/cla/gerrit/project/{project_id}/{contract_type}">Click on the link to Sign the CLA Agreement. </a>
        <html>""".format(
            contributor_base_url = contributor_base_url,
            project_id = project_id,
            contract_type = contract_type
        )
