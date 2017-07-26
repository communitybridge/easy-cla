import { Component } from '@angular/core';
import { NavController, ModalController, NavParams, IonicPage } from 'ionic-angular';
import { CincoService } from '../../services/cinco.service';
import { SortService } from '../../services/sort.service';
import { ProjectModel } from '../../models/project-model';

@IonicPage({
  segment: 'project/:projectId'
})
@Component({
  selector: 'project',
  templateUrl: 'project.html',
  providers: [CincoService]
})
export class ProjectPage {
  selectedProject: any;
  projectId: string;

  project = new ProjectModel();

  membersCount: number;
  loading: any;
  sort: any;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private cincoService: CincoService,
    private sortService: SortService,
    public modalCtrl: ModalController,
  ) {
    this.selectedProject = navParams.get('project');
    this.projectId = navParams.get('projectId');
    this.getDefaults();
  }

  ngOnInit() {
    this.getProject(this.projectId);
  }

  getProject(projectId) {
    let getMembers = true;
    this.cincoService.getProject(projectId, getMembers).subscribe(response => {
      if(response) {
        this.project.id = response.id;
        this.project.name = response.name;
        this.project.description = response.description;
        this.project.managers = response.managers;
        this.project.status = response.status;
        this.project.category = response.category;
        this.project.sector = response.sector;
        this.project.url = response.url;
        this.project.startDate = response.startDate;
        this.project.logoRef = response.logoRef;
        this.project.agreementRef = response.agreementRef;
        this.project.mailingListType = response.mailingListType;
        this.project.emailAliasType = response.emailAliasType;
        this.project.address = response.address;
        this.project.members = response.members;
        this.membersCount = this.project.members.length;
        this.loading.project = false;
      }
    });
  }

  memberSelected(event, memberId) {
    this.navCtrl.push('MemberPage', {
      projectId: this.projectId,
      memberId: memberId,
    });
  }

  viewProjectDetails(projectId){
    this.navCtrl.push('ProjectDetailsPage', {
      projectId: projectId
    });
  }

  openProjectUserManagementModal() {
    let modal = this.modalCtrl.create('ProjectUserManagementModal', {
      projectId: this.projectId,
      projectName: this.project.name,
    });
    modal.present();
  }

  openAssetManagementModal() {
    let modal = this.modalCtrl.create('AssetManagementModal', {
      projectId: this.projectId,
    });
    modal.present();
  }

  getDefaults() {
    this.loading = {
      project: true,
    };
    this.project = {
      id: "",
      name: "Project",
      description: "Description",
      managers: "",
      members: [],
      status: "",
      category: "",
      sector: "",
      url: "",
      startDate: "",
      logoRef: "",
      agreementRef: "",
      mailingListType: "",
      emailAliasType: "",
      address: {
        address: {
          administrativeArea: "",
          country: "",
          localityName: "",
          postalCode: "",
          thoroughfare: ""
        },
        type: ""
      }
    };
    this.sort = {
      alert: {
        arrayProp: 'alert',
        sortType: 'text',
        sort: null,
      },
      company: {
        arrayProp: 'org.name',
        sortType: 'text',
        sort: null,
      },
      product: {
        arrayProp: 'product',
        sortType: 'text',
        sort: null,
      },
      status: {
        arrayProp: 'invoices[0].status',
        sortType: 'text',
        sort: null,
      },
      dues: {
        arrayProp: 'annualDues',
        sortType: 'number',
        sort: null,
      },
      renewal: {
        arrayProp: 'renewalDate',
        sortType: 'date',
        sort: null,
      },
    };
  }

  sortMembers(prop) {
    this.sortService.toggleSort(
      this.sort,
      prop,
      this.project.members,
    );
  }

}
