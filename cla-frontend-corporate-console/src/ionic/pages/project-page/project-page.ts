import { Component } from "@angular/core";
import {
  NavController,
  ModalController,
  NavParams,
  IonicPage
} from "ionic-angular";
import { ClaService } from "../../services/cla.service";
import { ClaCompanyModel } from "../../models/cla-company";
import { ClaUserModel } from "../../models/cla-user";
import { ClaSignatureModel } from "../../models/cla-signature";
import { SortService } from "../../services/sort.service";
import { RolesService } from "../../services/roles.service";
import { Restricted } from "../../decorators/restricted";

@Restricted({
  roles: ["isAuthenticated"]
})
@IonicPage({
  segment: "company/:companyId/project/:projectId"
})
@Component({
  selector: "project-page",
  templateUrl: "project-page.html"
})
export class ProjectPage {
  signatures: ClaSignatureModel[];
  loading: any;
  companyId: string;
  projectId: string;
  company: ClaCompanyModel;
  manager: ClaUserModel;

  project: any;
  users: any;

  sort: any;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private claService: ClaService,
    public modalCtrl: ModalController,
    private rolesService: RolesService, // for @Restricted
    private sortService: SortService
  ) {
    this.companyId = navParams.get("companyId");
    this.projectId = navParams.get("projectId");
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

  }

  ngOnInit() {
    this.getProject();
    this.getProjectSignatures();
    this.getCompany();
  }

  getCompany() {
    this.claService.getCompany(this.companyId).subscribe(response => {
      this.company = response;
      this.getUser(this.company.company_manager_id);
    });
  }

  getProject() {
    this.claService.getProject(this.projectId).subscribe(response => {
      this.project = response;
    });
  }

  getProjectSignatures() {
    // TODO: remove this comment when EP is working correctly. currently returning cclas instead of employee clas. reported.
    this.claService
      .getCompanyProjectSignatures(this.companyId, this.projectId)
      .subscribe(response => {
        this.signatures = response.filter(sig => sig.signature_type === "cla");
        for (let signature of this.signatures) {
          this.getUser(signature.signature_reference_id);
        }
      });
  }

  getUser(userId) {
    if (!this.users[userId]) {
      this.claService.getUser(userId).subscribe(response => {
        this.users[userId] = response;
        this.manager = response;
      });
    }
  }

  openWhitelistEmailModal() {
    let modal = this.modalCtrl.create("WhitelistModal", {
      type: "email",
      company: this.company
    });
    modal.onDidDismiss(data => {
      // A refresh of data anytime the modal is dismissed
      this.getCompany();
    });
    modal.present();
  }

  openWhitelistDomainModal() {
    let modal = this.modalCtrl.create("WhitelistModal", {
      type: "domain",
      company: this.company
    });
    modal.onDidDismiss(data => {
      // A refresh of data anytime the modal is dismissed
      this.getCompany();
    });
    modal.present();
  }

  sortMembers(prop) {
    this.sortService.toggleSort(this.sort, prop, this.signatures);
  }
}
