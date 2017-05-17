import { Injectable } from '@angular/core';
import { Http, Headers } from '@angular/http';

import 'rxjs/Rx';

@Injectable()
export class CincoService{
  http: any;
  baseUrl: String;

  constructor(http: Http) {
    this.http = http;
    this.baseUrl = '';
  }

  postProject(newProject) {
    let headers = new Headers({ 'Content-Type': 'application/json' });
    let body = new FormData();
    body.append('project_name', newProject.project_name);
    // body.append('project_type', newProject.project_type);
    return this.http.post('/post_project', body, headers)
                .map((res) => res.json());
  }

  editProject(editProject) {
    let headers = new Headers({ 'Content-Type': 'application/json' });
    let body = new FormData();
    body.append('project_name', editProject.project_name);
    // body.append('project_type', newProject.project_type);
    return this.http.post('/edit_project', body, headers)
                .map((res) => res.json());
  }

  /*
    Projects:
    Resources to expose and manipulate details of projects
   */

   getAllProjects() {
     return this.http.get(this.baseUrl + '/projects')
             .map(res => res.json());
   }

   getProject(projectId) {
     return this.http.get(this.baseUrl + '/projects/' + projectId)
             .map(res => res.json());
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
    // body.append('boardMember', contact.boardMember);
    // body.append('primaryContact', contact.primaryContact);
    body.append('contactId', contact.contact.id);
    body.append('contactGivenName', contact.contact.givenName);
    body.append('contactFamilyName', contact.contact.familyName);
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
    // body.append('boardMember', contact.boardMember);
    // body.append('primaryContact', contact.primaryContact);
    body.append('contactId', contact.contact.id);
    body.append('contactGivenName', contact.contact.givenName);
    body.append('contactFamilyName', contact.contact.familyName);
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
    body.append('bio', contact.bio);
    body.append('email', contact.email);
    body.append('phone', contact.phone);
    return this.http.put('/organizations/' + organizationId + '/contacts/' + contactId, body, headers)
                .map((res) => res.json());
  }
}
