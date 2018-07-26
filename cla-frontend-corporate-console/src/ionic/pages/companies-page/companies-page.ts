import { Component } from "@angular/core";
import { NavController, ModalController, IonicPage } from "ionic-angular";
import { ClaService } from "../../services/cla.service";
import { ClaCompanyModel } from "../../models/cla-company";
import { RolesService } from "../../services/roles.service";
import { Restricted } from "../../decorators/restricted";

@Restricted({
  roles: ["isAuthenticated"]
})
@IonicPage({
  segment: "companies"
})
@Component({
  selector: "companies-page",
  templateUrl: "companies-page.html"
})
export class CompaniesPage {
  loading: any;
  companies: any;

  constructor(
    public navCtrl: NavController,
    private claService: ClaService,
    public modalCtrl: ModalController,
    private rolesService: RolesService // for @Restricted
  ) {
    this.getDefaults();
  }

  getDefaults() {
    this.loading = {
      companies: true
    };

    this.companies = [];
  }

  ngOnInit() {
    this.getCompanies();
  }

  openCompanyModal() {
    let modal = this.modalCtrl.create("AddCompanyModal", {});
    modal.onDidDismiss(data => {
      // A refresh of data anytime the modal is dismissed
      this.getCompanies();
    });
    modal.present();
  }

  getCompanies() {
    this.claService.getCompanies().subscribe(response => {
      this.companies = response;
      this.loading.companies = false;
    });
  }

  viewCompany(companyId) {
    this.navCtrl.setRoot("CompanyPage", {
      companyId: companyId
    });
  }
}
