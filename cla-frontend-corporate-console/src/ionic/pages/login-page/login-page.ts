import { Component } from '@angular/core';
import { NavController, IonicPage } from 'ionic-angular';
import { KeycloakService } from '../../services/keycloak/keycloak.service';

@IonicPage({
  name: 'LoginPage',
  segment: 'login'
})
@Component({
  selector: 'login-page',
  templateUrl: 'login-page.html'
})
export class LoginPage {

  constructor(public navCtrl: NavController, private keycloak: KeycloakService) {
  }

  ionViewWillEnter() {
    if(this.keycloak.authenticated())
    {
      this.navCtrl.setRoot('CompaniesPage');
      this.navCtrl.popToRoot();
    }
  }

  ionViewCanLeave() {
    return (this.keycloak.authenticated());
  }

  login() {
    if (this.keycloak.authenticated()) {
      this.navCtrl.setRoot('CompaniesPage');
    }
    else{
      this.keycloak.login();
    }
  }

}
