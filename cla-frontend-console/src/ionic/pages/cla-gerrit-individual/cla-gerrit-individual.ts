import { Component } from '@angular/core';
import { NavController, NavParams, IonicPage, ModalController, } from 'ionic-angular';
import { ClaService } from '../../services/cla.service';
import { RolesService } from "../../services/roles.service";
import { AuthService } from "../../services/auth.service";
@IonicPage({
  segment: 'cla/gerrit/project/:projectId/individual'
})
@Component({
  selector: 'cla-gerrit-individual',
  templateUrl: 'cla-gerrit-individual.html'
})
export class ClaGerritIndividualPage {
  projectId: string;

  user: any;
  project: any;
  signatureIntent: any;
  activeSignatures: boolean = true; // we assume true until otherwise
  signature: any;

  userRoles: any;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private modalCtrl: ModalController,
    private claService: ClaService,
    private rolesService: RolesService,
    private authService: AuthService,
  ) {
    this.getDefaults();
    this.projectId = navParams.get('projectId');
  }

  getDefaults() {
    this.userRoles = this.rolesService.userRoleDefaults;

    this.project = {
      project_name: "",
    };
    this.signature = {
      sign_url: "",
    };
  }

  ngOnInit() {
    this.getProject(this.projectId);
  }

  ionViewCanEnter(){
    if(!this.authService.isAuthenticated){
      setTimeout(()=>this.navCtrl.setRoot('LoginPage'))
    }
    return this.authService.isAuthenticated
  }


  ngAfterViewInit() {
  }

  getProject(projectId) {
    this.claService.getProject(projectId).subscribe(response => {
      this.project = response;
      if (!this.project.logoRef) {
        this.project.logoRef = "https://dummyimage.com/200x100/bbb/fff.png&text=+";
      }
    });
  }
  
}
