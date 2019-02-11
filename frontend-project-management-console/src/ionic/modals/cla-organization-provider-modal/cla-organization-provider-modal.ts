import { Component } from '@angular/core';
import { NavController, NavParams, ViewController, IonicPage, Events } from 'ionic-angular';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ClaService } from '../../services/cla.service';
import { Http } from '@angular/http';

@IonicPage({
  segment: 'cla-organization-provider-modal'
})
@Component({
  selector: 'cla-organization-provider-modal',
  templateUrl: 'cla-organization-provider-modal.html',
})
export class ClaOrganizationProviderModal {
  form: FormGroup;
  submitAttempt: boolean = false;
  currentlySubmitting: boolean = false;
  responseErrors: string[] = [];
  claProjectId: any;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    public viewCtrl: ViewController,
    private formBuilder: FormBuilder,
    public http: Http,
    public claService: ClaService,
    public events: Events
  ) {
    this.claProjectId = this.navParams.get('claProjectId');
    this.form = formBuilder.group({
      // provider: ['', Validators.required],
      orgName: ['', Validators.compose([Validators.required])/*, this.urlCheck.bind(this)*/],
    });

    events.subscribe('modal:close', () => {
      this.dismiss();
    });
  }

  getDefaults() {

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
    this.postClaGithubOrganization();
  }

  postClaGithubOrganization() {
    let organization = {
      organization_project_id: this.claProjectId,
      organization_name: this.form.value.orgName,
    };
    this.claService.postGithubOrganization(organization).subscribe((response) => {
      this.responseErrors = [];

      if (response.errors) {
        this.form.controls['orgName'].setErrors({'incorrect': true});

        for (let errorKey in response.errors) {
          this.responseErrors.push(response.errors[errorKey]);
        }

      } else {
        this.dismiss()
      }
    });
  }

  dismiss() {
    this.viewCtrl.dismiss();
  }

}
