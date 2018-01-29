import { Component } from '@angular/core';
import { NavController, ModalController, NavParams, IonicPage } from 'ionic-angular';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { CincoService } from '../../../services/cinco.service';
import { KeycloakService } from '../../../services/keycloak/keycloak.service';
import { DomSanitizer} from '@angular/platform-browser';
import { RolesService } from '../../../services/roles.service';
import { Restricted } from '../../../decorators/restricted';

@Restricted({
  roles: ['isAuthenticated', 'isPmcUser'],
})
@IonicPage({
  segment: 'project/:projectId/groups/create'
})
@Component({
  selector: 'project-groups-create',
  templateUrl: 'project-groups-create.html',
  providers: [CincoService]
})

export class ProjectGroupsCreatePage {

  projectId: string;
  keysGetter;
  projectPrivacy;

  groupName: string;
  groupDescription: string;
  groupPrivacy = [];

  form: FormGroup;
  submitAttempt: boolean = false;
  currentlySubmitting: boolean = false;

  group: any;
  projectGroups: any;
  approve_members: boolean;
  restrict_posts: boolean;
  approve_posts: boolean;
  allow_unsubscribed: boolean;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private cincoService: CincoService,
    private keycloak: KeycloakService,
    private domSanitizer : DomSanitizer,
    public modalCtrl: ModalController,
    public rolesService: RolesService,
    private formBuilder: FormBuilder,
  ) {
    this.projectId = navParams.get('projectId');

    this.form = formBuilder.group({
      groupName:[this.groupName, Validators.compose([Validators.minLength(3), Validators.required])],
      groupDescription:[this.groupDescription, Validators.compose([Validators.minLength(9), Validators.required])],
      groupPrivacy:[this.groupPrivacy, Validators.compose([Validators.required])],
      approve_members: [this.approve_members],
      restrict_posts: [this.restrict_posts],
      approve_posts: [this.approve_posts],
      allow_unsubscribed: [this.allow_unsubscribed]
    });

  }

  ngOnInit() {
    this.getProjectConfig(this.projectId);
    this.getDefaults();
  }

  getDefaults() {
    this.keysGetter = Object.keys;
    this.getProjectGroups();
    this.getGroupPrivacy();
  }

  getProjectConfig(projectId) {
    this.cincoService.getProjectConfig(projectId).subscribe(response => {
      if (response) {
        console.log(response);
        if (!response.mailingGroup) {
          console.log("no mailingGroup");
          console.log("creating a new mailingGroup");
          this.cincoService.createMainProjectGroup(this.projectId).subscribe(response => {
            console.log("new mailingGroup");
            console.log(response);
          });
        }
      }
    });
  }

  getProjectGroups() {
    this.cincoService.getAllProjectGroups(this.projectId).subscribe(response => {
      this.projectGroups = response;
      console.log(response);
    });
  }

  getGroupPrivacy() {
    this.groupPrivacy = [];
    // TODO Implement CINCO side
    // this.cincoService.getGroupPrivacy(this.projectId).subscribe(response => {
    //   this.groupPrivacy = response;
    // });
    this.groupPrivacy = [
      {
        value: "sub_group_privacy_none",
        description: "Group listed and archive publicly viewable"
      },
      {
        value: "sub_group_privacy_archives",
        description: "Group listed and archive privately viewable by members"
      },
      {
        value: "sub_group_privacy_unlisted",
        description: "Group hidden and archive privately viewable by members"
      }
    ];
  }

  submitGroup() {
    this.submitAttempt = true;
    this.currentlySubmitting = true;
    if (!this.form.valid) {
      this.currentlySubmitting = false;
      // prevent submit
      return;
    }
    // CINCO / Groups.io Flags
    // allow_unsubscribed - Allow unsubscribed members to post
    // approve_members    - Members required approval to join
    // approve_posts      - Posts require approval
    // restrict_posts     - Posts are restricted to moderators only
    this.group = {
      projectId: this.projectId,
      name: this.form.value.groupName,
      description: this.form.value.groupDescription,
      privacy: this.groupPrivacy[this.form.value.groupPrivacy].value,
      allow_unsubscribed: this.form.value.allow_unsubscribed,
      approve_members: this.form.value.approve_members,
      approve_posts: this.form.value.approve_posts,
      restrict_posts: this.form.value.restrict_posts
    };
    console.log(this.group);
    this.cincoService.createProjectGroup(this.projectId, this.group).subscribe(response => {
      this.currentlySubmitting = false;
      this.navCtrl.setRoot('ProjectGroupsPage', {
        projectId: this.projectId
      });
    });
  }

  getGroupDetails(groupName) {
    this.navCtrl.setRoot('ProjectGroupDetailsPage', {
      projectId: this.projectId,
      groupName: groupName
    });
  }

  getGroupsList() {
    this.navCtrl.setRoot('ProjectGroupsPage', {
      projectId: this.projectId
    });
  }

}
