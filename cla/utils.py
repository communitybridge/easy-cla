"""
Utility functions for the CLA project.
"""

import urllib.parse
import falcon
from requests_oauthlib import OAuth2Session
from hug.middleware import SessionMiddleware
import cla

def get_session_middleware():
    """Prepares the hug middleware to manage key-value session data."""
    keyvalue = cla.conf['KEYVALUE']
    if keyvalue == 'Memory':
        from hug.store import InMemoryStore as Store
    elif keyvalue == 'DynamoDB':
        from cla.models.dynamo_models import Store
    else:
        raise Exception('Invalid key-value store selected in configuration: %s' %keyvalue)
    return SessionMiddleware(Store(), context_name='session', cookie_name='cla-sid',
                             cookie_max_age=300, cookie_domain=None, cookie_path='/',
                             cookie_secure=False)

def create_database(conf=None):
    """
    Helper function to create the CLA database. Will utilize the appropriate database
    provider based on configuration.

    :param conf: Configuration dictionary/object - typically parsed from the CLA config file.
    :type conf: dict
    """
    if conf is None:
        conf = cla.conf
    cla.log.info('Creating CLA database in %s', conf['DATABASE'])
    if conf['DATABASE'] == 'DynamoDB':
        from cla.models.dynamo_models import create_database as cd
    else:
        raise Exception('Invalid database selection in configuration: %s' %conf['DATABASE'])
    cd()

def delete_database(conf=None):
    """
    Helper function to delete the CLA database. Will utilize the appropriate database
    provider based on configuration.

    :WARNING: Use with caution.

    :param conf: Configuration dictionary/object - typically parsed from the CLA config file.
    :type conf: dict
    """
    if conf is None:
        conf = cla.conf
    cla.log.warning('Deleting CLA database in %s', conf['DATABASE'])
    if conf['DATABASE'] == 'DynamoDB':
        from cla.models.dynamo_models import delete_database as dd
    else:
        raise Exception('Invalid database selection in configuration: %s' %conf['DATABASE'])
    dd()

def get_database_models(conf=None):
    """
    Returns the database models based on the configuration dict provided.

    :param conf: Configuration dictionary/object - typically parsed from the CLA config file.
    :type conf: dict
    :return: Dictionary of all the supported database object classes (User, Signature, Repository,
        company, Project, Document) - keyed by name:

            {'User': cla.models.model_interfaces.User,
             'Signature': cla.models.model_interfaces.Signature,...}

    :rtype: dict
    """
    if conf is None:
        conf = cla.conf
    if conf['DATABASE'] == 'DynamoDB':
        from cla.models.dynamo_models import User, Signature, Repository, \
                                             Company, Project, Document, GitHubOrg
        return {'User': User, 'Signature': Signature, 'Repository': Repository,
                'Company': Company, 'Project': Project, 'Document': Document,
                'GitHubOrg': GitHubOrg}
    else:
        raise Exception('Invalid database selection in configuration: %s' %conf['DATABASE'])

def get_user_instance(conf=None):
    """
    Helper function to get a database User model instance based on CLA configuration.

    :param conf: Same as get_database_models().
    :type conf: dict
    :return: A User model instance based on configuration specified.
    :rtype: cla.models.model_interfaces.User
    """
    return get_database_models(conf)['User']()

def get_signature_instance(conf=None):
    """
    Helper function to get a database Signature model instance based on CLA configuration.

    :param conf: Same as get_database_models().
    :type conf: dict
    :return: An Signature model instance based on configuration.
    :rtype: cla.models.model_interfaces.Signature
    """
    return get_database_models(conf)['Signature']()

def get_repository_instance(conf=None):
    """
    Helper function to get a database Repository model instance based on CLA configuration.

    :param conf: Same as get_database_models().
    :type conf: dict
    :return: A Repository model instance based on configuration specified.
    :rtype: cla.models.model_interfaces.Repository
    """
    return get_database_models(conf)['Repository']()

def get_github_organization_instance(conf=None):
    """
    Helper function to get a database GitHubOrg model instance based on CLA configuration.

    :param conf: Same as get_database_models().
    :type conf: dict
    :return: A Repository model instance based on configuration specified.
    :rtype: cla.models.model_interfaces.GitHubOrg
    """
    return get_database_models(conf)['GitHubOrg']()

def get_company_instance(conf=None):
    """
    Helper function to get a database company model instance based on CLA configuration.

    :param conf: Same as get_database_models().
    :type conf: dict
    :return: A company model instance based on configuration specified.
    :rtype: cla.models.model_interfaces.Company
    """
    return get_database_models(conf)['Company']()

def get_project_instance(conf=None):
    """
    Helper function to get a database Project model instance based on CLA configuration.

    :param conf: Same as get_database_models().
    :type conf: dict
    :return: A Project model instance based on configuration specified.
    :rtype: cla.models.model_interfaces.Project
    """
    return get_database_models(conf)['Project']()

def get_document_instance(conf=None):
    """
    Helper function to get a database Document model instance based on CLA configuration.

    :param conf: Same as get_database_models().
    :type conf: dict
    :return: A Document model instance based on configuration specified.
    :rtype: cla.models.model_interfaces.Document
    """
    return get_database_models(conf)['Document']()

def get_email_service(conf=None, initialize=True):
    """
    Helper function to get the configured email service instance.

    :param conf: Same as get_database_models().
    :type conf: dict
    :param initialize: Whether or not to run the initialize method on the instance.
    :type initialize: boolean
    :return: The email service model instance based on configuration specified.
    :rtype: EmailService
    """
    if conf is None:
        conf = cla.conf
    email_service = conf['EMAIL_SERVICE']
    if email_service == 'SMTP':
        from cla.models.smtp_models import SMTP as email
    elif email_service == 'MockSMTP':
        from cla.models.smtp_models import MockSMTP as email
    elif email_service == 'SES':
        from cla.models.ses_models import SES as email
    elif email_service == 'MockSES':
        from cla.models.ses_models import MockSES as email
    else:
        raise Exception('Invalid email service selected in configuration: %s' %email_service)
    email_instance = email()
    if initialize:
        email_instance.initialize(conf)
    return email_instance

def get_signing_service(conf=None, initialize=True):
    """
    Helper function to get the configured signing service instance.

    :param conf: Same as get_database_models().
    :type conf: dict
    :param initialize: Whether or not to run the initialize method on the instance.
    :type initialize: boolean
    :return: The signing service instance based on configuration specified.
    :rtype: SigningService
    """
    if conf is None:
        conf = cla.conf
    signing_service = conf['SIGNING_SERVICE']
    if signing_service == 'DocuSign':
        from cla.models.docusign_models import DocuSign as signing
    elif signing_service == 'MockDocuSign':
        from cla.models.docusign_models import MockDocuSign as signing
    else:
        raise Exception('Invalid signing service selected in configuration: %s' %signing_service)
    signing_service_instance = signing()
    if initialize:
        signing_service_instance.initialize(conf)
    return signing_service_instance

def get_storage_service(conf=None, initialize=True):
    """
    Helper function to get the configured storage service instance.

    :param conf: Same as get_database_models().
    :type conf: dict
    :param initialize: Whether or not to run the initialize method on the instance.
    :type initialize: boolean
    :return: The storage service instance based on configuration specified.
    :rtype: StorageService
    """
    if conf is None:
        conf = cla.conf
    storage_service = conf['STORAGE_SERVICE']
    if storage_service == 'LocalStorage':
        from cla.models.local_storage import LocalStorage as storage
    elif storage_service == 'S3Storage':
        from cla.models.s3_storage import S3Storage as storage
    elif storage_service == 'MockS3Storage':
        from cla.models.s3_storage import MockS3Storage as storage
    else:
        raise Exception('Invalid storage service selected in configuration: %s' %storage_service)
    storage_instance = storage()
    if initialize:
        storage_instance.initialize(conf)
    return storage_instance

def get_supported_repository_providers():
    """
    Returns a dict of supported repository service providers.

    :return: Dictionary of supported repository service providers in the following
        format: {'<provider_name>': <provider_class>}
    :rtype: dict
    """
    from cla.models.github_models import GitHub, MockGitHub
    #from cla.models.gitlab_models import GitLab, MockGitLab
    #return {'github': GitHub, 'mock_github': MockGitHub,
            #'gitlab': GitLab, 'mock_gitlab': MockGitLab}
    return {'github': GitHub, 'mock_github': MockGitHub}

def get_repository_service(provider, initialize=True):
    """
    Get a repository service instance by provider name.

    :param provider: The provider to load.
    :type provider: string
    :param initialize: Whether or not to call the initialize() method on the object.
    :type initialize: boolean
    :return: A repository provider instance (GitHub, Gerrit, etc).
    :rtype: RepositoryService
    """
    providers = get_supported_repository_providers()
    if provider not in providers:
        raise NotImplementedError('Provider not supported')
    instance = providers[provider]()
    if initialize:
        instance.initialize(cla.conf)
    return instance

def get_repository_service_by_repository(repository, initialize=True):
    """
    Helper function to get a repository service provider instance based
    on a repository.

    :param repository: The repository object or repository_id.
    :type repository: cla.models.model_interfaces.Repository | string
    :param initialize: Whether or not to call the initialize() method on the object.
    :type initialize: boolean
    :return: A repository provider instance (GitHub, Gerrit, etc).
    :rtype: RepositoryService
    """
    repository_model = get_database_models()['Repository']
    if isinstance(repository, repository_model):
        repo = repository
    else:
        repo = repository_model()
        repo.load(repository)
    provider = repo.get_repository_type()
    return get_repository_service(provider, initialize)

def get_supported_document_content_types(): # pylint: disable=invalid-name
    """
    Returns a list of supported document content types.

    :return: List of supported document content types.
    :rtype: dict
    """
    return ['pdf', 'url+pdf', 'storage+pdf']

def get_project_document(project, document_type, revision):
    """
    Helper function to get the specified document from a project.

    :param project: The project model object to look in.
    :type project: cla.models.model_interfaces.Project
    :param document_type: The type of document (individual or corporate).
    :type document_type: string
    :param revision: The revision number to look for.
    :type revision: integer
    :return: The document model if found.
    :rtype: cla.models.model_interfaces.Document
    """
    if document_type == 'individual':
        documents = project.get_project_individual_documents()
    else:
        documents = project.get_project_corporate_documents()
    for document in documents:
        if document.get_document_revision() == revision:
            return document
    return None

def get_last_revision(documents):
    """
    Helper function to get the last revision of the list of documents provided.

    :param documents: List of documents to check.
    :type documents: [cla.models.model_interfaces.Document]
    """
    last = 0
    for document in documents:
        if document.get_document_revision() > last:
            last = document.get_document_revision()
    return last

def user_signed_project_signature(user, repository):
    """
    Helper function to check if a user has signed a project signature tied to a repository.

    :param user: The user object to check for.
    :type user: cla.models.model_interfaces.User
    :param repository: The repository to check for.
    :type repository: cla.models.model_interfaces.Repository
    :return: Whether or not the user has an signature that's signed and approved
        for this project based on repository.
    :rtype: boolean
    """
    project_id = repository.get_repository_project_id()
    signatures = user.get_user_signatures(project_id=project_id)
    num_signatures = len(signatures)
    if num_signatures > 0:
        signature = signatures[0]
        cla.log.info('Signature found for this user on project %s: %s',
                     project_id, signature.get_signature_id())
        if signature.get_signature_signed() and signature.get_signature_approved():
            # Signature found and signed/approved.
            cla.log.info('User already has a signed and approved signature ' + \
                         'for project: %s', project_id)
            return True
        elif signature.get_signature_signed(): # Not approved yet.
            cla.log.warning('Signature (%s) has not been approved yet for ' + \
                            'user %s on project %s', \
                            signature.get_signature_id(),
                            user.get_user_email(),
                            project_id)
        else: # Not signed or approved yet.
            cla.log.info('Signature (%s) has not been signed by %s for project %s', \
                         signature.get_signature_id(),
                         user.get_user_email(),
                         project_id)
    else: # Signature not found for this user on this project.
        cla.log.info('Signature not found for project %s and user %s',
                     project_id, user.get_user_email())
    return False

def get_user_signature_by_repository(repository, user):
    """
    Helper function to get a user's signature for a specified repository.

    :param repository: The Repository object.
    :type repository: cla.models.model_interfaces.Repository
    :param user: The user object that represents the signature signer.
    :type user: cla.models.model_interfaces.User
    :return: The signature for this user on this repository, or None if not found.
    :rtype: cla.models.model_interfaces.Signature | None
    """

    project_id = repository.get_repository_project_id()
    signatures = user.get_user_signatures(project_id=project_id)
    num_signatures = len(signatures)
    if num_signatures > 0:
        return signatures[0]
    return None

def get_redirect_uri(repository_service, repository_id, change_request_id):
    """
    Function to generate the redirect_uri parameter for a repository service's OAuth2 process.

    :param repository_service: The repository service provider we're currently initiating the
        OAuth2 process with. Currently only supports 'github' and 'gitlab'.
    :type repository_service: string
    :param repository_id: The ID of the repository object that applies for this OAuth2 process.
    :type repository_id: string
    :param change_request_id: The ID of the change request in question. Is a PR number if
        repository_service is 'github'. Is a merge request if the repository_service is 'gitlab'.
    :type change_request_id: string
    :return: The redirect_uri parameter expected by the OAuth2 process.
    :rtype: string
    """
    params = {'repository_id': repository_id,
              'change_request_id': change_request_id}
    params = urllib.parse.urlencode(params)
    return cla.conf['BASE_URL'] + '/v1/repository-provider/' + repository_service + \
           '/oauth2_redirect?' + params

def get_full_sign_url(repository_service, repository_id, change_request_id):
    """
    Helper function to get the full sign URL that the user should click to initiate the signing
    workflow.

    :param repository_service: The repository service provider we're getting the sign url for.
        Should be one of the supported repository providers ('github', 'gitlab', etc).
    :type repository_service: string
    :param repository_id: The ID of the repository for this signature (used in order to figure out
        where to send the user once signing is complete.
    :type repository_id: int
    :param change_request_id: The change request ID for this signature (used in order to figure out
        where to send the user once signing is complete. Should be a pull request number when
        repository_service is 'github'. Should be a merge request ID when repository_service is
        'gitlab'.
    :type change_request_id: int
    """
    return cla.conf['BASE_URL'] + '/v1/repository-provider/' + repository_service + '/sign/' + \
           str(repository_id) + '/' + str(change_request_id)

def get_comment_badge(repository_type, all_signed, sign_url):
    """
    Returns the CLA badge that will appear on the change request comment (PR for 'github', merge
    request for 'gitlab', etc)

    :param repository_type: The repository service provider we're getting the badge for.
        Should be one of the supported repository providers ('github', 'gitlab', etc).
    :type repository_type: string
    :param all_signed: Whether or not all committers have signed the change request.
    :type all_signed: boolean
    :param sign_url: The URL for the user to click in order to initiate signing.
    :type sign_url: string
    """
    badge_url = cla.conf['BASE_URL'] + '/v1/repository-provider/' + repository_type + '/icon.svg'
    if all_signed:
        badge_url += '?signed=1'
    else:
        badge_url += '?signed=0'
    return '[![CLA Check](' + badge_url + ')](' + sign_url + ')'

def assemble_cla_status(author_name, signed=False):
    """
    Helper function to return the text that will display on a change request status.

    For GitLab there isn't much space here - we rely on the user hovering their mouse over the icon.
    For GitHub there is a 140 character limit.

    :param author_name: The name of the author of this commit.
    :type author_name: string
    :param signed: Whether or not the author has signed an signature.
    :type signed: boolean
    """
    if author_name is None:
        author_name = 'Unknown'
    if signed:
        return 'Thank you for signing the CLA.'
    return 'Still missing CLA signature from %s.' %author_name

def assemble_cla_comment(repository_type, repository_id, change_request_id, signed, missing):
    """
    Helper function to generate a CLA comment based on a a change request.

    :param repository_type: The type of repository this comment will be posted on ('github',
        'gitlab', etc).
    :type repository_type: string
    :param repository_id: The ID of the repository for this change request.
    :type repository_id: int
    :param change_request_id: The repository service's ID of this change request.
    :type change_request_id: id
    :param signed: The list of commit hashes and authors that have signed an signature for this
        change request.
    :type signed: [(string, string)]
    :param missing: The list of commit hashes and authors that have not signed for this
        change request.
    :type missing: [(string, string)]
    """
    num_missing = len(missing)
    sign_url = get_full_sign_url(repository_type, repository_id, change_request_id)
    comment = get_comment_body(repository_type, sign_url, signed, missing)
    all_signed = num_missing == 0
    badge = get_comment_badge(repository_type, all_signed, sign_url)
    return badge + '<br />' + comment

def get_comment_body(repository_type, sign_url, signed, missing):
    """
    Returns the CLA comment that will appear on the repository provider's change request item.

    :param repository_type: The repository type where this comment will be posted ('github',
        'gitlab', etc).
    :type repository_type: string
    :param sign_url: The URL for the user to click in order to initiate signing.
    :type sign_url: string
    :param signed: List of tuples containing the commit and author name of signers.
    :type signed: [(string, string)]
    :param missing: List of tuples containing the commit and author name of not-signed users.
    :type missing: [(string, string)]
    """
    cla.log.info('Getting comment body for repository type: %s', repository_type)
    unchecked = ':white_large_square:'
    checked = ':white_check_mark:'
    committers_comment = ''
    num_signed = len(signed)
    num_missing = len(missing)
    if num_signed > 0:
        # Group commits by author.
        committers = {}
        for commit, author in signed:
            if author is None:
                author = 'Unknown'
            if author not in committers:
                committers[author] = []
            committers[author].append(commit)
        # Print author commit information.
        committers_comment += '<ul>'
        for author, commit_hashes in committers.items():
            committers_comment += '<li>' + checked + '  ' + author + \
                                  ' (' + ", ".join(commit_hashes) + ')</li>'
        committers_comment += '</ul>'
    if num_missing > 0:
        text = 'Thank you! Please sign our [Contributor License Signature](' + \
               sign_url + ') before we can accept your contribution:'
        # Group commits by author.
        committers = {}
        for commit, author in missing:
            if author is None:
                author = 'Unknown'
            if author not in committers:
                committers[author] = []
            committers[author].append(commit)
        # Print author commit information.
        committers_comment += '<ul>'
        for author, commit_hashes in committers.items():
            committers_comment += '<li>[' + unchecked + '](' + sign_url + ')  ' + \
                                  author + ' (' + ", ".join(commit_hashes) + ')</li>'
        committers_comment += '</ul>'
        return text + committers_comment
    text = 'All committers have signed the CLA:'
    return text + committers_comment

def get_authorization_url_and_state(client_id, redirect_uri, scope, authorize_url):
    """
    Helper function to get an OAuth2 session authorization URL and state.

    :param client_id: The client ID for this OAuth2 session.
    :type client_id: string
    :param redirect_uri: The redirect URI to specify in this OAuth2 session.
    :type redirect_uri: string
    :param scope: The list of scope items to use for this OAuth2 session.
    :type scope: [string]
    :param authorize_url: The URL to submit the OAuth2 request.
    :type authorize_url: string
    """
    oauth = OAuth2Session(client_id, redirect_uri=redirect_uri, scope=scope)
    authorization_url, state = oauth.authorization_url(authorize_url)
    return authorization_url, state

def fetch_token(client_id, state, token_url, client_secret, code, redirect_uri=None): # pylint: disable=too-many-arguments
    """
    Helper function to fetch a OAuth2 session token.

    :param client_id: The client ID for this OAuth2 session.
    :type client_id: string
    :param state: The OAuth2 session state.
    :type state: string
    :param token_url: The token URL for this OAuth2 session.
    :type token_url: string
    :param code: The OAuth2 session code.
    :type code: string
    :param redirect_uri: The redirect URI for this OAuth2 session.
    :type redirect_uri: string
    """
    if redirect_uri is not None:
        oauth2 = OAuth2Session(client_id, state=state, redirect_uri=redirect_uri)
    else:
        oauth2 = OAuth2Session(client_id, state=state)
    return oauth2.fetch_token(token_url, client_secret=client_secret, code=code)

def redirect_user_by_signature(user, signature):
    """
    Helper method to redirect a user based on their signature status and return_url.

    :param user: The user object for this redirect.
    :type user: cla.models.model_interfaces.User
    :param signature: The signature object for this user.
    :type signature: cla.models.model_interfaces.Signature
    """
    return_url = signature.get_signature_return_url()
    if signature.get_signature_signed() and signature.get_signature_approved():
        # Signature already signed and approved.
        # TODO: Notify user of signed and approved signature somehow.
        cla.log.info('Signature already signed and approved for user: %s, %s',
                     user.get_user_email(), signature.get_signature_id())
        cla.log.info('Redirecting user back to %s', return_url)
        raise falcon.HTTPFound(return_url)
    elif signature.get_signature_signed():
        # Awaiting approval.
        # TODO: Notify user of pending approval somehow.
        cla.log.info('Signature signed but not approved yet: %s',
                     signature.get_signature_id())
        cla.log.info('Redirecting user back to %s', return_url)
        raise falcon.HTTPFound(return_url)
    else:
        # Signature awaiting signature.
        sign_url = signature.get_signature_sign_url()
        signature_id = signature.get_signature_id()
        cla.log.info('Signature exists, sending user to sign: %s (%s)', signature_id, sign_url)
        raise falcon.HTTPFound(sign_url)

def request_signature(repository, user, change_request_id, callback_url=None):
    """
    Helper function send the user off to sign an signature based on the repository.

    :param repository: The repository object in question.
    :type repository: cla.models.model_interfaces.Repository
    :param user: The user in question.
    :type user: cla.models.model_interfaces.User
    :param change_request_id: The change request ID (used to redirect the user after signing).
    :type change_request_id: string
    :param callback_url: Optionally provided a callback_url. Will default to
        <SIGNED_CALLBACK_URL>/<repo_id>/<change_request_id>.
    :type callback_url: string
    """
    project_id = repository.get_repository_project_id()
    repo_id = repository.get_repository_id()
    repo_service = get_repository_service(repository.get_repository_type())
    return_url = repo_service.get_return_url(repository.get_repository_external_id(),
                                             change_request_id)
    if callback_url is None:
        callback_url = cla.conf['SIGNED_CALLBACK_URL'] + \
                       '/' + str(repo_id) + '/' + str(change_request_id)
    signing_service = get_signing_service()
    signature_data = signing_service.request_signature(project_id,
                                                       user.get_user_id(),
                                                       return_url,
                                                       callback_url)
    if 'sign_url' in signature_data:
        raise falcon.HTTPFound(signature_data['sign_url'])
    cla.log.error('Could not get sign_url from signing service provider - sending user ' + \
                  'to return_url instead')
    raise falcon.HTTPFound(return_url)

def change_icon(provider, signed=False): # pylint: disable=unused-argument
    """
    Function called when the code chagne image/icon is requested.

    This will be a badge for GitHub and GitLab providers.

    TODO: Fire a hook here with the provider and signed variables as parameters. This will allow
    customizeable change icon images.

    :param provider: The repository service provider asking for the change icon image.
    :type provider: string
    :param signed: Whether this image is for a signed or unsigned CLA.
    :type signed: boolean
    :return: Anything compatible with hug's output_format.svg_xml_image.
    :rtype: file path | file handler | Pillow Image
    """
    if signed:
        return 'cla/resources/cla-signed.svg'
    return 'cla/resources/cla-unsigned.svg'
