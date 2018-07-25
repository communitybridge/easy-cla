export class MemberModel {

  // This project definition is based on CINCO ProjectMember and Organization class
  id: string;
  projectId: string;
  projectName: string;
  org: {
    id: string,
    name: string,
    parent: string,
    logoRef: string,
    url: string,
    addresses: any
  }
  product: string;
  tier: string;
  annualDues:  any;
  startDate: any;
  renewalDate: any;
  invoices: any[];

  constructor() {
  }

}
