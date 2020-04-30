// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { NgModule } from '@angular/core';
import { IonicPageModule } from 'ionic-angular';
import { CompaniesPage } from './companies-page';
import { LoadingSpinnerComponentModule } from '../../components/loading-spinner/loading-spinner.module';
import { LoadingDisplayDirectiveModule } from '../../directives/loading-display/loading-display.module';
import { LayoutModule } from '../../layout/layout.module';
import { NgxDatatableModule } from '@swimlane/ngx-datatable';
import { NgxPaginationModule } from 'ngx-pagination';
import { GetHelpComponentModule } from '../../components/get-help/get-help.module'; 
import { SharedModule } from '../../shared.module';

@NgModule({
  declarations: [CompaniesPage],
  imports: [
    LoadingSpinnerComponentModule,
    LoadingDisplayDirectiveModule,
    GetHelpComponentModule,
    IonicPageModule.forChild(CompaniesPage),
    LayoutModule,
    NgxDatatableModule,
    NgxPaginationModule,
    SharedModule
  ],
  entryComponents: [CompaniesPage]
})
export class CompaniesPageModule {}
