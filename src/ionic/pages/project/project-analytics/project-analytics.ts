import { Component } from '@angular/core';
import { NavController, ModalController, NavParams, IonicPage } from 'ionic-angular';
import { CincoService } from '../../../services/cinco.service';
import { KeycloakService } from '../../../services/keycloak/keycloak.service';
import { DomSanitizer} from '@angular/platform-browser';
import { RolesService } from '../../../services/roles.service';
import { Restricted } from '../../../decorators/restricted';

@Restricted({
  roles: ['isAuthenticated', 'isPmcUser'],
})
@IonicPage({
  segment: 'project/:projectId/analytics'
})
@Component({
  selector: 'project-analytics',
  templateUrl: 'project-analytics.html',
  providers: [CincoService]
})
export class ProjectAnalyticsPage {

  projectId: string;

  hasBitergia: boolean;
  analyticsUrl: any;
  sanitizedBitergiaUrl: any;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private cincoService: CincoService,
    private keycloak: KeycloakService,
    private domSanitizer : DomSanitizer,
    public modalCtrl: ModalController,
    public rolesService: RolesService,
  ) {
    this.projectId = navParams.get('projectId');
    this.getDefaults();
  }

  ngOnInit() {
    this.getProjectConfig(this.projectId);
  }

  getDefaults() {

  }

  getProjectConfig(projectId) {
    this.cincoService.getProjectConfig(projectId).subscribe(response => {
      if (response) {
        let projectConfig = response;
        if(projectConfig.analyticsUrl) {
          this.analyticsUrl = projectConfig.analyticsUrl;
          this.sanitizedBitergiaUrl = this.domSanitizer.bypassSecurityTrustResourceUrl(this.analyticsUrl);
          this.hasBitergia = true;
        }
        else{
          this.hasBitergia = true;
        }
      }
    });
  }

  openAnaylticsConfigModal(projectId) {
    let modal = this.modalCtrl.create('AnalyticsConfigModal', {
      projectId: projectId,
    });
    modal.onDidDismiss(analyticsUrl => {
      if(analyticsUrl){
        this.analyticsUrl = analyticsUrl;
        this.hasBitergia = true;
        this.sanitizedBitergiaUrl = this.domSanitizer.bypassSecurityTrustResourceUrl(this.analyticsUrl);
      }
    });
    modal.present();
  }

}
