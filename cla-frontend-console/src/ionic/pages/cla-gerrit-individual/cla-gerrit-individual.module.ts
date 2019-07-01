// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { NgModule } from '@angular/core';
import { IonicPageModule } from 'ionic-angular';
import { ClaGerritIndividualPage } from './cla-gerrit-individual';

@NgModule({
  declarations: [
    ClaGerritIndividualPage,
  ],
  imports: [
    IonicPageModule.forChild(ClaGerritIndividualPage),
  ],
  entryComponents: [
    ClaGerritIndividualPage
  ]
})
export class ClaGerritIndividualPageModule {}
