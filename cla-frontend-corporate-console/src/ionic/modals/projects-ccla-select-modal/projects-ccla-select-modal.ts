import { Component, ChangeDetectorRef } from "@angular/core";
import {
  NavController,
  NavParams,
  ModalController,
  ViewController,
  AlertController,
  IonicPage
} from "ionic-angular";
import { FormBuilder, FormGroup, Validators } from "@angular/forms";
import { ClaService } from "../../services/cla.service";
import { ClaCompanyModel } from "../../models/cla-company";

@IonicPage({
  segment: "projects-ccla-select-modal"
})
@Component({
  selector: "projects-ccla-select-modal",
  templateUrl: "projects-ccla-select-modal.html"
})
export class ProjectsCclaSelectModal {
  projects: any;
  company: ClaCompanyModel;

  constructor(
    public navParams: NavParams,
    public navCtrl: NavController,
    public viewCtrl: ViewController,
    public formBuilder: FormBuilder,
    private claService: ClaService
  ) {
    this.getDefaults();
  }

  getDefaults() {
    this.company = this.navParams.get("company");
  }

  ngOnInit() {
    this.getProjectsCcla();
  }

  getProjectsCcla() {
    this.claService.getProjects().subscribe(response => {
      this.projects = response;
    });
    // TODO: update when endpoint is ready
    // this.claService.getProjectsCcla().subscribe((response) => {
    //   this.projects = response;
    // });
  }

  selectProject(project) {
    this.navCtrl.push("ClaCorporatePage", {
      projectId: project.project_id,
      company: this.company
    });
  }

  dismiss() {
    this.viewCtrl.dismiss();
  }
}
