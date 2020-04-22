// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { NgModule } from '@angular/core';
import { IonicPageModule } from 'ionic-angular';
import { AuthorityYesnoPage } from './authority-yesno-page';
import { LayoutModule } from '../../layout/layout.module';
import { GetHelpComponentModule } from '../../components/get-help/get-help.module';

@NgModule({
  declarations: [AuthorityYesnoPage],
  imports: [IonicPageModule.forChild(AuthorityYesnoPage), LayoutModule, GetHelpComponentModule],
  entryComponents: [AuthorityYesnoPage]
})
export class AuthorityYesnoPageModule { }
