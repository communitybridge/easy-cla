// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { NgModule } from '@angular/core';
import { IonicPageModule } from 'ionic-angular';
import { ProjectDetailsPage } from './project-details';
import { LoadingSpinnerComponentModule } from '../../components/loading-spinner/loading-spinner.module';
import { LoadingDisplayDirectiveModule } from '../../directives/loading-display/loading-display.module';

@NgModule({
  declarations: [ProjectDetailsPage],
  imports: [LoadingSpinnerComponentModule, LoadingDisplayDirectiveModule, IonicPageModule.forChild(ProjectDetailsPage)]
})
export class ProjectDetailsPageModule {}
