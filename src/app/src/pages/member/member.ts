import { Component } from '@angular/core';
import { NavController, ModalController, NavParams, IonicPage } from 'ionic-angular';
import { CincoService } from '../../app/services/cinco.service'
import { ContactUpdate } from '../contact-update/contact-update';


@IonicPage({
  segment: 'project-page/:projectId/member-page/:memberId'
})
@Component({
  selector: 'page-member',
  templateUrl: 'member.html'
})
export class MemberPage {
  projectId: any;
  memberId: any;
  member: any;

  constructor(
    public navCtrl: NavController,
    public modalCtrl: ModalController,
    public navParams: NavParams,
    private cincoService: CincoService
  ) {
    // If we navigated to this page, we will have an item available as a nav param
    this.projectId = navParams.get('projectId');
    this.memberId = navParams.get('memberId');

    this.getDefaults();
  }

  ngOnInit() {
    // use selectedProject and selectedMember to get data from CINCO and populate this.contacts
    // this.getMember(this.projectId, this.memberId);
  }

  getMember(projectId, memberId) {
    this.cincoService.getMember(projectId, memberId).subscribe(response => {
      if(response) {
        this.member = response;
      }
    });
  }

  addMemberContact() {
    let modal = this.modalCtrl.create(ContactUpdate, {
      projectId: this.projectId,
      member: this.member,
    });
    modal.present();
  }

  contactSelected(event, contact) {
    let modal = this.modalCtrl.create(ContactUpdate, {
      projectId: this.projectId,
      member: this.member,
      contact: contact,
    });
    modal.present();
  }

  getDefaults() {
    this.member = {
      id: this.memberId,
      contacts: [
        {
          bio: "",
          email: "",
          familyName: "",
          givenName: "",
          headshotRef: "",
          id: "",
          phone: "",
          type: "",
        }
      ],
      invoices: [],
      org: {
        addresses: [],
        id: "",
        logoRef: "",
        name: "",
        phone: "",
        url: "",
      },
      projectId: "",
      renewalDate: "",
      startDate: "",
      tier: {
        qualifier: "",
        type: "",
      }
    };

  }

}
