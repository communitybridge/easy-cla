// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import {Component, ViewChild} from "@angular/core";
import {IonicPage, Nav, NavController, NavParams} from "ionic-angular";
import {ClaService} from "../../../services/cla.service";
import {Restricted} from "../../../decorators/restricted";
import {DomSanitizer} from '@angular/platform-browser';

@Restricted({
  roles: ["isAuthenticated", "isPmcUser"]
})
@IonicPage({
  segment: "project/:projectId/cla/template/:projectTemplateId"
})
@Component({
  selector: "project-cla-template",
  templateUrl: "project-cla-template.html"
})
export class ProjectClaTemplatePage {
  sfdcProjectId: string;
  projectId: string;
  templates: any[] = [];
  selectedTemplate: any;
  templateValues = {};
  pdfPath = {
    corporatePDFURL: '',
    individualPDFURL: ''
  };
  currentPDF = 'corporatePDFURL';
  step = 'selection';
  buttonGenerateEnabled = true;
  message = null;
  loading = {
    documents: false
  };

  @ViewChild(Nav) nav: Nav;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    public claService: ClaService,
    public sanitizer: DomSanitizer
  ) {
    this.sfdcProjectId = navParams.get("sfdcProjectId");
    this.projectId = navParams.get("projectId");
    this.getDefaults();
  }

  getDefaults() {
    this.getTemplates();
  }

  getTemplates() {
    this.claService.getTemplates().subscribe(templates => this.templates = templates);
  }

  ngOnInit() {
    this.setLoadingSpinner(false);
  }

  getPdfPath() {
    return this.sanitizer.bypassSecurityTrustResourceUrl(this.pdfPath[this.currentPDF]);
  }

  showPDF(type) {
    this.currentPDF = type;
  }

  selectTemplate(template) {
    this.selectedTemplate = template;
    this.step = 'values';
  }

  reviewSelectedTemplate() {
    this.setLoadingSpinner(true);
    this.buttonGenerateEnabled = false;
    this.message = 'Generating PDFs...';

    const metaFields = this.selectedTemplate.metaFields;
    metaFields.forEach(metaField => {
      if (this.templateValues.hasOwnProperty(metaField.templateVariable)) {
        metaField.value = this.templateValues[metaField.templateVariable]
      }

    });
    let data = {
      templateID: this.selectedTemplate.ID,
      metaFields: metaFields
    };

    this.claService.postClaGroupTemplate(this.projectId, data)
      .subscribe(response => {
        this.setLoadingSpinner(false);
        this.buttonGenerateEnabled = true;
        this.message = null;
        this.pdfPath = response;
        this.goToStep('review');
      }, (error) => {
        this.setLoadingSpinner(false);
        this.buttonGenerateEnabled = true;
        this.message = 'Error creating PDFs: ' + error;
      });
  }

  goToStep(step) {
    this.step = step;
  }

  backToProject() {
    this.navCtrl.pop();
  }

  setLoadingSpinner(value: boolean) {
    this.loading = {
      documents: value
    };
  }
}
