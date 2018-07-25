import { Component } from '@angular/core';
import { NavController, NavParams, ViewController, IonicPage, } from 'ionic-angular';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ClaService } from '../../services/cla.service'

@IonicPage({
  segment: 'cla-contract-config-modal'
})
@Component({
  selector: 'cla-contract-config-modal',
  templateUrl: 'cla-contract-config-modal.html',
})
export class ClaContractConfigModal {
  form: FormGroup;
  submitAttempt: boolean = false;
  currentlySubmitting: boolean = false;

  projectId: string;
  claProject: any;
  newClaProject: boolean;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    public viewCtrl: ViewController,
    private formBuilder: FormBuilder,
    private claService: ClaService,
  ) {
    this.projectId = this.navParams.get('projectId');
    console.log("projectId" + this.projectId);
    this.claProject = this.navParams.get('claProject');
    this.getDefaults();
    this.form = formBuilder.group({
      name:[this.claProject.project_name, Validators.compose([Validators.required])],
      ccla:[this.claProject.project_ccla_enabled],
      cclaAndIcla:[this.claProject.project_ccla_requires_icla_signature],
      icla:[this.claProject.project_icla_enabled],
    });
  }

  getDefaults() {
    this.newClaProject = false; // we assume we have an existing cla project
    // if claProject is not passed
    if (!this.claProject) {
      this.newClaProject = true; // change to creating new project
      this.claProject = {
        project_external_id: this.projectId,
        project_name: '',
        project_ccla_enabled: false,
        project_ccla_requires_icla_signature: false,
        project_icla_enabled: false,
      };
    }
  }

  ngOnInit() {

  }

  submit() {
    this.submitAttempt = true;
    this.currentlySubmitting = true;
    if (!this.form.valid) {
      this.currentlySubmitting = false;
      // prevent submit
      return;
    }
    if (this.newClaProject) {
      console.log('post');
      this.postProject();
    } else {
      console.log('put');
      this.putProject();
    }
  }

  postProject() {
    let claProject = {
      project_external_id: this.claProject.project_external_id,
      project_name: this.form.value.name,
      project_ccla_enabled: this.form.value.ccla,
      project_ccla_requires_icla_signature: this.form.value.cclaAndIcla,
      project_icla_enabled: this.form.value.icla,
    };
    this.claService.postProject(claProject).subscribe((response) => {
      this.dismiss();
    });
  }

  putProject() {
    // rebuild the claProject object from existing data and form data
    let claProject = {
      project_id: this.claProject.project_id,
      project_external_id: this.claProject.project_external_id,
      project_name: this.form.value.name,
      project_ccla_enabled: this.form.value.ccla,
      project_ccla_requires_icla_signature: this.form.value.cclaAndIcla,
      project_icla_enabled: this.form.value.icla,
    };
    this.claService.putProject(claProject).subscribe((response) => {
      this.dismiss();
    });
  }

  dismiss() {
    this.viewCtrl.dismiss();
  }

}
