import { NgModule } from '@angular/core';
import { IonicPageModule } from 'ionic-angular';
import { ClaContractsContributorsPage } from './cla-contracts-contributors';
import { LoadingSpinnerComponentModule } from '../../components/loading-spinner/loading-spinner.module';
import { LoadingDisplayDirectiveModule } from '../../directives/loading-display/loading-display.module';
import { SortingDisplayComponentModule } from '../../components/sorting-display/sorting-display.module';

@NgModule({
  declarations: [
    ClaContractsContributorsPage,
  ],
  imports: [
    LoadingSpinnerComponentModule,
    LoadingDisplayDirectiveModule,
    SortingDisplayComponentModule,
    IonicPageModule.forChild(ClaContractsContributorsPage)
  ],
  entryComponents: [
    ClaContractsContributorsPage,
  ]
})
export class ClaContractsContributorsPageModule {}
