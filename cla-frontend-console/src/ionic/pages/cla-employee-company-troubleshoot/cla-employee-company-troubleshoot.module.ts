import { NgModule } from '@angular/core';
import { IonicPageModule } from 'ionic-angular';
import { ClaEmployeeCompanyTroubleshootPage } from './cla-employee-company-troubleshoot';
import { LayoutModule } from "../../layout/layout.module";

@NgModule({
  declarations: [
    ClaEmployeeCompanyTroubleshootPage,
  ],
  imports: [
    IonicPageModule.forChild(ClaEmployeeCompanyTroubleshootPage),
    LayoutModule
  ],
  entryComponents: [
    ClaEmployeeCompanyTroubleshootPage
  ]
})
export class ClaEmployeeCompanyTroubleshootPageModule {}
