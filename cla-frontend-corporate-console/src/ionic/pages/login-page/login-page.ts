// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: AGPL-3.0-or-later

import { Component } from '@angular/core';
import { NavController, IonicPage } from 'ionic-angular';
import { AuthService } from "../../services/auth.service";

@IonicPage({
  name: 'LoginPage',
  segment: 'login'
})
@Component({
  selector: 'login-page',
  templateUrl: 'login-page.html'
})
export class LoginPage {

  constructor(
    public navCtrl: NavController,
    public authService: AuthService
  ) {
  }

  login() {
    this.authService.login();
  }
}
