import { Component } from '@angular/core';
import { NavController, ModalController, NavParams, IonicPage } from 'ionic-angular';
import { CincoService } from '../../../services/cinco.service';
import { KeycloakService } from '../../../services/keycloak/keycloak.service';
import { AnalyticsService } from '../../../services/analytics.service';
import { DomSanitizer} from '@angular/platform-browser';
import { RolesService } from '../../../services/roles.service';
import { Restricted } from '../../../decorators/restricted';
import { HostListener } from '@angular/core'
import { RoundProgressConfig } from 'angular-svg-round-progressbar';

@Restricted({
  roles: ['isAuthenticated', 'isPmcUser'],
})
@IonicPage({
  segment: 'project/:projectId/analytics'
})
@Component({
  selector: 'project-analytics',
  templateUrl: 'project-analytics.html'
})

export class ProjectAnalyticsPage {

  projectId: string;
  hasAnalyticsUrl: boolean;
  analyticsUrl: any;
  sanitizedAnalyticsUrl: any;
  index:any;
  timeNow:any;
  span:any;
  claContributors:any = [];
  organizationContributors:any = [];
  firstResponseTimeCurrent: any;
  firstResponseTimeGoal: any;
  closeTimeCurrent: any;
  closeTimeGoal: any;
  sumOpenPRs: any;
  newContributors: any;
  totalContributors: any;

  constructor(
    public navCtrl: NavController,
    public navParams: NavParams,
    private cincoService: CincoService,
    private keycloak: KeycloakService,
    private analyticsService: AnalyticsService,
    private domSanitizer : DomSanitizer,
    public modalCtrl: ModalController,
    public rolesService: RolesService,
    private gaugeConfig: RoundProgressConfig
  ) {
    this.projectId = navParams.get('projectId');
    this.getDefaults();
  }

  ngOnInit() {
    this.getProjectConfig(this.projectId);
  }

  getDefaults() {
    this.setTimeNow();
    this.span = 'month';
    this.index = 'hyperledger5';
    this.getCommitActivity(this.span);
    this.getcommitsDistribution('year');
    this.getIssuesStatus(this.span);
    this.getIssuesActivity(this.span);
    this.getPrsPipeline('year');
    this.getPrsActivity(this.span);
    this.getPageViews(this.span);
    this.getMaintainers('year');
    this.redrawCharts();
    this.sumOpenPRs = 0;
    this.claContributors = [{
      name: "Nick Young",
      email: "swaggyp@dubs.com",
      date: "11/29/17",
    },{
      name: "Patrick McCaw",
      email: "pmccaw@unlv.edu",
      date: "11/29/17",
    },{
      name: "David West",
      email: "david@west.com",
      date: "11/28/17",
    },{
      name: "Javale McGee	",
      email: "javale@fools.com",
      date: "11/27/17",
    },{
      name: "Shaun Livingston",
      email: "shaun@living.net",
      date: "11/27/17",
    },{
      name: "Andre Iguodala",
      email: "iggie@dubs.com",
      date: "11/27/17",
    }];
    this.firstResponseTimeCurrent = 1.6;
    this.firstResponseTimeGoal = 1.5;
    this.closeTimeCurrent = 9.2;
    this.closeTimeGoal = 10;
    this.newContributors = 28;
    this.totalContributors = 826;
    this.organizationContributors= [{
      name: "Google",
      commits: "2,745,342",
      distribution: "21.21%",
    },{
      name: "Intel",
      commits: "811,861",
      distribution: "6.33%",
    },{
      name: "Red Hat",
      commits: "447,876",
      distribution: "3.49%",
    },{
      name: "Code Aurora Forum",
      commits: "327,851",
      distribution: "2.56%",
    },{
      name: "SUSE",
      commits: "303,751",
      distribution: "2.37%",
    },{
      name: "Linux Foundation",
      commits: "268,299",
      distribution: "2.09%",
    },{
      name: "Linaro",
      commits: "253,494",
      distribution: "1.98%",
    },{
      name: "IBM",
      commits: "227,232",
      distribution: "1.73%",
    },{
      name: "Samsung",
      commits: "186,446",
      distribution: "1.45%",
    },{
      name: "Other",
      commits: "156,256",
      distribution: "1.36%",
    }];
    this.gaugeConfig.setDefaults({
      color: '#2bb3e2',
      semicircle: true,
      stroke: 30,
      rounded: true,
      responsive: true,
    });
  }

  setTimeNow() {
    this.timeNow = new Date().getTime();
  }

  calculateTsFrom(span) {
    let rest;
    if(span == 'year') { rest = 365; }
    else if(span == 'quarter') { rest = 90; }
    else if(span == 'month') { rest = 30; }
    else if(span == 'week') { rest = 7; }
    else if(span == 'day') { rest = 1; }
    else { rest = 30; } // otherwise query to a month
    let date = new Date();
    let previousDate = date.getDate() - rest;
    date.setDate(previousDate);
    let tsFrom = date.getTime();
    return tsFrom;
  }

  getProjectConfig(projectId) {
    this.cincoService.getProjectConfig(projectId).subscribe(response => {
      if (response) {
        let projectConfig = response;
        if(projectConfig.analyticsUrl) {
          this.analyticsUrl = projectConfig.analyticsUrl;
          this.sanitizedAnalyticsUrl = this.domSanitizer.bypassSecurityTrustResourceUrl(this.analyticsUrl);
          this.hasAnalyticsUrl = true;
        }
        else{
          this.hasAnalyticsUrl = true;
        }
      }
    });
  }

  openAnaylticsConfigModal(projectId) {
    let modal = this.modalCtrl.create('AnalyticsConfigModal', {
      projectId: projectId,
    });
    modal.onDidDismiss(analyticsUrl => {
      if(analyticsUrl){
        this.analyticsUrl = analyticsUrl;
        this.hasAnalyticsUrl = true;
        this.sanitizedAnalyticsUrl = this.domSanitizer.bypassSecurityTrustResourceUrl(this.analyticsUrl);
      }
    });
    modal.present();
  }

  formatDate(date) {
    var d = new Date(date),
        month = '' + (d.getMonth() + 1),
        day = '' + d.getDate(),
        year = d.getFullYear();
    if (month.length < 2) month = '0' + month;
    if (day.length < 2) day = '0' + day;
    return [month, day].join('-');
  }

  isEmpty(object) {
    if(!Object.keys(object).length) { return true; }
    else { return false; }
  }

  getCommitActivity(span) {
    let index = this.index;
    let metricType = 'code.commits';
    let groupBy = 'day';
    let tsFrom = this.calculateTsFrom(span);
    let tsTo = this.timeNow;
    this.analyticsService.getMetrics(index, metricType, groupBy, tsFrom, tsTo).subscribe(metrics => {
      this.commitsActivityChart.dataTable = [
        ['Date', 'Commits'] // Clean Array
      ];
      if(!this.isEmpty(metrics.value)) { // Check Object response is not empty
        Object.entries(metrics.value).forEach(
          ([key, value]) => {
            this.commitsActivityChart.dataTable.push([this.formatDate(key), value]);
          }
        );
      }
      else {
        this.commitsActivityChart.dataTable.push(['No commits for a ' + span + ' now', 0]);
      }
      this.commitsActivityChart = Object.create(this.commitsActivityChart);
    });
  }

  getcommitsDistribution(span) {
    let index = this.index;
    let metricType = 'maintainers';
    let groupBy = 'year,author';
    let tsFrom = this.calculateTsFrom(span);
    let tsTo = this.timeNow;
    let maintainers;
    let maintainersCommitsTop10 = 0;
    let maintainersCommitsTotal = 0;
    let top10Percentage = 0;
    let restPercentage = 0;
    this.analyticsService.getMetrics(index, metricType, groupBy, tsFrom, tsTo).subscribe(metrics => {
      this.commitsDistributionChart.dataTable = [
        ['Date', 'Commits'] // Clean Array
      ];
      if(!this.isEmpty(metrics.value)) {
        Object.entries(metrics.value).forEach(
          ([key, value]) => {
            if(value) {
              maintainers = value;
            }
          }
        );
        let i = 0;
        Object.entries(maintainers.value).forEach(
          ([key, value]) => {
            if(value) {
              if(i < 10) {
                maintainersCommitsTop10 = maintainersCommitsTop10 + value;
              }
              maintainersCommitsTotal = maintainersCommitsTotal + value;
              i++
            }
          }
        );
        top10Percentage = Math.round( maintainersCommitsTop10 * 100 / maintainersCommitsTotal );
        restPercentage = 100 - top10Percentage;
        this.commitsDistributionChart.dataTable.push(['Top 10', top10Percentage])
        this.commitsDistributionChart.dataTable.push(['Rest', restPercentage]);
      }
      else {
        this.commitsDistributionChart.dataTable.push(['No commits for a ' + span + ' now', 100]);
      }
      this.commitsDistributionChart = Object.create(this.commitsDistributionChart);
    });
  }

  getIssuesStatus(span) {
    let index = this.index;
    let metricType = 'issues';
    let groupBy = 'year,issue_status';
    let tsFrom = this.calculateTsFrom(span);
    let tsTo = this.timeNow;
    let issuesStatus;
    this.analyticsService.getMetrics(index, metricType, groupBy, tsFrom, tsTo).subscribe(metrics => {
      this.issuesStatusChart.dataTable = [
        ['Status', 'Issues'] // Clean Array
      ];
      if(!this.isEmpty(metrics.value)) {
        Object.entries(metrics.value).forEach(
          ([key, value]) => {
            issuesStatus = value;
          }
        );
        Object.entries(issuesStatus.value).forEach(
          ([key, value]) => {
            this.issuesStatusChart.dataTable.push([key, value]);
          }
        );
      }
      else {
        this.issuesStatusChart.dataTable.push(['No issues for a ' + span + ' now', 0]);
      }
      this.issuesStatusChart = Object.create(this.issuesStatusChart);
    });
  }

  getIssuesActivity(span) {
    let index = this.index;
    let metricType = 'issues';
    let groupBy = 'day,issue_status';
    let tsFrom = this.calculateTsFrom(span);
    let tsTo = this.timeNow;
    this.analyticsService.getMetrics(index, metricType, groupBy, tsFrom, tsTo).subscribe(metrics => {
      this.issuesActivityChart.dataTable = [
        ['Date', 'Issues Open', 'Issues Closed'] // Clean Array
      ];
      if(!this.isEmpty(metrics.value)) {
        Object.entries(metrics.value).forEach(
          ([key, value]) => {
            key = this.formatDate(key);
            if(value.value['open'] && !value.value['closed']) this.issuesActivityChart.dataTable.push([key, value.value['open'], 0]);
            if(!value.value['open']  && value.value['closed']) this.issuesActivityChart.dataTable.push([key, 0, value.value['closed']]);
            if(value.value['open']  && value.value['closed']) this.issuesActivityChart.dataTable.push([key, value.value['open'], value.value['closed']]);
          }
        );
      }
      else {
        this.issuesActivityChart.dataTable.push(['No issues for a ' + span + ' now', 0, 0]);
      }
      this.issuesActivityChart = Object.create(this.issuesActivityChart);
    });
  }

  getPrsPipeline(span) {
    let index = this.index;
    let metricType = 'prs.open';
    let groupBy = 'year';
    let tsFrom = this.calculateTsFrom(span);
    let tsTo = this.timeNow;
    this.analyticsService.getMetrics(index, metricType, groupBy, tsFrom, tsTo).subscribe(metrics => {
      this.sumOpenPRs = 0; // Clean Data
      if(!this.isEmpty(metrics.value)) {
        Object.entries(metrics.value).forEach(
          ([key, value]) => {
            this.sumOpenPRs = this.sumOpenPRs + value;
          }
        );
      }
    });
  }

  getPrsActivity(span) {
    let index = this.index;
    let metricType = 'prs';
    let groupBy = 'day,issue_status';
    let tsFrom = this.calculateTsFrom(span);
    let tsTo = this.timeNow;
    this.analyticsService.getMetrics(index, metricType, groupBy, tsFrom, tsTo).subscribe(metrics => {
      this.prsActivityChart.dataTable = [
        ['Date', 'PRs Open', 'PRs Merged', 'PRs Closed'] // Clean Array
      ];
      if(!this.isEmpty(metrics.value)) {
        Object.entries(metrics.value).forEach(
          ([key, value]) => {
            key = this.formatDate(key);
            if(value.value.open && !value.value.merged && !value.value.closed) {
              this.prsActivityChart.dataTable.push([key, value.value.open, 0, 0]);
            }
            if(value.value.open && value.value.merged && !value.value.closed) {
              this.prsActivityChart.dataTable.push([key, value.value.open, value.value.merged, 0]);
            }
            if(value.value.open && value.value.merged && value.value.closed) {
              this.prsActivityChart.dataTable.push([key, value.value.open, value.value.merged, value.value.closed]);
            }
            if(value.value.merged && !value.value.open && !value.value.closed) {
              this.prsActivityChart.dataTable.push([key, 0, value.value.merged, 0]);
            }
            if(value.value.merged && !value.value.open && value.value.closed) {
              this.prsActivityChart.dataTable.push([key, 0, value.value.merged, value.value.closed]);
            }
            if(value.value.closed && !value.value.open && !value.value.merged) {
              this.prsActivityChart.dataTable.push([key, 0, 0, value.value.closed]);
            }
            if(value.value.closed && value.value.open && !value.value.merged) {
              this.prsActivityChart.dataTable.push([key, value.value.open, 0, value.value.closed]);
            }
          }
        );
      }
      else {
        this.prsActivityChart.dataTable.push(['No PRs for a ' + span + ' now', 0, 0, 0]);
      }
      this.prsActivityChart = Object.create(this.prsActivityChart);
    });
  }

  getPageViews(span) {
    let index = this.index;
    //TODO: To query actual Page Views. EP doens't exist yet.
    let metricType = 'code.commits';
    let groupBy = 'day';
    let tsFrom = this.calculateTsFrom(span);
    let tsTo = this.timeNow;
    this.analyticsService.getMetrics(index, metricType, groupBy, tsFrom, tsTo).subscribe(metrics => {
      this.pageViewsChart.dataTable = [
        ['Date', 'Views'] // Clean Array
      ];
      if(!this.isEmpty(metrics.value)) {
        Object.entries(metrics.value).forEach(
          ([key, value]) => {
            key = this.formatDate(key);
            this.pageViewsChart.dataTable.push([key, value]);
          }
        );
      }
      else {
        this.pageViewsChart.dataTable.push(['No page views for a ' + span + ' now', 0]);
      }
      this.pageViewsChart = Object.create(this.pageViewsChart);
    });

  }

  getMaintainers(span) {
    let index = this.index;
    let metricType = 'maintainers';
    let groupBy = 'year,author';
    let tsFrom = this.calculateTsFrom(span);
    let tsTo = this.timeNow;
    let maintainers;
    this.analyticsService.getMetrics(index, metricType, groupBy, tsFrom, tsTo).subscribe(metrics => {
      this.maintainersTable.dataTable = [
        ['Contributor', 'Commits'] // Clean Array
      ];
      if(!this.isEmpty(metrics.value)) {
        Object.entries(metrics.value).forEach(
          ([key, value]) => {
            if(value) {
              maintainers = value;
            }
          }
        );
        Object.entries(maintainers.value).forEach(
          ([key, value]) => {
            if(value) {
              this.maintainersTable.dataTable.push([key, value]);
            }
          }
        );
      }
      else {
        this.maintainersTable.dataTable.push(['No maintainers for a ' + span + ' now', 0]);
      }
      this.maintainersTable = Object.create(this.maintainersTable);
    });
  }

  redrawCharts() {
    this.commitsActivityChart = Object.create(this.commitsActivityChart);
    this.commitsDistributionChart = Object.create(this.commitsDistributionChart);
    this.issuesStatusChart = Object.create(this.issuesStatusChart);
    this.issuesActivityChart = Object.create(this.issuesActivityChart);
    this.prsActivityChart = Object.create(this.prsActivityChart);
    this.pageViewsChart = Object.create(this.pageViewsChart);
  }

  @HostListener('window:resize', ['$event'])
  onResize(event) {
    event.target.innerWidth;
    this.redrawCharts();
  }

  public commitsActivityChart:any =  {
    chartType: 'ColumnChart',
    dataTable: [
      ['Date', 'Commits']
    ],
    options: {
      hAxis: {
        textStyle:{ color: '#0b4e73'},
        gridlines: {
          color: "#0b4e73"
        },
        baselineColor: '#0b4e73',
        format: 'h:mm a',
      },
      vAxis: {title: '# of commits'},
      colors: ['#2bb3e2'],
      backgroundColor: '#ffffff',
      legend: 'none',
    }
  };

  public commitsDistributionChart:any =  {
    chartType: 'PieChart',
    dataTable: [
      ['Date', 'Commits']
    ],
    options: {
      hAxis: {
        textStyle:{ color: '#1cb2e4'},
        gridlines: {
          color: "#1cb2e4"
        },
        baselineColor: '#1cb2e4'
      },
      chartArea: {width: 400, height: 300},
      colors: ['#1cb2e4','#ebebeb'],
      backgroundColor: '#ffffff',
      legend: 'none'
    }
  };

  public issuesStatusChart:any =  {
    chartType: 'BarChart',
    dataTable: [
      ['Status', 'Issues']
    ],
    options: {
      hAxis: {
        textStyle:{ color: '#0b4e73'},
        gridlines: {
          color: "#0b4e73"
        },
        baselineColor: '#0b4e73'
      },
      vAxis: {},
      colors: ['#2bb3e2'],
      backgroundColor: '#ffffff',
      legend: 'none',
    }
  };

  public issuesActivityChart:any =  {
    chartType: 'AreaChart',
    dataTable: [
      ['Date', 'Issues Open', 'Issues Closed']
    ],
    options: {
      hAxis: {
        textStyle:{ color: '#0b4e73'},
        gridlines: {
          color: "#0b4e73"
        },
        baselineColor: '#FFFFFF'
      },
      vAxis: {title: '# of Issues'},
      colors: ['#2bb3e2', '#0b4e73'],
      backgroundColor: '#ffffff',
      legend: 'none'
    }
  };

  public prsActivityChart:any =  {
    chartType: 'AreaChart',
    dataTable: [
      ['Date', 'PRs Open', 'PRs Merged', 'PRs Closed']
    ],
    options: {
      hAxis: {
        textStyle:{ color: '#0b4e73'},
        gridlines: {
          color: "#0b4e73"
        },
        baselineColor: '#0b4e73'
      },
      vAxis: {title: '# of PRs'},
      colors: ['#2bb3e2', '#0b4e73'],
      backgroundColor: '#ffffff',
      legend: 'none'
    }
  };

  public pageViewsChart:any =  {
    chartType: 'AreaChart',
    dataTable: [
      ['Date', 'Page Views']
    ],
    options: {
      hAxis: {
        textStyle:{ color: '#0b4e73'},
        gridlines: {
          color: "#0b4e73"
        },
        baselineColor: '#0b4e73'
      },
      vAxis: {title: 'Page Views (in thousands)', minValue: 0, maxValue: 15},
      colors: ['#2bb3e2'],
      backgroundColor: '#ffffff',
      legend: 'none'
    }
  };

  public cssClassNames:any = {
    'headerRow': 'header-row',
    'tableRow': 'table-row',
    'oddTableRow': 'odd-table-row',
    'selectedTableRow': 'selected-table-row',
    'hoverTableRow': 'hover-table-row',
    'headerCell': 'header-cell',
    'tableCell': 'table-cell',
    'rowNumberCell': 'row-number-cell'
  };

  public maintainersTable:any =  {
    chartType: 'Table',
    dataTable: [
      ['Contributor', 'Commits'],
    ],
    options: {
      title: 'Maintainers',
      allowHtml: true,
      alternatingRowStyle: false,
      width: '100%',
      cssClassNames: this.cssClassNames,
      sortColumn: 1,
      sortAscending: false,
    }
  };

}
