// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { NgModule } from '@angular/core';
import { IonicPageModule } from 'ionic-angular';
import { ViewCLAManagerModal } from './view-cla-managers-modal';
import { SharedModule } from '../../shared.module';
@NgModule({
  declarations: [ViewCLAManagerModal],
  imports: [IonicPageModule.forChild(ViewCLAManagerModal), SharedModule],
  entryComponents: [ViewCLAManagerModal]
})
export class ViewCLAManagerModalModule { }
