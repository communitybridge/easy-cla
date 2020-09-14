// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { Component } from '@angular/core';
import { NavController, NavParams, IonicPage } from 'ionic-angular';
import { ClaService } from '../../services/cla.service';
import { RolesService } from '../../services/roles.service';
import { AuthService } from '../../services/auth.service';
import { Restricted } from '../../decorators/restricted';
import { generalConstants } from '../../constants/general';

@Restricted({
  roles: ['isAuthenticated']
})
@IonicPage({
  segment: 'cla/gerrit/project/:gerritId/individual'
})
@Component({
  selector: 'cla-gerrit-individual',
  templateUrl: 'cla-gerrit-individual.html'
})
export class ClaGerritIndividualPage {
  gerritId: string;
  projectId: string;
  project: any;
  gerrit: any;
  userId: string;
  user: any;
  signatureIntent: any;
  activeSignatures: boolean = true; // we assume true until otherwise
  signature: any;

  userRoles: any;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private claService: ClaService,
    private rolesService: RolesService,
    private authService: AuthService,
  ) {
    this.getDefaults();
    this.gerritId = navParams.get('gerritId');
    localStorage.setItem('gerritId', this.gerritId);
    localStorage.setItem('gerritClaType', 'ICLA');
  }

  getDefaults() {
    this.userRoles = this.rolesService.userRoleDefaults;

    this.project = {
      project_name: ''
    };
    this.signature = {
      sign_url: ''
    };
  }

  ngOnInit() {
    this.getProject(this.gerritId);
  }

  ionViewCanEnter() {
    if (!this.authService.isAuthenticated) {
      setTimeout(() => this.navCtrl.setRoot('LoginPage'));
    }
    return this.authService.isAuthenticated;
  }

  ngAfterViewInit() { }

  getProject(gerritId) {
    //retrieve projectId from this Gerrit
    this.claService.getGerrit(gerritId).subscribe((gerrit) => {
      this.gerrit = gerrit;
      this.projectId = gerrit.project_id;

      //retrieve project info with project Id
      this.claService.getProjectWithAuthToken(gerrit.project_id).subscribe((project) => {
        this.project = project;
        localStorage.setItem(generalConstants.PROJECT_MODEL, JSON.stringify(project));
        // retrieve userInfo from auth0 service
        this.claService.postOrGetUserForGerrit().subscribe((user) => {
          this.userId = user.user_id;
          localStorage.setItem(generalConstants.USER_MODEL, JSON.stringify(user));
          // get signatureIntent object, similar to the Github flow.
          this.postSignatureRequest();
        });
      });
    });
  }

  postSignatureRequest() {
    let signatureRequest = {
      project_id: this.projectId,
      user_id: this.userId,
      return_url_type: 'Gerrit'
    };
    this.claService.postIndividualSignatureRequest(signatureRequest).subscribe((response) => {
      this.signature = response;
    });
  }

  openClaAgreement() {
    if (!this.signature.sign_url) {
      return;
    }
    window.open(this.signature.sign_url, '_self');
  }
}
