import { Injectable } from '@angular/core';
import { Http, Headers, RequestOptions } from '@angular/http';

import 'rxjs/Rx';

@Injectable()
export class CincoService{
  http: any;
  baseUrl: String;

  constructor(http: Http) {
    this.http = http;
    this.baseUrl = '';
  }

  /*
    Projects:
    Resources to expose and manipulate details of projects
   */
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

   getProject(projectId, getMembers) {
     if (getMembers) { projectId = projectId + '?members=true' ; }
     return this.http.get(this.baseUrl + '/projects/' + projectId)
             .map(res => res.json());
   }

   postProject(newProject) {
     var headers = new Headers();
     headers.append("Accept", 'application/json');
     headers.append('Content-Type', 'application/json' );
     let options = new RequestOptions({ headers: headers });
     return this.http.post('/projects', newProject, options)
                 .map((res) => res.json());
   }

   editProject(projectId, editProject) {
     var headers = new Headers();
     headers.append("Accept", 'application/json');
     headers.append('Content-Type', 'application/json' );
     let options = new RequestOptions({ headers: headers });
     return this.http.post('/edit_project/' + projectId, editProject, options)
                 .map((res) => res.json());
   }

   getProjectConfig(projectId) {
     return this.http.get(this.baseUrl + '/projects/' + projectId + '/config')
             .map(res => res.json());
   }

   updateProjectManagers(projectId, managers) {
     var headers = new Headers();
     headers.append("Accept", 'application/json');
     headers.append('Content-Type', 'application/json' );
     let options = new RequestOptions({ headers: headers });
     return this.http.put('/projects/' + projectId + '/managers', managers, options)
                 .map((res) => res.json());
   }

  /*
    Projects - Members:
    Resources for getting details about project members
   */

  getProjectMembers(projectId) {
    return this.http.get(this.baseUrl + '/projects/' + projectId + '/members')
            .map(res => res.json());
  }

  getMember(projectId, memberId) {
    var response = this.http.get(this.baseUrl + '/projects/' + projectId + '/members/' + memberId)
            .map(res => res.json());
    return response;
  }

  /*
    Projects - Members - Contacts:
    Resources for getting and manipulating contacts of project members
   */

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

  addMemberContact(projectId, memberId, contactId, contact) {
    let headers = new Headers({ 'Content-Type': 'application/json' });
    let body = new FormData();
    body.append('id', contact.id);
    body.append('memberId', memberId);
    body.append('type', contact.type);
    body.append('boardMember', contact.boardMember);
    body.append('primaryContact', contact.primaryContact);
    body.append('contactId', contact.contact.id);
    body.append('contactGivenName', contact.contact.givenName);
    body.append('contactFamilyName', contact.contact.familyName);
    body.append('contactTitle', contact.contact.title);
    body.append('contactBio', contact.contact.bio);
    body.append('contactEmail', contact.contact.email);
    body.append('contactPhone', contact.contact.phone);
    body.append('contactHeadshotRef', contact.contact.headshotRef);
    body.append('contactType', contact.contact.type);
    return this.http.post('/projects/' + projectId + '/members/' + memberId + '/contacts/' + contactId, body, headers)
                .map((res) => res.json());
  }

  removeMemberContact(projectId, memberId, contactId, roleId) {
    let headers = new Headers({ 'Content-Type': 'application/json' });
    let body = new FormData();
    return this.http.delete('/projects/' + projectId + '/members/' + memberId + '/contacts/' + contactId + '/roles/' + roleId, body, headers)
                .map((res) => res.json());
  }

  updateMemberContact(projectId, memberId, contactId, roleId, contact) {
    let headers = new Headers({ 'Content-Type': 'application/json' });
    let body = new FormData();
    body.append('id', contact.id);
    body.append('memberId', memberId);
    body.append('type', contact.type);
    body.append('boardMember', contact.boardMember);
    body.append('primaryContact', contact.primaryContact);
    body.append('contactId', contact.contact.id);
    body.append('contactGivenName', contact.contact.givenName);
    body.append('contactFamilyName', contact.contact.familyName);
    body.append('contactTitle', contact.contact.title);
    body.append('contactBio', contact.contact.bio);
    body.append('contactEmail', contact.contact.email);
    body.append('contactPhone', contact.contact.phone);
    body.append('contactHeadshotRef', contact.contact.headshotRef);
    body.append('contactType', contact.contact.type);
    return this.http.put('/projects/' + projectId + '/members/' + memberId + '/contacts/' + contactId + '/roles/' + roleId, body, headers)
                .map((res) => res.json());
  }

  /*
    Organizations - Contacts:
    Resources for getting and manipulating contacts of organizations
   */

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

  createOrganizationContact(organizationId, contact) {
    let headers = new Headers({ 'Content-Type': 'application/json' });
    let body = new FormData();
    body.append('type', contact.type);
    body.append('givenName', contact.givenName);
    body.append('familyName', contact.familyName);
    body.append('title', contact.title);
    body.append('bio', contact.bio);
    body.append('email', contact.email);
    body.append('phone', contact.phone);
    return this.http.post('/organizations/' + organizationId + '/contacts', body, headers)
                .map((res) => res.json());
  }

  getOrganizationContact(organizationId, contactId) {
    var response = this.http.get(this.baseUrl + '/organizations/' + organizationId + '/contacts/' + contactId)
            .map(res => res.json());
    return response;
  }

  updateOrganizationContact(organizationId, contactId, contact) {
    let headers = new Headers({ 'Content-Type': 'application/json' });
    let body = new FormData();
    body.append('type', contact.type);
    body.append('givenName', contact.givenName);
    body.append('familyName', contact.familyName);
    body.append('title', contact.title);
    body.append('bio', contact.bio);
    body.append('email', contact.email);
    body.append('phone', contact.phone);
    return this.http.put('/organizations/' + organizationId + '/contacts/' + contactId, body, headers)
                .map((res) => res.json());
  }

  /*
    Organizations - Projects:
    Resources for getting details about an organizations project membership
   */

  getOrganizationProjectMemberships(organizationId) {
    var response = this.http.get(this.baseUrl + '/organizations/' + organizationId + '/projects_member')
            .map(res => res.json());
    return response;
  }

  /*
    Users:
    Resources to manage internal LF users and roles
   */

  getCurrentUser() {
    return this.http.get(this.baseUrl + '/user')
            .map(res => res.json());
  }

  getAllUsers() {
    return this.http.get(this.baseUrl + '/users')
            .map(res => res.json());
  }

  getUser(userId) {
    return this.http.get(this.baseUrl + '/users/' + userId)
            .map(res => res.json());
  }

  updateUser(userId, user) {
    let headers = new Headers({ 'Content-Type': 'application/json' });
    let body = new FormData();
    body.append('userId', user.userId);
    body.append('email', user.email);
    body.append('calendar', user.calendar);
    return this.http.put('/users/' + userId, body, headers)
            .map(res => res.json());
  }

}
