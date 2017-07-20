import { Injectable } from '@angular/core';
import { Http } from '@angular/http';

import 'rxjs/Rx';

@Injectable()
export class CincoService {
  http: any;
  baseUrl: String;

  constructor(http: Http) {
    this.http = http;
    this.baseUrl = '';
  }

  //////////////////////////////////////////////////////////////////////////////
  getSessionData() {
    return this.http.get(this.baseUrl + '/session_data')
      .map(res => res.json());
  }

  /**
  * Projects
  * Resources to expose and manipulate details of projects
  **/

  getProjectStatuses() {
    return this.http.get(this.baseUrl + '/project/status')
      .map(res => res.json());
  }

  getProjectCategories() {
    return this.http.get(this.baseUrl + '/project/categories')
      .map(res => res.json());
  }

  getProjectSectors() {
    return this.http.get(this.baseUrl + '/project/sectors')
      .map(res => res.json());
  }

  getAllProjects() {
    return this.http.get(this.baseUrl + '/projects')
      .map(res => res.json());
  }

  //  Disabled for PMC v1.0
  //  createProject(newProject) {
  //    return this.http.post('/projects', newProject)
  //            .map((res) => res.json());
  //  }

  getProject(projectId, getMembers) {
    if (getMembers) { projectId = projectId + '?members=true'; }
    return this.http.get(this.baseUrl + '/projects/' + projectId)
      .map(res => res.json());
  }

  editProject(projectId, editProject) {
    return this.http.put('/projects/' + projectId, editProject)
      .map((res) => res.json());
  }

  getProjectConfig(projectId) {
    return this.http.get(this.baseUrl + '/projects/' + projectId + '/config')
      .map(res => res.json());
  }

  updateProjectManagers(projectId, updatedManagers) {
    return this.http.put('/projects/' + projectId + '/managers', updatedManagers)
      .map((res) => res.json());
  }

  //////////////////////////////////////////////////////////////////////////////

  /**
  * Projects - Members
  * Resources for getting details about project members
  **/

  getProjectMembers(projectId) {
    return this.http.get(this.baseUrl + '/projects/' + projectId + '/members')
      .map(res => res.json());
  }

  getMember(projectId, memberId) {
    var response = this.http.get(this.baseUrl + '/projects/' + projectId + '/members/' + memberId)
      .map(res => res.json());
    return response;
  }

  //////////////////////////////////////////////////////////////////////////////

  /**
  * Projects - Members - Contacts
  * Resources for getting and manipulating contacts of project members
  **/

  getMemberContactRoles() {
    var response = this.http.get(this.baseUrl + '/project/members/contacts/types')
      .map(res => res.json());
    return response;
  }

  getMemberContacts(projectId, memberId) {
    var response = this.http.get(this.baseUrl + '/projects/' + projectId + '/members/' + memberId + '/contacts')
      .map(res => res.json());
    return response;
  }

  addMemberContact(projectId, memberId, contactId, newContact) {
    return this.http.post('/projects/' + projectId + '/members/' + memberId + '/contacts/' + contactId, newContact)
      .map((res) => res.json());
  }

  removeMemberContact(projectId, memberId, contactId, roleId) {
    return this.http.delete('/projects/' + projectId + '/members/' + memberId + '/contacts/' + contactId + '/roles/' + roleId)
      .map((res) => res.json());
  }

  updateMemberContact(projectId, memberId, contactId, roleId, updatedContact) {
    return this.http.put('/projects/' + projectId + '/members/' + memberId + '/contacts/' + contactId + '/roles/' + roleId, updatedContact)
      .map((res) => res.json());
  }

  //////////////////////////////////////////////////////////////////////////////

  /**
  * Organizations - Contacts
  * Resources for getting and manipulating contacts of organizations
  **/

  getOrganizationContactTypes() {
    var response = this.http.get(this.baseUrl + '/organizations/contacts/types')
      .map(res => res.json());
    return response;
  }

  getOrganizationContacts(organizationId) {
    var response = this.http.get(this.baseUrl + '/organizations/' + organizationId + '/contacts')
      .map(res => res.json());
    return response;
  }

  createOrganizationContact(organizationId, newContact) {
    return this.http.post('/organizations/' + organizationId + '/contacts', newContact)
      .map((res) => res.json());
  }

  getOrganizationContact(organizationId, contactId) {
    var response = this.http.get(this.baseUrl + '/organizations/' + organizationId + '/contacts/' + contactId)
      .map(res => res.json());
    return response;
  }

  updateOrganizationContact(organizationId, contactId, contact) {
    return this.http.put('/organizations/' + organizationId + '/contacts/' + contactId, contact)
      .map((res) => res.json());
  }

  //////////////////////////////////////////////////////////////////////////////

  /**
  * Organizations - Projects:
  * Resources for getting details about an organizations project membership
  **/

  getOrganizationProjectMemberships(organizationId) {
    var response = this.http.get(this.baseUrl + '/organizations/' + organizationId + '/projects_member')
      .map(res => res.json());
    return response;
  }

  //////////////////////////////////////////////////////////////////////////////

  /**
  * Users
  * Resources to manage internal LF users and roles
  **/

  getCurrentUser() {
    return this.http.get(this.baseUrl + '/user')
      .map(res => res.json());
  }

  getAllUsers() {
    return this.http.get(this.baseUrl + '/users')
      .map(res => res.json());
  }

  createUser(user) {
    return this.http.post('/users', { user: user })
      .map(res => res.json());
  }

  removeUser(userId) {
    return this.http.delete(this.baseUrl + '/users/' + userId)
      .map(res => res.json());
  }

  getUser(userId) {
    return this.http.get(this.baseUrl + '/users/' + userId)
      .map(res => res.json());
  }

  getUserRoles() {
    return this.http.get(this.baseUrl + '/users/roles')
            .map(res => res.json());
  }

  updateUser(userId, user) {
    return this.http.put('/users/' + userId, user)
      .map(res => res.json());
  }

  addUserRole(userId, role) {
    return this.http.post('/users/' + userId + '/role', { role: role })
      .map(res => res.json());
  }

  removeUserRole(userId, roleId) {
    return this.http.delete('/users/' + userId + '/role/' + roleId)
      .map(res => res.json());
  }

  //////////////////////////////////////////////////////////////////////////////

}
