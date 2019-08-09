// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

import { Component } from "@angular/core";
import { NavController, IonicPage } from "ionic-angular";
import { ClaService } from "../../services/cla.service";
import { FilterService } from "../../services/filter.service";
import { RolesService } from "../../services/roles.service";
import { Restricted } from "../../decorators/restricted";

@Restricted({
  roles: ["isAuthenticated", "isPmcUser"]
})
@IonicPage({
  name: "AllProjectsPage",
  segment: "projects"
})
@Component({
  selector: "all-projects",
  templateUrl: "all-projects.html"
})
export class AllProjectsPage {
  loading: any;
  projectSectors: any;
  allProjects: any;
  allFilteredProjects: any;
  userRoles: any;
  errorMessage = null;

  constructor(
    public navCtrl: NavController,
    private claService: ClaService,
    private rolesService: RolesService,
    private filterService: FilterService
  ) {
    this.getDefaults();
  }

  async ngOnInit() {
    this.getAllProjectFromSFDC();
  }

  getAllProjectFromSFDC() {
    this.claService.getAllProjectsFromSFDC().subscribe(response => {
      this.allProjects = response;
      this.allFilteredProjects = this.filterService.resetFilter(
        this.allProjects
      );
      this.loading.projects = false;
    }, (error) => this.handleErrors(error));
  }

  handleErrors (error) {
    this.setLoadingSpinner(false);

    switch (error.status) {
      case 401:
        this.errorMessage = `You don't have permissions to see any projects.`;
        break;

      default:
        this.errorMessage = `An unknown error has occurred when retrieving the projects`;
    }
  }

  viewProjectCLA(projectId) {
    this.navCtrl.setRoot("ProjectClaPage", {
      projectId: projectId
    });
  }

  getDefaults() {
    this.userRoles = this.rolesService.userRoleDefaults;

    this.setLoadingSpinner(true);
  }

  setLoadingSpinner (value) {
    this.loading = {
      projects: value
    };
  }

  filterAllProjects(projectProperty, keyword) {
    if (keyword == "NO_FILTER") {
      this.allFilteredProjects = this.filterService.resetFilter(
        this.allProjects
      );
    } else {
      this.allFilteredProjects = this.filterService.filterAllProjects(
        this.allProjects,
        projectProperty,
        keyword
      );
    }
  }

  openAccessPage() {
    window.open("https://docs.linuxfoundation.org/pages/viewpage.action?pageId=7411265", "_blank");
  }
}
