// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { Component } from '@angular/core';
import { AlertController, IonicPage, NavParams, ViewController } from 'ionic-angular';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ClaService } from '../../services/cla.service';
import { ClaCompanyModel } from '../../models/cla-company';
import { AuthService } from '../../services/auth.service';

@IonicPage({
  segment: 'add-company-modal'
})
@Component({
  selector: 'add-company-modal',
  templateUrl: 'add-company-modal.html'
})
export class AddCompanyModal {
  form: FormGroup;
  submitAttempt: boolean = false;
  currentlySubmitting: boolean = false;

  company: ClaCompanyModel;
  companyName: string;
  userEmail: string;
  userId: string;
  userName: string;
  companies: any[];
  filteredCompanies: any[];
  companySet: boolean = false;
  joinExistingCompany: boolean = true;
  addNewCompany: boolean = false;
  enableJoinButton: boolean = false;
  existingCompanyId: string;
  mode: string = 'add';
  loading: any;
  searching: boolean;
  actionButtonsEnabled: boolean;
  activateButtons: boolean;
  join: boolean
  add: boolean

  constructor(
    public navParams: NavParams,
    public viewCtrl: ViewController,
    public formBuilder: FormBuilder,
    private claService: ClaService,
    private authService: AuthService,
    public alertCtrl: AlertController
  ) {
    this.getDefaults();
  }

  getDefaults() {
    this.searching = false;
    this.userName = localStorage.getItem('user_name');
    this.userId = localStorage.getItem('userid');
    this.company = this.navParams.get('company');
    this.mode = this.navParams.get('mode') || 'add';
    this.companies = [];
    this.filteredCompanies = [];
    this.loading = {
      submit: false,
      companies: true
    };
    this.addNewCompany = true;
    this.actionButtonsEnabled = true;


    this.add = true;
    this.join = false;
    this.activateButtons = true;

    this.form = this.formBuilder.group({
      companyName: [this.companyName, Validators.compose([Validators.required])],
      // userEmail: [this.userEmail, Validators.compose([Validators.required])],
      // userName: [this.userName, Validators.compose([Validators.required])]
    });
  }

  ngOnInit() {
    this.getAllCompanies();
  }

  ionViewDidEnter() {
    this.updateUserInfoBasedLFID();
  }

  submit() {
    this.submitAttempt = true;
    this.currentlySubmitting = true;
    this.addCompany();
  }

  addCompany() {
    this.loading.submit = true;
    let company = {
      company_name: this.companyName,
      company_manager_user_email: this.userEmail,
      company_manager_user_name: this.userName,
      company_manager_id: this.userId
    };
    this.claService.postCompany(company).subscribe(
      (response) => {
        this.currentlySubmitting = false;
        // this.getAllCompanies();
        window.location.reload(true);
        this.dismiss();
      },
      (err: any) => {
        if (err.status === 409) {
          let errJSON = err.json();
          this.companyExistAlert(errJSON.company_id);
        }
        this.currentlySubmitting = false;
      }
    );
  }

  sendCompanyNotification() {
    this.loading.submit = false;
    let alert = this.alertCtrl.create({
      title: 'Notification Sent!',
      subTitle: `A Notification has been sent to the CLA Manager for ${this.companyName}`,
      buttons: [
        {
          text: 'Ok',
          role: 'dismiss',
        }
      ]
    });
    alert.onDidDismiss(() => window.location.reload(true));
    alert.present();
  }

  joinCompany() {
    this.loading.submit = true;
    const user = {
      'lfUsername': this.userId, // required
      // Additional fields that should be updated - API only allow a few fields to be updated
      'companyID': this.existingCompanyId,
    };
    this.claService.updateUserV3(user).subscribe(
      () => {
        this.dismiss();
        this.sendCompanyNotification()
      },
      (exception) => {
        this.loading.submit = false;
        console.log('Exception while calling: updateUserV3() for user ' + this.userName +
          ' and company ID: ' + this.existingCompanyId);
        console.log(exception);
      }
    );
  }

  dismiss() {
    this.viewCtrl.dismiss(this.existingCompanyId);
  }

  companyExistAlert(company_id) {
    let alert = this.alertCtrl.create({
      title: 'Company ' + this.companyName + ' already exists',
      message: 'The company you tried to create already exists in the CLA system. Would you like to request access?',
      buttons: [
        {
          text: 'Request',
          handler: () => {
            const userId = localStorage.getItem('userid');
            const userEmail = localStorage.getItem('user_email');
            const userName = localStorage.getItem('user_name');
            this.claService
              .sendInviteRequestEmail(company_id, userId, userEmail, userName)
              .subscribe(() => this.dismiss());
          }
        },
        {
          text: 'Cancel',
          role: 'cancel',
          handler: () => {
            console.log('No clicked');
          }
        }
      ]
    });
    alert.present();
  }


  getAllCompanies() {
    if (!this.companies) {
      this.loading.companies = true;
    }
    this.claService.getAllV3Companies().subscribe((response) => {
      this.loading.companies = false;
      this.companies = response.companies;
    });
  }

  findCompany(event) {
    this.getAllCompanies();
    this.filteredCompanies = [];
    if (!this.companies) {
      this.searching = true;
    }

    if (!this.companySet) {
      this.join = false;
      this.add = true;
    } else {
      this.join = true;
      this.add = false;
    }

    this.companies.length >= 0 && this.getAllCompanies();
    // Remove all non-alpha numeric, -, _ values
    let companyName = event.value;
    if (companyName.length > 0 && this.companies) {
      this.activateButtons = false;
      this.actionButtonEnabled()
      this.searching = false;
      this.companySet = false;
      this.filteredCompanies = this.companies
        .map((company) => {
          let formattedCompany;
          if (company.companyName.toLowerCase().includes(companyName.toLowerCase())) {
            formattedCompany = company.companyName.replace(
              new RegExp(companyName, 'gi'),
              (match) => '<span class="highlightText">' + match + '</span>'
            );
          }
          if (formattedCompany === undefined && companyName.length > 4) {
            this.enableJoinButton = true
          }
          company.filteredCompany = formattedCompany;
          return company;
        })
        .filter((company) => company.filteredCompany);
    } else {
      this.activateButtons = true;
    }

    // console.log('Company Name:' + companyName);
    // console.log('Filtered Companies Length:' + this.filteredCompanies.length);

    /* Not working as desired
    if (companyName.length >= 2 && this.filteredCompanies.length === 0) {
      this.addNewCompany = true;
      this.joinExistingCompany = false;
    }
     */
    if (companyName.length >= 2) {
      this.addNewCompany = false;
      this.joinExistingCompany = true;
    }
  }

  setCompanyName(company) {
    this.companySet = true;
    this.companyName = company.companyName;
    this.existingCompanyId = company.companyID;
  }

  private updateUserInfoBasedLFID() {
    if (this.authService.isAuthenticated()) {
      this.authService
        .getIdToken()
        .then((token) => {
          return this.authService.parseIdToken(token);
        })
        .then((tokenParsed) => {
          if (tokenParsed && tokenParsed['email']) {
            this.userEmail = tokenParsed['email'];
          }
          if (tokenParsed && tokenParsed['name']) {
            this.userName = tokenParsed['name'];
          }
        })
        .catch((error) => {
          console.log(JSON.stringify(error));
          return;
        });
    }
    return;
  }

  //  Move action methods

  actionButtonEnabled() {
    this.actionButtonsEnabled = false
  }

  addButtonDisabled(): boolean {
    return false;
  }

  joinButtonDisabled(): boolean {
    return !this.enableJoinButton;
  }

  addButtonColorDisabled(): string {
    if (this.addNewCompany) {
      return 'gray';
    } else {
      return 'secondary';
    }
  }

  joinButtonColorDisabled(): string {
    if (this.joinExistingCompany) {
      return 'gray';
    } else {
      return 'secondary';
    }
  }

}
