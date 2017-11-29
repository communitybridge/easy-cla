import { Component } from '@angular/core';
import { NavController, ModalController, NavParams, IonicPage } from 'ionic-angular';
import { CincoService } from '../../../services/cinco.service';
import { KeycloakService } from '../../../services/keycloak/keycloak.service';
import { SortService } from '../../../services/sort.service';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { CheckboxValidator } from  '../../../validators/checkbox';

@IonicPage({
  segment: 'project/:projectId/details'
})
@Component({
  selector: 'project-details',
  templateUrl: 'project-details.html',
})
export class ProjectDetailsPage {
  projectId: string;

  details: any;

  loading: any;
  sort: any;

  contracts: any;

  iclaUploadInfo: any;
  cclaUploadInfo: any;

  form: FormGroup;
  submitAttempt: boolean = false;
  currentlySubmitting: boolean = false;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private cincoService: CincoService,
    private sortService: SortService,
    public modalCtrl: ModalController,
    private formBuilder: FormBuilder,
    private keycloak: KeycloakService
  ) {
    this.projectId = navParams.get('projectId');
    this.form = formBuilder.group({
      confirm:[false, Validators.compose([CheckboxValidator.isChecked])],
    });
    this.getDefaults();
  }

  ionViewCanEnter() {
    if(!this.keycloak.authenticated())
    {
      this.navCtrl.setRoot('LoginPage');
      this.navCtrl.popToRoot();
    }
    return this.keycloak.authenticated();
  }

  ionViewWillEnter() {
    if(!this.keycloak.authenticated())
    {
      this.navCtrl.push('LoginPage');
    }
  }

  ngOnInit() {
    this.getProjectDetails();
  }

  getProjectDetails() {
    // this.cincoService.getProjectDetails(this.projectId).subscribe(response => {
    //   if (response) {
    //     this.details = response;
    //   }
    //   this.loading.details = false;
    // });
    setTimeout((function(){
      this.details = {
        product: "Gold Membership",
        startTier: "",
        endTier: "",
        level: "End User",
        startDate: "1/1/2017",
        endDate: "12/31/2017",
        poNumber: "0123456789",
        description: "Gold Membership, Sample Project",
        netsuiteMemo:"400 character memo field goes here. This is a memo field where someone might choose to write some sort of memo or other information that relates to specifically how the member should receive their invoice, for instance one time Jen had to write out a detailed fee schedule.",
        amount: "$500,000",
      };
      this.loading.details = false;
    }).bind(this),2000);

  }

  getDefaults() {
    this.loading = {
      details: true,
    };
    this.details = {
      product: "",
      startTier: "",
      endTier: "",
      level: "",
      startDate: "",
      endDate: "",
      poNumber: "",
      description: "",
      netsuiteMemo:"",
      ammount:"",
    };

  }

  submit() {

  }
}
