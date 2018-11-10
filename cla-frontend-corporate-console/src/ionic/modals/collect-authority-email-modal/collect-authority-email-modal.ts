import { Component, ChangeDetectorRef } from '@angular/core';
import { NavController, NavParams, ModalController, ViewController, AlertController, IonicPage } from 'ionic-angular';
import { FormBuilder, FormGroup } from '@angular/forms';
import { Validators } from '@angular/forms';
import { EmailValidator } from  '../../validators/email';
import { ClaService } from '../../services/cla.service';
import { EnvConfig } from '../../services/cla.env.utils';

@IonicPage({
  segment: 'cla/project/:projectId/collect-authority-email'
})
@Component({
  selector: 'collect-authority-email-modal',
  templateUrl: 'collect-authority-email-modal.html',
})
export class CollectAuthorityEmailModal {
  projectId: string;
  companyId: string;

  projectName: string;
  companyName: string; 

  form: FormGroup;
  submitAttempt: boolean = false;
  currentlySubmitting: boolean = false;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    public modalCtrl: ModalController,
    public viewCtrl: ViewController,
    public alertCtrl: AlertController,
    private changeDetectorRef: ChangeDetectorRef,
    private formBuilder: FormBuilder,
    private claService: ClaService,
  ) {
    this.getDefaults();
    this.projectName = navParams.get('projectName');
    this.companyName = navParams.get('companyName');
    this.projectId = navParams.get('projectId');
    this.companyId = navParams.get('companyId');
    this.form = formBuilder.group({
      authorityemail:['', Validators.compose([Validators.required, EmailValidator.isValid])],
      authorityname:[''],
    });
  }

  getDefaults() {
  }

  ngOnInit() {
  }

  dismiss() {
    this.viewCtrl.dismiss();
  }

  emailSent() {
    let alert = this.alertCtrl.create({
      title: 'E-Mail Successfully Sent!',
      subTitle: 'An E-Mail has been successfully sent. Please wait for your corporate authority to review and sign the CCLA Agreement. ',
      buttons: ['Dismiss']
    });
    alert.present();
  }

  submit() {
    this.submitAttempt = true;
    this.currentlySubmitting = true;
    if (!this.form.valid) {
      this.currentlySubmitting = false;
      return;
    }
    console.log(this.projectName);
    let emailRequest = {
      project_id: this.projectId,
      company_id: this.companyId,
      send_as_email: true, 
      authority_name: this.form.value.authorityname, 
      authority_email: this.form.value.authorityemail, 
    };
    
    this.claService
      .postCorporateSignatureRequest(emailRequest)
      .subscribe(response => {
        if (response.errors) {
          //TODO: CREATE error message
          console.log(response.errors); 
        }
        this.emailSent();
        this.dismiss();
      });
  }

}
