import { Component, Renderer, } from '@angular/core';
import { NavController, NavParams, ViewController, AlertController, ToastController, IonicPage, ModalController, } from 'ionic-angular';
import { CincoService } from '../../app/services/cinco.service'

@IonicPage({
  segment: 'project-user-management'
})
@Component({
  selector: 'project-user-management',
  templateUrl: 'project-user-management.html',
  providers: [CincoService]
})
export class ProjectUserManagementModal {
  projectId: string;
  projectName: string;
  userIds: any;
  users: any;
  selectedUsers: any;
  enteredUserId: string;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    public viewCtrl: ViewController,
    private cincoService: CincoService,
    public toastCtrl: ToastController,
    private renderer: Renderer,
    public modalCtrl: ModalController,
    public alertCtrl: AlertController,
  ) {
    this.projectId = this.navParams.get('projectId');
    this.projectName = this.navParams.get('projectName');
    this.getDefaults();
  }

  ngOnInit() {
    this.getProjectConfig();
  }

  getDefaults() {
    this.users = [];
    this.selectedUsers = [];
  }

  // ContactUpdate modal dismiss
  dismiss() {
    this.viewCtrl.dismiss();
  }

  getProjectConfig() {
    this.cincoService.getProjectConfig(this.projectId).subscribe(response => {
      if(response) {
        this.userIds = response.programManagers;
        this.updateUsers();
      }
    });
  }

  updateUsers() {
    let users = this.userIds;
    this.users = [];
    for(let i = 0; i < users.length; i++) {
      this.appendUser(users[i]);
    }
  }

  appendUser(userId) {
    this.cincoService.getUser(userId).subscribe(response => {
      if(response) {
        this.users.push(response);
      }
    });
  }

  removeUser(userId) {
    let index = this.userIds.indexOf(userId);
    if(index !== -1) {
      this.userIds.splice(index, 1);
      let updatedManagers = {
        managers: JSON.stringify(this.userIds)
      }
      this.cincoService.updateProjectManagers(this.projectId, updatedManagers).subscribe(response => {
        if(response) {
          this.getProjectConfig();
        }
      });
    }
  }

  openAssignUserModal() {
    let modal = this.modalCtrl.create('AssignUserModal', {
    });
    modal.onDidDismiss(data => {
      // A refresh of data anytime the modal is dismissed
      if(data && data.selectedUsers && data.selectedUsers.length > 0) {
        for(let i=0; i<data.selectedUsers.length; i++) {
          let userId = data.selectedUsers[i].userId;
          this.userIds.push(userId);
        }
        let updatedManagers = {
          managers: JSON.stringify(this.userIds)
        }
        this.cincoService.updateProjectManagers(this.projectId, updatedManagers).subscribe(response => {
          if(response) {
            this.getProjectConfig();
          }
        });
      }
    });
    modal.present();
  }

}
