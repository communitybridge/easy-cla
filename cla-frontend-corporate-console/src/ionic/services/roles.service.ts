import { Injectable } from "@angular/core";
import { Observable } from "rxjs/Observable";
import { KeycloakService } from "./keycloak/keycloak.service";
import { AuthService } from "./auth.service";

@Injectable()
export class RolesService {
  public userAuthenticated: boolean;
  public userRoleDefaults: any;
  public userRoles: any;
  private getDataObserver: any;
  public getData: any;
  private rolesFetched: boolean;

  private LF_CUSTOM_CLAIM = "https://sso.linuxfoundation.org/claims/roles";
  private CLA_PROJECT_ADMIN = "cla-system-admin";
  private projectSet = new Set(["cla-admin-project-mvp"]); // Here we may need to generate this array from Salesforce API ??

  constructor(
    private keycloak: KeycloakService,
    private authService: AuthService
  ) {
    this.rolesFetched = false;
    this.userRoleDefaults = {
      isAuthenticated: this.authService.isAuthenticated(),
      isPmcUser: false,
      isStaffInc: false,
      isDirectorInc: false,
      isStaffDirect: false,
      isDirectorDirect: false,
      isExec: false,
      isAdmin: false
    };
    this.userRoles = this.userRoleDefaults;
  }

  //////////////////////////////////////////////////////////////////////////////

  /**
   * This service should ONLY contain methods for user roles
   **/

  //////////////////////////////////////////////////////////////////////////////
  //////////////////////////////////////////////////////////////////////////////

  getUserRolesPromise() {
    console.log("Get UserRole Promise.");
    if (this.authService.isAuthenticated()) {
      return this.authService
        .getIdToken()
        .then(token => {
          return this.authService.parseIdToken(token);
        })
        .then(tokenParsed => {
          if (tokenParsed && tokenParsed[this.LF_CUSTOM_CLAIM]) {
            let customRules = tokenParsed[this.LF_CUSTOM_CLAIM];
            this.userRoles = {
              isAuthenticated: this.authService.isAuthenticated(),
              isPmcUser: this.isInProjectSet(customRules, this.projectSet),
              isStaffInc: false,
              isDirectorInc: false,
              isStaffDirect: false,
              isDirectorDirect: false,
              isExec: false,
              isAdmin: this.isInArray(customRules, this.CLA_PROJECT_ADMIN)
            };
            console.log(this.userRoles);
            return this.userRoles;
          }
          return this.userRoleDefaults;
        })
        .catch(error => {
          return Promise.resolve(this.userRoleDefaults);
        });
    } else {
      // not authenticated. can't decode token. just return defaults
      return Promise.resolve(this.userRoleDefaults);
    }
  }

  private isInArray(roles, role) {
    for (let i = 0; i < roles.length; i++) {
      if (roles[i].toLowerCase() === role.toLowerCase()) {
        return true;
      }
    }
    return false;
  }

  private isInProjectSet(roles, projectSet) {
    return true;

    // for (let i = 0; i < roles.length; i++) {
    //   if (projectSet.has(roles[i])) {
    //     return true;
    //   }
    // }
    // return false;
  }

  //////////////////////////////////////////////////////////////////////////////
}
