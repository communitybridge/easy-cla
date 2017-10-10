import { Injectable } from '@angular/core';
import { Http } from '@angular/http';

import 'rxjs/Rx';

import { CLA_API_URL } from './constants';

@Injectable()
export class ClaService {
  http: any;
  claApiUrl: String;

  constructor(http: Http) {
    this.http = http;
    this.claApiUrl = CLA_API_URL;
  }

  //////////////////////////////////////////////////////////////////////////////

  /**
  * This service should ONLY contain methods calling CLA API
  **/

  //////////////////////////////////////////////////////////////////////////////

  /**
  * /user
  **/

  getUsers() {
    return this.http.get(this.claApiUrl + '/user')
      .map(res => res.json());
  }

  postUser(user) {
    /*
      {
        'user_email': 'user@email.com',
        'user_name': 'User Name',
        'user_company_id': '<org-id>',
        'user_github_id': 12345
      }
     */
    return this.http.post(this.claApiUrl + '/user', user)
      .map((res) => res.json());
  }

  putUser(user) {
    /*
      {
        'user_id': '<user-id>',
        'user_email': 'user@email.com',
        'user_name': 'User Name',
        'user_company_id': '<org-id>',
        'user_github_id': 12345
      }
     */
    return this.http.put(this.claApiUrl + '/user', user)
      .map((res) => res.json());
  }

  /**
  * /user/{user_id}
  **/

  getUser(userId) {
    return this.http.get(this.claApiUrl + '/user/' + userId)
      .map((res) => res.json());
  }

  deleteUser(userId) {
    return this.http.delete(this.claApiUrl + '/user/' + userId)
      .map((res) => res.json());
  }

  /**
  * /user/email/{user_email}
  **/

  getUserByEmail(userEmail) {
    return this.http.get(this.claApiUrl + '/user/email/' + userEmail)
      .map((res) => res.json());
  }

  /**
  * /user/github/{user_github_id}
  **/

  getUserByGithubId(userGithubId) {
    return this.http.get(this.claApiUrl + '/user/github/' + userGithubId)
      .map((res) => res.json());
  }

  /**
  * /user/{user_id}/signatures
  **/

  getUserSignatures(userId) {
    return this.http.get(this.claApiUrl + '/user/' + userId + '/signatures')
      .map((res) => res.json());
  }

  /**
  * /users/company/{user_company_id}
  **/

  getUsersByCompanyId(userCompanyId) {
    return this.http.get(this.claApiUrl + '/users/company/' + userCompanyId)
      .map((res) => res.json());
  }

  /**
  * /user/{user_id}/request-company-whitelist/{company_id}
  **/

  postUserMessageToCompanyManager(userId, companyId) {
    return this.http.post(this.claApiUrl + '/user/' + userId + '/request-company-whitelist/' + companyId)
      .map((res) => res.json());
  }

  // This endpoint should actually have a message object.
  // postUserMessageToCompanyManager(userId, companyId, message) {
  //   return this.http.post(this.claApiUrl + '/user/' + userId + '/request-company-whitelist/' + companyId, message)
  //     .map((res) => res.json());
  // }

  /**
  * /signature
  **/

  getSignatures() {
    return this.http.get(this.claApiUrl + '/signature')
      .map((res) => res.json());
  }

  postSignature(signature) {
    /*
      signature: {
        'signature_type': ('cla' | 'dco'),
        'signature_signed': true,
        'signature_approved': true,
        'signature_sign_url': 'http://sign.com/here',
        'signature_return_url': 'http://cla-system.com/signed',
        'signature_project_id': '<project-id>',
        'signature_reference_id': '<ref-id>',
        'signature_reference_type': ('individual' | 'corporate'),
      }
      */
    return this.http.post(this.claApiUrl + '/signature', signature)
      .map((res) => res.json());
  }

  putSignature(signature) {
    /*
      signature: {
        'signature_id': '<signature-id>',
        'signature_type': ('cla' | 'dco'),
        'signature_signed': true,
        'signature_approved': true,
        'signature_sign_url': 'http://sign.com/here',
        'signature_return_url': 'http://cla-system.com/signed',
        'signature_project_id': '<project-id>',
        'signature_reference_id': '<ref-id>',
        'signature_reference_type': ('individual' | 'corporate'),
      }
      */
    return this.http.put(this.claApiUrl + '/signature', signature)
      .map((res) => res.json());
  }

  /**
  * /signature/{signature_id}
  **/

  getSignature(signatureId) {
    return this.http.get(this.claApiUrl + '/signature/' + signatureId)
      .map((res) => res.json());
  }

  deleteSignature(signatureId) {
    return this.http.delete(this.claApiUrl + '/signature/' + signatureId)
      .map((res) => res.json());
  }

  /**
  * /signatures/user/{user_id}
  **/

  getSignaturesUser(userId) {
    return this.http.get(this.claApiUrl + '/signatures/user/' + userId)
      .map((res) => res.json());
  }

  /**
  * /signatures/company/{company_id}
  **/

  getCompanySignatures(companyId) {
    return this.http.get(this.claApiUrl + '/signatures/company/' + companyId)
      .map((res) => res.json());
  }

  /**
  * /signatures/project/{project_id}
  **/

  getProjectSignatures(projectId) {
    return this.http.get(this.claApiUrl + '/signatures/project/' + projectId)
      .map((res) => res.json());
  }

  /**
  * /repository
  **/

  getRepositories() {
    return this.http.get(this.claApiUrl + '/repository')
      .map((res) => res.json());
  }

  postRepository(repository) {
    /*
      repository: {
        'repository_project_id': '<project-id>',
        'repository_external_id': 'repo1',
        'repository_name': 'Repo Name',
        'repository_type': 'github',
        'repository_url': 'http://url-to-repo.com'
      }
     */
    return this.http.post(this.claApiUrl + '/repository', repository)
       .map((res) => res.json());
  }

  putRepository(repository) {
    /*
      repository: {
        'repository_id': '<repo-id>',
        'repository_project_id': '<project-id>',
        'repository_external_id': 'repo1',
        'repository_name': 'Repo Name',
        'repository_type': 'github',
        'repository_url': 'http://url-to-repo.com'
      }
     */
    return this.http.put(this.claApiUrl + '/repository', repository)
       .map((res) => res.json());
  }

  /**
  * /repository/{repository_id}
  **/

  getRepository(repositoryId) {
    return this.http.get(this.claApiUrl + '/repository/' + repositoryId)
      .map((res) => res.json());
  }

  deleteRepository(repositoryId) {
    return this.http.delete(this.claApiUrl + '/repository/' + repositoryId)
      .map((res) => res.json());
  }

  /**
  * /company
  **/

  getCompanies() {
    return this.http.get(this.claApiUrl + '/company')
      .map((res) => res.json());
  }

  postCompany(company) {
    /*
      {
        'company_name': 'Org Name',
        'company_whitelist': ['safe@email.org'],
        'company_whitelist': ['*@email.org']
      }
     */
    return this.http.post(this.claApiUrl + '/company', company)
      .map((res) => res.json());
  }

  putCompany(company) {
    /*
      {
        'company_id': '<company-id>',
        'company_name': 'New Company Name'
      }
     */
    return this.http.put(this.claApiUrl + '/company', company)
      .map((res) => res.json());
  }

  /**
  * /company/{company_id}
  **/

  getCompany(companyId) {
    return this.http.get(this.claApiUrl + '/company/' + companyId)
      .map((res) => res.json());
  }

  deleteCompany(companyId) {
    return this.http.delete(this.claApiUrl + '/company/' + companyId)
      .map((res) => res.json());
  }

  /**
  * /project
  **/

  getProjects() {
    return this.http.get(this.claApiUrl + '/project')
      .map((res) => res.json());
  }

  postProject(project) {
    /*
      {
        'project_external_id': '<proj-external-id>',
        'project_name': 'Project Name', 'project_ccla_requires_icla_signature': True
      }
     */
    return this.http.post(this.claApiUrl + '/project', project)
      .map((res) => res.json());
  }

  putProject(project) {
    /*
      {
        'project_id': '<project-id>',
        'project_name': 'New Project Name'
      }
     */
     return this.http.put(this.claApiUrl + '/project', project)
       .map((res) => res.json());
  }

  /**
  * /project/{project_id}
  **/

  getProject(projectId) {
    return this.http.get(this.claApiUrl + '/project/' + projectId)
      .map((res) => res.json());
  }

  deleteProject(projectId) {
    return this.http.delete(this.claApiUrl + '/project/' + projectId)
      .map((res) => res.json());
  }

  /**
  * /project/{project_id}/repositories
  **/

  getProjectRepositories(projectId) {
    return this.http.get(this.claApiUrl + '/project/' + projectId + '/repositories')
      .map((res) => res.json());
  }

  /**
  * /project/{project_id}/document/{document_type}
  **/

  getProjectDocument(projectId, documentType) {
    return this.http.get(this.claApiUrl + '/project/' + projectId + '/document/' + documentType)
      .map((res) => res.json());
  }

  postProjectDocument(projectId, documentType, document) {
    /*
      {
        'document_name': 'doc_name.pdf',
        'document_content_type': 'url+pdf',
        'document_content': 'http://url.com/doc.pdf'
      }
     */
    return this.http.post(this.claApiUrl + '/project/' + projectId + '/document/' + documentType, document)
      .map((res) => res.json());
  }

  /**
  * /project/{project_id}/document/{document_type}/{major_version}/{minor_version}
  **/

  deleteProjectDocumentRevision(projectId, documentType, majorVersion, minorVersion) {
    return this.http.delete(this.claApiUrl + '/project/' + projectId + '/document/' + documentType + '/' + majorVersion + '/' + minorVersion)
      .map((res) => res.json());
  }

  /**
  * /request-signature
  **/

  postSignatureRequest(signatureRequest) {
    /*
      {
        'project_id': 'some-project-id',
        'user_id': 'some-user-uuid',
        'return_url': 'https://github.com/linuxfoundation/cla,
        'callback_url': 'http://cla.system/signed-callback'
      }
     */
    return this.http.post(this.claApiUrl + '/request-signature', signatureRequest)
      .map((res) => res.json());
  }

  /**
  * /signed/{installation_id}/{github_repository_id}/{change_request_id}
  **/

  postSigned(installationId, githubRepositoryId, changeRequestId) {
    return this.http.post(this.claApiUrl + '/signed/' + installationId + '/' + githubRepositoryId + '/' + changeRequestId)
      .map((res) => res.json());
  }

  /**
  * /return-url/{signature_id}
  **/

  getReturnUrl(signatureId) {
    return this.http.get(this.claApiUrl + '/return-url/' + signatureId)
      .map((res) => res.json());
  }

  /**
  * /repository-provider/{provider}/sign/{installation_id}/{github_repository_id}/{change_request_id}
  **/

  getSignRequest(provider, installationId, githubRepositoryId, changeRequestId) {
    return this.http.get(this.claApiUrl + '/repository-provider/' + provider + '/sign/' + installationId + '/' + githubRepositoryId + '/' + changeRequestId)
      .map((res) => res.json());
  }

  /**
  * /repository-provider/{provider}/icon.svg
  **/

  getChangeIcon(provider) {
    // This probably won't map to json, but instead to svg/xml
    return this.http.get(this.claApiUrl + '/repository-provider/' + provider + '/icon.svg')
      .map((res) => res.json());
  }

  /**
  * /repository-provider/{provider}/activity
  **/

  postReceivedActivity(provider) {
    return this.http.post(this.claApiUrl + '/repository-provider/' + provider + '/activity')
      .map((res) => res.json());
  }

  /**
  * /github/organizations
  **/

  getGithubOrganizations() {
    return this.http.get(this.claApiUrl + '/github/organizations')
      .map((res) => res.json());
  }

  postGithubOrganization(organization) {
    /*
      organization: {
        'organization_project_id': '<project-id>',
        'organization_name': 'org-name'
      }
     */
    return this.http.post(this.claApiUrl + '/github/organizations', organization)
      .map((res) => res.json());
  }

  /**
  * /github/organizations/{organization_name}
  **/

  getGithubOrganization(organizationName) {
    return this.http.get(this.claApiUrl + '/github/organizations' + organizationName)
      .map((res) => res.json());
  }

  deleteGithubOrganization(organizationName) {
    return this.http.delete(this.claApiUrl + '/github/organizations' + organizationName)
      .map((res) => res.json());
  }

  /**
  * /github/organizations/{organization_name}/repositories
  **/

  getGithubOrganizationRepositories(organizationName) {
    return this.http.get(this.claApiUrl + '/github/organizations' + organizationName + '/repositories')
      .map((res) => res.json());
  }

  /**
  * /github/installation
  **/

  getGithubInstallation() {
    return this.http.get(this.claApiUrl + '/github/installation')
      .map((res) => res.json());
  }

  postGithubInstallation() {
    return this.http.post(this.claApiUrl + '/github/installation')
      .map((res) => res.json());
  }

  /**
  * /github/activity
  **/

  postGithubActivity() {
    return this.http.post(this.claApiUrl + '/github/activity')
      .map((res) => res.json());
  }

  /**
  * /github/validate
  **/

  postGithubValidate() {
    return this.http.post(this.claApiUrl + '/github/validate')
      .map((res) => res.json());
  }

  //////////////////////////////////////////////////////////////////////////////

}
