// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: AGPL-3.0-or-later

import { NgModule } from '@angular/core';

import { IonicPageModule } from 'ionic-angular';

import { AllMembersPage } from './all-members';

@NgModule({
  declarations: [
    AllMembersPage
  ],
  imports: [
    IonicPageModule.forChild(AllMembersPage)
  ],
})
export class AllMembersPageModule {}
