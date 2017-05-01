import { Component } from '@angular/core';
import { NavController, NavParams, IonicPage } from 'ionic-angular';
import { CincoService } from '../../app/services/cinco.service'

@IonicPage({
  segment: 'project-page/:projectId'
})
@Component({
  selector: 'page-project',
  templateUrl: 'project.html'
})
export class ProjectPage {
  selectedProject: any;
  projectId: string;
  project: {
    icon: string,
    name: string,
    description: string,
    datas: Array<{
      label: string,
      value: string,
    }>,
  };

  members: any;
  // members: Array<{
  //   id: string,
  //   alert?: string,
  //   name: string,
  //   level: string,
  //   status: string,
  //   annual_dues: string,
  //   renewal_date: string,
  // }>;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private cincoService: CincoService
  ) {
    // If we navigated to this page, we will have an item available as a nav param
    this.selectedProject = navParams.get('project');
    this.projectId = navParams.get('projectId');
    this.getDefaults();
  }

  ngOnInit() {
    this.getProject(this.projectId);
    this.getProjectMembers(this.projectId);
  };

  getProject(projectId) {
    this.cincoService.getProject(projectId).subscribe(response => {
      if(response) {
        this.project.name = response.name;
        this.project.icon = response.name;
        this.project.datas.push({
            label: "Status",
            value: response.status
        });
        this.project.datas.push({
            label: "Type",
            value: response.type
        });
        this.project.datas.push({
            label: "ID",
            value: response.id
        });
      }
    });
  }

  getProjectMembers(projectId) {
    console.log("getProjectMembers:");
    this.cincoService.getProjectMembers(projectId).subscribe(response => {
      if(response) {
        console.log(response);
        this.members = response;
      }
    });
  }

  memberSelected(event, memberId) {
    this.navCtrl.push('MemberPage', {
      projectId: this.projectId,
      memberId: memberId,
    });
  }

  getDefaults() {

    this.project = {
      icon: "https://dummyimage.com/600x250/ffffff/000.png&text=project+logo",
      name: "Project",
      description: "This project is a small, scalable, real-time operating system for use on resource-constraned systems supporting multiple architectures...",
      datas: [
        {
          label: "Budget",
          value: "$3,000,000",
        },
        {
          label: "Categories",
          value: "Embedded & IoT",
        },
        {
          label: "Launched",
          value: "5/1/2016",
        },
        {
          label: "Current",
          value: "$2,000,000 ($1,000,000)",
        },
        {
          label: "Members",
          value: "41",
        },
      ],
    };

    // this.members = [
    //   {
    //     id: 'm00000000001',
    //     alert: '',
    //     name: 'Abbie',
    //     level: 'Gold',
    //     status: 'Invoice Paid',
    //     annual_dues: '$30,000',
    //     renewal_date: '3/1/2017',
    //   },
    //   {
    //     id: 'm00000000002',
    //     alert: 'alert',
    //     name: 'Acrombie',
    //     level: 'Gold',
    //     status: 'Invoice Sent (Late)',
    //     annual_dues: '$30,000',
    //     renewal_date: '3/2/2017',
    //   },
    //   {
    //     id: 'm00000000003',
    //     alert: 'notice',
    //     name: 'Adobe',
    //     level: 'Gold',
    //     status: 'Contract: Pending',
    //     annual_dues: '$30,000',
    //     renewal_date: '4/1/2017',
    //   },
    //   {
    //     id: 'm00000000004',
    //     alert: '',
    //     name: 'ADP',
    //     level: 'Gold',
    //     status: 'Invoice Sent',
    //     annual_dues: '$30,000',
    //     renewal_date: '4/1/2017',
    //   },
    //   {
    //     id: 'm00000000005',
    //     alert: '',
    //     name: 'BlackRock',
    //     level: 'Bronze',
    //     status: 'Renewal < 60',
    //     annual_dues: '$30,000',
    //     renewal_date: '6/1/2017',
    //   },
    //   {
    //     id: 'm00000000006',
    //     alert: '',
    //     name: 'Fox',
    //     level: 'Bronze',
    //     status: 'Renewal < 60',
    //     annual_dues: '$30,000',
    //     renewal_date: '10/1/2017',
    //   },
    //   {
    //     id: 'm00000000007',
    //     alert: '',
    //     name: 'Google',
    //     level: 'Gold',
    //     status: 'Renewal < 60',
    //     annual_dues: '$30,000',
    //     renewal_date: '10/1/2017',
    //   },
    //   {
    //     id: 'm00000000008',
    //     alert: '',
    //     name: 'Joyent',
    //     level: 'Gold',
    //     status: 'Renewal < 60',
    //     annual_dues: '$30,000',
    //     renewal_date: '10/1/2017',
    //   },
    //   {
    //     id: 'm00000000009',
    //     alert: '',
    //     name: 'KrVolk',
    //     level: 'Gold',
    //     status: 'Renewal < 60',
    //     annual_dues: '$30,000',
    //     renewal_date: '10/1/2017',
    //   },
    //   {
    //     id: 'm00000000010',
    //     alert: '',
    //     name: 'Netflix',
    //     level: 'Gold',
    //     status: 'Renewal < 60',
    //     annual_dues: '$30,000',
    //     renewal_date: '10/1/2017',
    //   },
    //   {
    //     id: 'm00000000011',
    //     alert: '',
    //     name: 'Company Name',
    //     level: 'Silver',
    //     status: 'Renewal < 60',
    //     annual_dues: '$30,000',
    //     renewal_date: '10/1/2017',
    //   },
    // ];
  }
}
