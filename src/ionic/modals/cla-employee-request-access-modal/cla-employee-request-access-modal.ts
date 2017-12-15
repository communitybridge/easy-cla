import { Component, ChangeDetectorRef } from '@angular/core';
import { NavController, NavParams, ModalController, ViewController, AlertController, IonicPage } from 'ionic-angular';
import { FormBuilder, FormGroup } from '@angular/forms';
import { Validators } from '@angular/forms';
import { EmailValidator } from  '../../validators/email';
import { ClaService } from 'cla-service';

@IonicPage({
  segment: 'cla/project/:projectId/repository/:repositoryId/user/:userId/employee/company/contact'
})
@Component({
  selector: 'cla-employee-request-access-modal',
  templateUrl: 'cla-employee-request-access-modal.html',
})
export class ClaEmployeeRequestAccessModal {
  projectId: string;
  repositoryId: string;
  userId: string;
  companyId: string;

  userEmails: Array<string>;

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
    this.projectId = navParams.get('projectId');
    this.repositoryId = navParams.get('repositoryId');
    this.userId = navParams.get('userId');
    this.companyId = navParams.get('companyId');
    this.form = formBuilder.group({
      email:['', Validators.compose([Validators.required, EmailValidator.isValid])],
      message:[''], // Validators.compose([Validators.required])
    });
  }

  getDefaults() {
    this.userEmails = [];
  }

  ngOnInit() {
    this.getUser();
  }

  getUser() {
    this.claService.getUser(this.userId).subscribe(user => {
      this.userEmails = user.user_emails;
    });
  }
  // ContactUpdateModal modal dismiss
  dismiss() {
    this.viewCtrl.dismiss();
  }

  submit() {
    this.submitAttempt = true;
    this.currentlySubmitting = true;
    if (!this.form.valid) {
      this.currentlySubmitting = false;
      // prevent submit
      return;
    }
    let message = {
      user_email: this.form.value.email,
      message: this.form.value.message,
    };
    this.claService.postUserMessageToCompanyManager(this.userId, this.companyId, message).subscribe(response => {
      this.openClaMessageSentPage();
    });
  }

  openClaMessageSentPage() {
    this.navCtrl.push('ClaMessageSentPage', {
      projectId: this.projectId,
      repositoryId: this.repositoryId,
      userId: this.userId,
      companyId: this.companyId,
    });
  }

}
