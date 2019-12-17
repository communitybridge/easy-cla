import { Component } from '@angular/core';
import { IonicPage, NavController, AlertController } from 'ionic-angular';
import { ClaService } from '../../services/cla.service';
import { RolesService } from '../../services/roles.service';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { EmailValidator } from '../../validators/email';

@IonicPage()
@Component({
  selector: 'page-cla-manager-onboarding',
  templateUrl: 'cla-manager-onboarding.html'
})
export class ClaManagerOnboardingPage {
  projectId: any;
  companyId: any;
  selectedProjects: any;
  foundProject: boolean;
  filteredProjects: any;
  loading: any;
  companies: any;

  userId: string;
  userEmail: string;
  userName: string;

  manager: string;
  formErrors: any[];
  form: FormGroup;
  submitAttempt: boolean = false;
  currentlySubmitting: boolean = false;
  formSuccessfullySubmitted: boolean = false;
  claManagerApproved: boolean = false;

  searchTerm: string = '';
  searchProject: string = '';
  filteredCompanies: any;
  searching: boolean;
  foundCompay: boolean;
  allProjects: any;

  constructor(
    public navCtrl: NavController,
    private claService: ClaService,
    private formBuilder: FormBuilder,
    private rolesService: RolesService, // for @Restricted
    public alertCtrl: AlertController
  ) {
    this.form = formBuilder.group({
      project_name: ['', Validators.compose([Validators.required])],
      company_name: ['', Validators.compose([Validators.required])],
      full_name: ['', Validators.compose([Validators.required])],
      lfid: ['', Validators.compose([Validators.required])],
      email_address: ['', Validators.compose([Validators.required, EmailValidator.isValid])],
      reason_for_request: ['']
    });
    this.formErrors = [];
  }

  ngOnInit() {
    this.getDefaults();
    this.getAllCompanies();
    this.getProjects();
  }

  getDefaults() {
    this.loading = {
      companies: true,
      projects: true
    };
    this.searching = false;
    this.userId = localStorage.getItem('userid');
    this.userEmail = localStorage.getItem('user_email');
    this.userName = localStorage.getItem('user_name');
    this.setUserDetails();
    this.foundCompay = true;
    this.foundProject = true;
    this.filteredCompanies = [];
  }

  submit() {
    // Reset our status and error messages
    this.submitAttempt = true;
    this.currentlySubmitting = true;

    this.claService.getCompanyProjectSignatures(this.companyId, this.projectId).subscribe((response) => {
      if (response) {
        this.currentlySubmitting = false;
        if (response.signatures) {
          this.initialManagerAlreadyExists('gjg');
        } else {
          this.sendEmailToLFAdmin();
        }
      }
    });
  }

  initialManagerAlreadyExists(claManager) {
    let alert = this.alertCtrl.create({
      title: 'Your request will be sent to the Primary CLA Manager for approval',
      buttons: [
        {
          text: 'Ok',
          handler: () => {
            this.navCtrl.push('CompaniesPage');
          }
        }
      ]
    });
    alert.present();
  }

  sendEmailToLFAdmin() {
    let alert = this.alertCtrl.create({
      title: 'An email has been sent to admin@lfcla.com with all the information provided in the form',
      buttons: [
        {
          text: 'Ok',
          handler: () => {
            this.navCtrl.push('CompaniesPage');
          }
        }
      ]
    });
    alert.present();
    // TODO: send email
  }

  sendEmailToInitialCLAManager(emailAddress, lfid, projectName, reason, claManager) {}

  getAllCompanies() {
    if (!this.companies) {
      this.loading.companies = true;
    }
    this.claService.getAllCompanies().subscribe((response) => {
      if (response) {
        this.loading.companies = false;
        // Cleanup - Remove any companies that don't have a name
        this.companies = response.filter((company) => {
          return company.company_name && company.company_name.trim().length > 0;
        });
      }
    });
  }

  getProjects() {
    if (!this.allProjects) {
      this.loading.projects = true;
    }
    this.claService.getProjects().subscribe(
      (response) => {
        if (response) {
          this.allProjects = this.sortProjects(response);
        }
      },
      (error) => error
    );
  }

  sortProjects(projects) {
    if (projects == null || projects.length == 0) {
      this.loading.projects = false;
      return projects;
    }

    return projects.sort((a, b) => {
      this.loading.projects = false;
      return a.project_name.localeCompare(b.project_name);
    });
  }

  searchProjects(project) {
    let projectName = project.replace(/[^\w-]+/g, '');
    if (!this.allProjects) {
      this.loading.projects = true;
    }
    if (projectName.length > 0 && this.allProjects) {
      this.loading.projects = false;
      return (
        this.allProjects &&
        this.allProjects
          .map((project) => {
            let formattedProject;
            if (project.project_name.toLowerCase().includes(projectName.toLowerCase())) {
              formattedProject = project.project_name.replace(
                new RegExp(projectName, 'gi'),
                (match) => '<span class="highlightText">' + match + '</span>'
              );
            }
            project.filteredProject = formattedProject;
            return project;
          })
          .filter((project) => project.filteredProject)
      );
    }
  }

  setFilteredProjects() {
    this.getProjects();
    this.filteredProjects = this.searchProjects(this.searchProject);
    if (this.searchProject.length > 3 && this.filteredProjects.length === 0) {
      this.foundProject = false;
    } else {
      this.foundProject = true;
    }
  }

  findCompany(event) {
    this.getAllCompanies();
    this.filteredCompanies = [];
    let companyName = event.value.replace(/[^\w-]+/g, '');
    if (!this.companies) {
      this.searching = true;
    }
    if (companyName.length > 0 && this.companies) {
      this.searching = false;
      this.filteredCompanies =
        this.companies &&
        this.companies
          .map((company) => {
            let formattedCompany;
            if (company.company_name.toLowerCase().includes(companyName.toLowerCase())) {
              formattedCompany = company.company_name.replace(
                new RegExp(companyName, 'gi'),
                (match) => '<span class="highlightText">' + match + '</span>'
              );
            }
            company.filteredCompany = formattedCompany;
            return company;
          })
          .filter((company) => company.filteredCompany);
    }
  }

  setFilteredCompanies() {
    this.getAllCompanies();
    this.filteredCompanies = this.findCompany(this.searchTerm);
    if (this.searchTerm.length > 3 && this.filteredCompanies.length === 0) {
      this.foundCompay = false;
    } else {
      this.foundCompay = true;
    }
  }

  setCompanyName(company) {
    this.searchTerm = company.company_name;
    this.form.controls['company_name'].setValue(this.searchTerm);
    this.filteredCompanies = [];
    this.companyId = company.company_id;
  }

  setProjectName(project) {
    this.searchProject = project.project_name;
    this.form.controls['project_name'].setValue(this.searchProject);
  }

  setUserDetails() {
    this.form.controls['lfid'].setValue(this.userId);
    this.form.controls['email_address'].setValue(this.userEmail);
    this.form.controls['full_name'].setValue(this.userName);
  }

  projectSelectChanged(value) {
    this.form.controls['project_name'].setValue(value.project_name);
    this.projectId = value.project_id;
  }
}
