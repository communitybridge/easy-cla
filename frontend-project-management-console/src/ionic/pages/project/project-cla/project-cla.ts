import {Component, ViewChild} from "@angular/core";
import {
  NavController,
  ModalController,
  NavParams,
  IonicPage, Nav, Events
} from "ionic-angular";
import { CincoService } from "../../../services/cinco.service";
import { KeycloakService } from "../../../services/keycloak/keycloak.service";
import { SortService } from "../../../services/sort.service";
import { PopoverController } from "ionic-angular";
import { ClaService } from "../../../services/cla.service";
import { RolesService } from "../../../services/roles.service";
import { Restricted } from "../../../decorators/restricted";

@Restricted({
  roles: ["isAuthenticated", "isPmcUser"]
})
@IonicPage({
  segment: "project/:projectId/cla"
})
@Component({
  selector: "project-cla",
  templateUrl: "project-cla.html"
})
export class ProjectClaPage {
  loading: any;

  projectId: string;

  claProjects: any;

  iclaUploadInfo: any;
  cclaUploadInfo: any;
  @ViewChild(Nav) nav: Nav;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private cincoService: CincoService,
    private sortService: SortService,
    public modalCtrl: ModalController,
    private keycloak: KeycloakService,
    private popoverCtrl: PopoverController,
    public claService: ClaService,
    public rolesService: RolesService,
    public events: Events
  ) {
    this.projectId = navParams.get("projectId");
    this.getDefaults();
  }

  getDefaults() {
    this.loading = {
      claProjects: true
    };
    this.claProjects = [];
  }

  ngOnInit() {
    this.getClaProjects();
  }

  getClaProjects() {
    this.loading.claProjects = true;
    this.claService
      .getProjectsByExternalId(this.projectId)
      .subscribe(projects => {
        this.claProjects = projects;
        this.loading.claProjects = false;
        for (let project of projects) {

          //Get Github Orgs. 
          this.claService
            .getProjectOrganizations(project.project_id)
            .subscribe(organizations => {
              project.organizations = organizations;
              for (let organization of organizations) {
                this.claService
                  .getGithubGetNamespace(organization.organization_name)
                  .subscribe(providerInfo => {
                    organization.providerInfo = providerInfo;
                  });
                if (organization.organization_installation_id) {
                  this.claService
                    .getGithubOrganizationRepositories(
                      organization.organization_name
                    )
                    .subscribe(repositories => {
                      organization.repositories = repositories;
                    });
                }
              }
            });

            //Get Gerrit Instances
            this.claService
              .getGerritInstance(project.project_id)
              .subscribe(gerrits => {
                project.gerrits = gerrits;
            });
        }
      });
  }

  backToProjects () {
    this.events.publish('nav:allProjects');
  }

  openClaContractConfigModal(claProject) {
    let modal;
    if (claProject) {
      modal = this.modalCtrl.create("ClaContractConfigModal", {
        claProject: claProject
      });
    } else {
      modal = this.modalCtrl.create("ClaContractConfigModal", {
        projectId: this.projectId
      });
    }
    modal.onDidDismiss(data => {
      this.getClaProjects();
    });
    modal.present();
  }

  openClaContractUploadModal(claProjectId, documentType) {
    let modal = this.modalCtrl.create("ClaContractUploadModal", {
      claProjectId: claProjectId,
      documentType: documentType
    });
    modal.onDidDismiss(data => {
      this.getClaProjects();
    });
    modal.present();
  }

  openClaViewSignaturesModal(project_id) {
    let modal = this.modalCtrl.create("ClaContractViewSignaturesModal", {
      claProjectId: project_id,
    }, {
      cssClass: 'medium'
    });
    modal.onDidDismiss(data => {
      this.getClaProjects();
    });
    modal.present();
  }

  openClaContractVersionModal(claProjectId, documentType, documents) {
    let modal = this.modalCtrl.create("ClaContractVersionModal", {
      claProjectId: claProjectId,
      documentType: documentType,
      documents: documents
    });
    modal.present();
  }

  openClaOrganizationProviderModal(claProjectId) {
    let modal = this.modalCtrl.create("ClaOrganizationProviderModal", {
      claProjectId: claProjectId
    });
    modal.onDidDismiss(data => {
      this.getClaProjects();
    });
    modal.present();
  }

  openClaGerritModal(projectId) {
    let modal = this.modalCtrl.create("ClaGerritModal", {
      projectId: projectId
    });
    modal.onDidDismiss(data => {
      this.getClaProjects();
    });
    modal.present();
  }

  openClaOrganizationAppModal(orgName) {
    let modal = this.modalCtrl.create("ClaOrganizationAppModal", {
      orgName: orgName
    });
    modal.onDidDismiss(data => {
      this.getClaProjects();
    });
    modal.present();
  }

  openClaContractCompaniesModal(claProjectId) {
    let modal = this.modalCtrl.create("ClaContractCompaniesModal", {
      claProjectId: claProjectId
    });
    modal.present();
  }

  openClaContractsContributorsPage(claProjectId) {
    this.navCtrl.push("ClaContractsContributorsPage", {
      claProjectId: claProjectId
    });
  }

  organizationPopover(ev, organization) {
    let actions = {
      items: [
        {
          label: "Delete",
          callback: "deleteClaGithubOrganization",
          callbackData: {
            organization: organization
          }
        }
      ]
    };
    let popover = this.popoverCtrl.create("ActionPopoverComponent", actions);

    popover.present({
      ev: ev
    });

    popover.onDidDismiss(popoverData => {
      if (popoverData) {
        this.popoverResponse(popoverData);
      }
    });
  }

  deleteClaGithubOrganization(organization) {
    this.claService
      .deleteGithubOrganization(organization.organization_name)
      .subscribe(response => {
        this.getClaProjects();
      });
  }


  deleteGerritInstance(gerrit) {
    this.claService
      .deleteGerritInstance(gerrit.gerrit_id)
      .subscribe(response => {
        this.getClaProjects();
      });
  }

  /**
   * Called if popover dismissed with data. Passes data to a callback function
   * @param  {object} popoverData should contain .callback and .callbackData
   */
  popoverResponse(popoverData) {
    let callback = popoverData.callback;
    if (this[callback]) {
      this[callback](popoverData.callbackData);
    }
  }
}
