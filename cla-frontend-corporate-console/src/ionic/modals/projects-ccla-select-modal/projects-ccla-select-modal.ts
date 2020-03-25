// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { Component } from '@angular/core';
import { IonicPage, NavController, NavParams, ViewController } from 'ionic-angular';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ClaService } from '../../services/cla.service';
import { ClaCompanyModel } from '../../models/cla-company';
import { PlatformLocation } from '@angular/common';

@IonicPage({
  segment: 'projects-ccla-select-modal'
})
@Component({
  selector: 'projects-ccla-select-modal',
  templateUrl: 'projects-ccla-select-modal.html'
})
export class ProjectsCclaSelectModal {
  submitDisabled: boolean;
  projectId: any;
  form: FormGroup;
  projects: any;
  projectsFiltered: any;
  loading: any;
  company: ClaCompanyModel;

  constructor(
    public navParams: NavParams,
    public navCtrl: NavController,
    public viewCtrl: ViewController,
    public formBuilder: FormBuilder,
    private claService: ClaService,
    private location: PlatformLocation
  ) {
    this.form = formBuilder.group({
      // provider: ['', Validators.required],
      search: ['', Validators.compose([Validators.required]) /*, this.urlCheck.bind(this)*/]
    });
    this.location.onPopState(() => {
      this.viewCtrl.dismiss(false);
    });
  }

  getDefaults() {
    this.loading = true;
    this.submitDisabled = true;
    this.company = this.navParams.get('company');
  }

  ngOnInit() {
    this.getDefaults();
    this.getProjectsCcla();
  }

  getProjectsCcla() {
    const companyId = this.navParams.get('companyId');
    this.claService.getCompanyUnsignedProjects(companyId).subscribe((response) => {
      this.loading = false;
      // Sort on the project name field after filtering empty project names
      this.projects = response
        .filter((a) => a != null && a.project_name != null && a.project_name.trim().length > 0)
        .sort((a, b) => {
          // force project_name to be a string to avoid any exceptions - sort use users locale, trim any whitespace out
          return ('' + a.project_name.trim()).localeCompare(b.project_name.trim());
        });

      // Reset our filtered search
      this.form.value.search = '';
      this.projectsFiltered = this.projects;
    });
  }

  /**
   * onSearch simply filters the projects view
   */
  onSearch() {
    const searchTerm = this.form.value.search;
    if (searchTerm === '') {
      this.projectsFiltered = this.projects;
    } else {
      this.projectsFiltered = this.projects.filter((a) => {
        return a.project_name.toLowerCase().includes(searchTerm.toLowerCase());
      });
    }
  }

  selectProject(project) {
    this.submitDisabled = false
    this.form.controls['search'].setValue(project.project_name);
    this.projectId = project.project_id

  }

  submit() {
    this.navCtrl.push('AuthorityYesnoPage', {
      projectId: this.projectId,
      company: this.company
    });

    this.dismiss();
  }

  /**
   * Returns the project name formatted based on the search filter - should highlight the matching text
   * @param project
   */
  formatProject(project) {
    const searchTerm = this.form.value.search;

    // If no search term, just return the plain value
    if (searchTerm == null || searchTerm === '') {
      return project;
    }

    // Grab the index of the matching characters
    const index = project.toLowerCase().indexOf(searchTerm.toLowerCase());

    // If we have a match...
    if (index >= 0) {
      //console.log(component);
      // this.el.nativeElement.innerHTML
      // console.log('Styling matching project...index = ' + index);
      return (
        project.substring(0, index) +
        '<span class="highlight">' +
        project.substring(index, index + searchTerm.length) +
        '</span>' +
        project.substring(index + searchTerm.length)
      );
    } else {
      // No match, just return the plain text
      return project;
    }
  }

  dismiss() {
    this.viewCtrl.dismiss();
  }
}
