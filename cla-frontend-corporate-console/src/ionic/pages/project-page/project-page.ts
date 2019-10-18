// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import {Component} from "@angular/core";
import {AlertController, IonicPage, ModalController, NavController, NavParams} from "ionic-angular";
import {ClaService} from "../../services/cla.service";
import {ClaCompanyModel} from "../../models/cla-company";
import {ClaUserModel} from "../../models/cla-user";
import {ClaSignatureModel} from "../../models/cla-signature";
import {ClaManager} from "../../models/cla-manager";
import {SortService} from "../../services/sort.service";
import {RolesService} from "../../services/roles.service";
import {Restricted} from "../../decorators/restricted";
import {WhitelistModal} from "../../modals/whitelist-modal/whitelist-modal";

@Restricted({
  roles: ["isAuthenticated"]
})
@IonicPage({
  segment: "company/:companyId/project/:projectId/:modal"
})
@Component({
  selector: "project-page",
  templateUrl: "project-page.html"
})
export class ProjectPage {
  cclaSignature: any;
  employeeSignatures: any[];
  githubOrgWhitelist: any[] = [];
  githubEnabledWhitelist: any[] = [];
  loading: any;
  companyId: string;
  projectId: string;
  managers: ClaManager[];
  company: ClaCompanyModel;
  manager: ClaUserModel;
  showModal: any;

  project: any;
  users: any;

  sort: any;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private claService: ClaService,
    public alertCtrl: AlertController,
    public modalCtrl: ModalController,
    private rolesService: RolesService, // for @Restricted
    private sortService: SortService
  ) {
    this.companyId = navParams.get("companyId");
    this.projectId = navParams.get("projectId");
    this.showModal = navParams.get("modal");

    if (this.showModal === 'orgwhitelist') {
      this.openGithubOrgWhitelistModal();
    }

    this.getDefaults();
  }

  getDefaults() {
    this.loading = {};
    this.users = {};
    this.sort = {
      date: {
        arrayProp: "date_modified",
        sortType: "date",
        sort: null
      }
    };
    this.company = new ClaCompanyModel();
    this.cclaSignature = new ClaSignatureModel();
  }

  ngOnInit() {
    this.getProject();
    this.getCompany();
  }

  getCompany() {
    this.claService.getCompany(this.companyId).subscribe(response => {
      this.company = response;
      this.getManager(this.company.company_manager_id);
    });
  }

  getProject() {
    // console.log('Loading project: ' + this.projectId);
    this.claService.getProject(this.projectId).subscribe(response => {
      console.log('Project response:');
      console.log(response);
      this.project = response;
      this.getProjectSignatures();
    });
  }

  getGitHubOrgWhitelist() {
    this.claService.getGithubOrganizationWhitelistEntries(this.cclaSignature.signatureID).subscribe(organizations => {
      this.githubOrgWhitelist = organizations;
      this.githubOrgWhitelistEnabled();
    });
  }

  githubOrgWhitelistEnabled() {
    this.githubEnabledWhitelist = this.githubOrgWhitelist.filter((org) => org.selected);
  }

  getCLAManagers() {
    this.claService.getCLAManagers(this.cclaSignature.signatureID).subscribe(response => {
      this.managers = response;
    });
  }

  getProjectSignatures() {
    // get CCLA signatures
    this.claService.getCompanyProjectSignatures(this.companyId, this.projectId)
      .subscribe(response => {
          // console.log('Project signatures:');
          // console.log(response);
          if (response.signatures) {
            let cclaSignatures = response.signatures.filter(sig => sig.signatureType === 'ccla');
            if (cclaSignatures.length) {
              console.log(cclaSignatures);
              this.cclaSignature = cclaSignatures[0];

              // Sort the values
              if (this.cclaSignature.domainWhitelist) {
                const sortedList: string[] = this.cclaSignature.domainWhitelist.sort((a, b) => {
                  return a.trim().localeCompare(b.trim())
                });
                // Remove duplicates - set doesn't allow dups
                this.cclaSignature.domainWhitelist = Array.from(new Set(sortedList));
              }
              if (this.cclaSignature.emailWhitelist) {
                const sortedList: string[] = this.cclaSignature.emailWhitelist.sort((a, b) => {
                  return a.trim().localeCompare(b.trim())
                });
                // Remove duplicates - set doesn't allow dups
                this.cclaSignature.emailWhitelist = Array.from(new Set(sortedList));
              }
              if (this.cclaSignature.githubWhitelist) {
                const sortedList: string[] = this.cclaSignature.githubWhitelist.sort((a, b) => {
                  return a.trim().localeCompare(b.trim())
                });
                // Remove duplicates - set doesn't allow dups
                this.cclaSignature.githubWhitelist = Array.from(new Set(sortedList));
              }
              if (this.cclaSignature.githubOrgWhitelist) {
                const sortedList: string[] = this.cclaSignature.githubOrgWhitelist.sort((a, b) => {
                  return a.trim().localeCompare(b.trim())
                });
                // Remove duplicates - set doesn't allow dups
                this.cclaSignature.githubOrgWhitelist = Array.from(new Set(sortedList));
              }
              this.getCLAManagers();
              this.getGitHubOrgWhitelist();
            }
          }
        },
        exception => {
          console.log("Exception while calling: getCompanyProjectSignatures() for company ID: " +
            this.companyId + ' and project ID: ' + this.projectId);
          console.log(exception);
        });

    // get employee signatures
    this.claService.getEmployeeProjectSignatures(this.companyId, this.projectId)
      .subscribe(response => {
          console.log('Employee signatures:');
          console.log(response);
          if (response.signatures) {
            const sigs = response.signatures || [];
            this.employeeSignatures = sigs.sort((a, b) => {
              if (a.userName != null && b.userName != null) {
                return a.userName.trim().localeCompare(b.userName.trim())
              } else {
                return 0;
              }
            });
          }
        },
        exception => {
          console.log("Exception while calling: getEmployeeProjectSignatures() for company ID: " +
            this.companyId + ' and project ID: ' + this.projectId);
          console.log(exception);
        });
  }

  getManager(userId) {
    console.log('Looking up manager: ' + userId);
    this.claService.getUser(userId).subscribe(response => {
      this.manager = response;
    });
  }

  openWhitelistEmailModal() {
    let modal = this.modalCtrl.create("WhitelistModal", {
      type: "email",
      projectName: this.project.project_name,
      companyName: this.cclaSignature.companyName,
      projectId: this.cclaSignature.projectID,
      companyId: this.companyId,
      signatureId: this.cclaSignature.signatureID,
      whitelist: this.cclaSignature.emailWhitelist
    });
    modal.onDidDismiss(data => {
      // A refresh of data anytime the modal is dismissed
      this.getProjectSignatures();
    });
    modal.present();
  }

  openWhitelistDomainModal() {
    let modal = this.modalCtrl.create("WhitelistModal", {
      type: "domain",
      projectName: this.project.project_name,
      companyName: this.cclaSignature.companyName,
      projectId: this.cclaSignature.projectID,
      signatureId: this.cclaSignature.signatureID,
      whitelist: this.cclaSignature.domainWhitelist
    });
    modal.onDidDismiss(data => {
      // A refresh of data anytime the modal is dismissed
      this.getProjectSignatures();
    });
    modal.present();
  }

  openWhitelistGithubModal() {
    let modal = this.modalCtrl.create("WhitelistModal", {
      type: "github",
      projectName: this.project.project_name,
      companyName: this.cclaSignature.companyName,
      projectId: this.cclaSignature.projectID,
      signatureId: this.cclaSignature.signatureID,
      whitelist: this.cclaSignature.githubWhitelist
    });
    modal.onDidDismiss(data => {
      // A refresh of data anytime the modal is dismissed
      this.getProjectSignatures();
    });
    modal.present();

  }

  sortMembers(prop) {
    this.sortService.toggleSort(this.sort, prop, this.employeeSignatures);
  }

  deleteManagerConfirmation(payload: ClaManager) {
    let alert = this.alertCtrl.create({
      subTitle: `Delete Manager`,
      message: `Are you sure you want to delete this Manager?`,
      buttons: [
        {
          text: 'Cancel',
          role: 'cancel',
          cssClass: 'secondary',
          handler: () => {
          }
        }, {
          text: 'Delete',
          handler: () => {
            this.deleteManager(payload);
          }
        }
      ]
    });

    alert.present();
  }

  deleteManager(payload: ClaManager) {
    this.claService.deleteCLAManager(this.cclaSignature.signatureID, payload)
      .subscribe(() => this.getCLAManagers())
  }

  openManagerModal() {
    let modal = this.modalCtrl.create("AddManagerModal", {
      signatureId: this.cclaSignature.signatureID
    });
    modal.onDidDismiss(data => {
      if (data) {
        this.getCLAManagers();
      }
    });
    modal.present();
  }

  openGithubOrgWhitelistModal() {
    // Maybe this happens if we authorize GH and come back after the flow and we haven't fully loaded...
    if (this.cclaSignature == null) {
      this.claService.getCompanyProjectSignatures(this.companyId, this.projectId)
        .subscribe(response => {
            let cclaSignatures = response.filter(sig => sig.signatureType === 'ccla');
            if (cclaSignatures.length) {
              this.cclaSignature = cclaSignatures[0];
              console.log('CCLA Signature:');
              console.log(this.cclaSignature);
              this.getCLAManagers();
              this.getGitHubOrgWhitelist();

              // Ok to open the modal now that we have signatures loaded
              let modal = this.modalCtrl.create("GithubOrgWhitelistModal", {
                projectName: this.project.project_name,
                companyName: this.cclaSignature.companyName,
                projectId: this.cclaSignature.projectID,
                companyId: this.companyId,
                corporateClaId: this.projectId,
                signatureId: this.cclaSignature.signatureID
              });
              modal.onDidDismiss(data => {
                // Refresh the list
                this.getGitHubOrgWhitelist();
              });
              modal.present();
            }
          },
          exception => {
            console.log("Exception while calling: getCompanyProjectSignatures() for company ID: " +
              this.companyId + ' and project ID: ' + this.projectId);
            console.log(exception);
          });
    } else {
      let modal = this.modalCtrl.create("GithubOrgWhitelistModal", {
        projectName: this.project.project_name,
        companyName: this.cclaSignature.companyName,
        projectId: this.cclaSignature.projectID,
        companyId: this.companyId,
        corporateClaId: this.projectId,
        signatureId: this.cclaSignature.signatureID
      });
      modal.onDidDismiss(data => {
        // Refresh the list
        this.getGitHubOrgWhitelist();
      });
      modal.present();
    }
  }
}
