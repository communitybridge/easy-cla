// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package gitlab_organizations

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"

	"github.com/communitybridge/easycla/cla-backend-go/v2/repositories"

	"github.com/communitybridge/easycla/cla-backend-go/v2/common"

	"github.com/communitybridge/easycla/cla-backend-go/config"
	gitlab_api "github.com/communitybridge/easycla/cla-backend-go/gitlab_api"
	"github.com/go-openapi/strfmt"

	project_service "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"

	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	v2ProjectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
	"github.com/sirupsen/logrus"
	goGitLab "github.com/xanzy/go-gitlab"
)

// ServiceInterface contains functions of GitlabOrganizations service
type ServiceInterface interface {
	AddGitLabOrganization(ctx context.Context, projectSFID string, input *models.GitlabCreateOrganization) (*models.GitlabProjectOrganizations, error)
	GetGitLabOrganization(ctx context.Context, gitLabOrganizationID string) (*models.GitlabOrganization, error)
	GetGitLabOrganizationByID(ctx context.Context, gitLabOrganizationID string) (*common.GitLabOrganization, error)
	GetGitLabOrganizationByName(ctx context.Context, gitLabOrganizationName string) (*models.GitlabOrganization, error)
	GetGitLabOrganizationByFullPath(ctx context.Context, gitLabOrganizationFullPath string) (*models.GitlabOrganization, error)
	GetGitLabOrganizationByGroupID(ctx context.Context, gitLabGroupID int64) (*models.GitlabOrganization, error)
	GetGitLabOrganizations(ctx context.Context, projectSFID string) (*models.GitlabProjectOrganizations, error)
	GetGitLabOrganizationByState(ctx context.Context, gitLabOrganizationID, authState string) (*models.GitlabOrganization, error)
	UpdateGitLabOrganization(ctx context.Context, projectSFID string, groupID int64, organizationName, groupFullPath string, autoEnabled bool, autoEnabledClaGroupID string, branchProtectionEnabled bool) error
	UpdateGitLabOrganizationAuth(ctx context.Context, gitLabOrganizationID string, oauthResp *gitlab_api.OauthSuccessResponse) error
	DeleteGitLabOrganizationByFullPath(ctx context.Context, projectSFID string, gitlabOrgFullPath string) error
}

// Service data model
type Service struct {
	repo               RepositoryInterface
	v2GitRepoService   repositories.ServiceInterface
	claGroupRepository projects_cla_groups.Repository
	gitLabApp          *gitlab_api.App
}

// NewService creates a new gitlab organization service
func NewService(repo RepositoryInterface, v2GitRepoService repositories.ServiceInterface, claGroupRepository projects_cla_groups.Repository) *Service {
	return &Service{
		repo:               repo,
		v2GitRepoService:   v2GitRepoService,
		claGroupRepository: claGroupRepository,
		gitLabApp:          gitlab_api.Init(config.GetConfig().Gitlab.AppClientID, config.GetConfig().Gitlab.AppClientSecret, config.GetConfig().Gitlab.AppPrivateKey),
	}
}

// AddGitLabOrganization adds the specified GitLab organization
func (s *Service) AddGitLabOrganization(ctx context.Context, projectSFID string, input *models.GitlabCreateOrganization) (*models.GitlabProjectOrganizations, error) {
	f := logrus.Fields{
		"functionName":            "v2.gitlab_organizations.service.AddGitLabOrganization",
		utils.XREQUESTID:          ctx.Value(utils.XREQUESTID),
		"projectSFID":             projectSFID,
		"autoEnabled":             utils.BoolValue(input.AutoEnabled),
		"branchProtectionEnabled": utils.BoolValue(input.BranchProtectionEnabled),
		"groupID":                 input.GroupID,
		"groupFullPath":           input.OrganizationFullPath,
	}

	var existingModel *models.GitlabOrganization
	var getErr error
	if input.OrganizationFullPath != "" {
		existingModel, getErr = s.GetGitLabOrganizationByFullPath(ctx, input.OrganizationFullPath)
		if getErr != nil {
			log.WithFields(f).WithError(getErr).Warnf("problem querying GitLab group/organization using full path: %s", input.OrganizationFullPath)
			return nil, getErr
		}
	}
	if input.GroupID > 0 {
		existingModel, getErr = s.GetGitLabOrganizationByGroupID(ctx, input.GroupID)
		if getErr != nil {
			log.WithFields(f).WithError(getErr).Warnf("problem querying GitLab group/organization using group ID: %d", input.GroupID)
			return nil, getErr
		}
	}

	// If we have an existing record/entry
	if existingModel != nil {
		// Check to make sure another project doesn't own this GitLab Group - only care about conflicts if it is enabled
		if existingModel.ProjectSfid != projectSFID && existingModel.Enabled {
			psc := project_service.GetClient()
			requestedProjectModel, projectLookupErr := psc.GetProject(projectSFID)
			if projectLookupErr != nil || requestedProjectModel == nil {
				return nil, projectLookupErr
			}
			existingProjectModel, projectLookupErr := psc.GetProject(existingModel.ProjectSfid)
			if projectLookupErr != nil || existingProjectModel == nil {
				log.WithFields(f).WithError(projectLookupErr).Warnf("unable to lookup project with SFID: %s", existingModel.ProjectSfid)
				return nil, projectLookupErr
			}
			msg := fmt.Sprintf("unable to add or update the GitLab Group/Organization - already taken by another project: %s (%s) - unable to add to this project: %s (%s)",
				existingProjectModel.Name, existingModel.ProjectSfid,
				requestedProjectModel.Name, projectSFID)
			log.WithFields(f).Warn(msg)

			// Return the error model
			return nil, &utils.ProjectConflict{
				Message: "unable to add or update the GitLab Group/Organization - already taken by another project",
				ProjectA: utils.ProjectSummary{
					Name: requestedProjectModel.Name,
					ID:   projectSFID,
				},
				ProjectB: utils.ProjectSummary{
					Name: existingProjectModel.Name,
					ID:   existingModel.ProjectSfid,
				},
			}
		}

		updateErr := s.UpdateGitLabOrganization(ctx, projectSFID, input.GroupID, "", input.OrganizationFullPath, utils.BoolValue(input.AutoEnabled), input.AutoEnabledClaGroupID, utils.BoolValue(input.BranchProtectionEnabled))
		if updateErr != nil {
			log.WithFields(f).WithError(updateErr).Warnf("problem updating GitLab group/organization, error: %+v", updateErr)
			return nil, getErr
		}
		return s.GetGitLabOrganizations(ctx, projectSFID)
	}

	psc := v2ProjectService.GetClient()
	project, err := psc.GetProject(projectSFID)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem loading project details from the project service")
		return nil, err
	}

	var parentProjectSFID string
	if utils.StringValue(project.Parent) == "" || (project.Foundation != nil &&
		(project.Foundation.Name == utils.TheLinuxFoundation || project.Foundation.Name == utils.LFProjectsLLC)) {
		parentProjectSFID = projectSFID
	} else {
		parentProjectSFID = utils.StringValue(project.Parent)
	}
	f["parentProjectSFID"] = parentProjectSFID
	log.WithFields(f).Debug("located parentProjectID...")

	log.WithFields(f).Debug("adding GitLab organization...")
	autoEnabled := false
	if input.AutoEnabled != nil {
		autoEnabled = utils.BoolValue(input.AutoEnabled)
	}
	branchProtectionEnabled := false
	if input.BranchProtectionEnabled != nil {
		branchProtectionEnabled = utils.BoolValue(input.BranchProtectionEnabled)
	}

	resp, err := s.repo.AddGitLabOrganization(ctx, parentProjectSFID, projectSFID, input.GroupID, "", input.OrganizationFullPath, autoEnabled, input.AutoEnabledClaGroupID, branchProtectionEnabled, true)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem adding gitlab organization for project")
		return nil, err
	}
	log.WithFields(f).Debugf("created GitLab organization with ID: %s", resp.OrganizationID)

	return s.GetGitLabOrganizations(ctx, projectSFID)
}

// GetGitLabOrganization returns the GitLab organization based on the specified GitLab Organization ID
func (s *Service) GetGitLabOrganization(ctx context.Context, gitlabOrganizationID string) (*models.GitlabOrganization, error) {
	dbModel, err := s.GetGitLabOrganizationByID(ctx, gitlabOrganizationID)
	if err != nil {
		return nil, err
	}

	if dbModel == nil {
		return nil, nil
	}

	return common.ToModel(dbModel), err
}

// GetGitLabOrganizationByID returns the record associated with the GitLab Organization ID
func (s *Service) GetGitLabOrganizationByID(ctx context.Context, gitLabOrganizationID string) (*common.GitLabOrganization, error) {
	f := logrus.Fields{
		"functionName":         "v2.gitlab_organizations.service.GetGitLabOrganizationByID",
		utils.XREQUESTID:       ctx.Value(utils.XREQUESTID),
		"gitLabOrganizationID": gitLabOrganizationID,
	}

	log.WithFields(f).Debugf("fetching gitlab organization for gitlab org id: %s", gitLabOrganizationID)
	dbModel, err := s.repo.GetGitLabOrganization(ctx, gitLabOrganizationID)
	if err != nil {
		return nil, err
	}

	return dbModel, nil
}

// GetGitLabOrganizationByName returns the gitlab organization based on the Group/Org name
func (s *Service) GetGitLabOrganizationByName(ctx context.Context, gitLabOrganizationName string) (*models.GitlabOrganization, error) {
	f := logrus.Fields{
		"functionName":         "v2.gitlab_organizations.service.GetGitLabOrganizationByName",
		utils.XREQUESTID:       ctx.Value(utils.XREQUESTID),
		"gitlabOrganizationID": gitLabOrganizationName,
	}

	log.WithFields(f).Debugf("fetching gitlab organization for gitlab org id: %s", gitLabOrganizationName)
	dbModel, err := s.repo.GetGitLabOrganizationByName(ctx, gitLabOrganizationName)
	if err != nil {
		return nil, err
	}

	return common.ToModel(dbModel), nil
}

// GetGitLabOrganizationByFullPath returns the GitLab group/organization using the specified full path
func (s *Service) GetGitLabOrganizationByFullPath(ctx context.Context, gitLabOrganizationFullPath string) (*models.GitlabOrganization, error) {
	f := logrus.Fields{
		"functionName":               "v2.gitlab_organizations.service.GetGitLabOrganizationByFullPath",
		utils.XREQUESTID:             ctx.Value(utils.XREQUESTID),
		"gitLabOrganizationFullPath": gitLabOrganizationFullPath,
	}

	log.WithFields(f).Debugf("fetching gitlab group/organization using full path: %s", gitLabOrganizationFullPath)
	dbModel, err := s.repo.GetGitLabOrganizationByFullPath(ctx, gitLabOrganizationFullPath)
	if err != nil {
		return nil, err
	}

	if dbModel == nil {
		return nil, nil
	}
	return common.ToModel(dbModel), nil
}

// GetGitLabOrganizationByGroupID returns the GitLab group/organization using the specified group ID
func (s *Service) GetGitLabOrganizationByGroupID(ctx context.Context, gitLabGroupID int64) (*models.GitlabOrganization, error) {
	f := logrus.Fields{
		"functionName":   "v2.gitlab_organizations.service.GetGitLabOrganizationByGroupID",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"gitLabGroupID":  gitLabGroupID,
	}

	log.WithFields(f).Debugf("fetching gitlab group/organization using group ID: %d", gitLabGroupID)
	dbModel, err := s.repo.GetGitLabOrganizationByExternalID(ctx, gitLabGroupID)
	if err != nil {
		return nil, err
	}

	if dbModel == nil {
		return nil, nil
	}

	return common.ToModel(dbModel), nil
}

// GetGitLabOrganizations returns a collection of GitLab organizations based on the specified project SFID value
func (s *Service) GetGitLabOrganizations(ctx context.Context, projectSFID string) (*models.GitlabProjectOrganizations, error) {
	f := logrus.Fields{
		"functionName":   "v2.gitlab_organizations.service.GetGitLabOrganizations",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
	}

	psc := v2ProjectService.GetClient()
	log.WithFields(f).Debug("loading project details from the project service...")
	projectServiceRecord, err := psc.GetProject(projectSFID)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem loading project details from the project service")
		return nil, err
	}

	var parentProjectSFID string
	if utils.IsProjectHasRootParent(projectServiceRecord) {
		parentProjectSFID = projectSFID
	} else {
		parentProjectSFID = utils.StringValue(projectServiceRecord.Parent)
	}
	f["parentProjectSFID"] = parentProjectSFID
	log.WithFields(f).Debug("located parentProjectID...")

	// Load the GitLab Organization and Repository details - result will be missing CLA Group info and ProjectSFID details
	log.WithFields(f).Debugf("loading Gitlab organizations for projectSFID: %s", projectSFID)
	orgList, err := s.repo.GetGitLabOrganizations(ctx, projectSFID)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem loading gitlab organizations from the project service")
		return nil, err
	}
	log.WithFields(f).Debugf("loaded %d Gitlab organizations for projectSFID: %s", len(orgList.List), projectSFID)

	// Our response model
	out := &models.GitlabProjectOrganizations{
		List: make([]*models.GitlabProjectOrganization, 0),
	}

	orgMap := make(map[string]*models.GitlabProjectOrganization)
	for _, org := range orgList.List {
		autoEnabledCLAGroupName := ""
		if org.AutoEnabledClaGroupID != "" {
			log.WithFields(f).Debugf("loading CLA Group by ID: %s to obtain the name for GitLab auth enabled CLA Group response", org.AutoEnabledClaGroupID)
			claGroupMode, claGroupLookupErr := s.claGroupRepository.GetCLAGroup(ctx, org.AutoEnabledClaGroupID)
			if claGroupLookupErr != nil {
				log.WithFields(f).WithError(claGroupLookupErr).Warnf("Unable to lookup CLA Group by ID: %s", org.AutoEnabledClaGroupID)
			}
			if claGroupMode != nil {
				autoEnabledCLAGroupName = claGroupMode.ProjectName
			}
		}

		log.WithFields(f).Debugf("loading GitLab organization by organization ID: %s", org.OrganizationID)
		orgDetailed, orgErr := s.repo.GetGitLabOrganization(ctx, org.OrganizationID)
		if orgErr != nil {
			log.WithFields(f).Errorf("fetching gitlab org failed : %s : %v", org.OrganizationID, orgErr)
			continue
		}

		log.WithFields(f).Debugf("filtering repositories based on group path: %s", org.OrganizationFullPath)
		repoList, repoErr := s.v2GitRepoService.GitLabGetRepositoriesByNamePrefix(ctx, fmt.Sprintf("%s/", org.OrganizationFullPath))
		if repoErr != nil {
			if _, ok := err.(*utils.GitLabRepositoryNotFound); ok {
				log.WithFields(f).WithError(repoErr).Debugf("no GitLab repositories onboarded for project : %s", projectSFID)
			} else {
				log.WithFields(f).WithError(repoErr).Debugf("unexpected error while fetching GitLab group repositories for project: %s, error type: %T, error: %v", projectSFID, repoErr, repoErr)
			}
		}

		rorg := &models.GitlabProjectOrganization{
			AutoEnabled:             org.AutoEnabled,
			AutoEnableClaGroupID:    org.AutoEnabledClaGroupID,
			AutoEnabledClaGroupName: strings.TrimSpace(autoEnabledCLAGroupName),
			OrganizationName:        org.OrganizationName,
			OrganizationURL:         org.OrganizationURL,
			OrganizationFullPath:    org.OrganizationFullPath,
			OrganizationExternalID:  org.OrganizationExternalID,
			InstallationURL:         buildInstallationURL(org.OrganizationID, orgDetailed.AuthState),
			BranchProtectionEnabled: org.BranchProtectionEnabled,
			ConnectionStatus:        "",                                  // updated below
			Repositories:            []*models.GitlabProjectRepository{}, // updated below
		}

		if orgDetailed.AuthInfo == "" {
			rorg.ConnectionStatus = utils.NoConnection
		} else {
			if err != nil {
				log.WithFields(f).Errorf("initializing gitlab client for gitlab org : %s failed : %v", org.OrganizationID, err)
				rorg.ConnectionStatus = utils.ConnectionFailure
			} else {
				// We've been authenticated by the user - great, see if we can determine the list of repos...
				glClient, clientErr := gitlab_api.NewGitlabOauthClient(orgDetailed.AuthInfo, s.gitLabApp)
				if clientErr != nil {
					log.WithFields(f).Errorf("using gitlab client for gitlab org : %s failed : %v", org.OrganizationID, clientErr)
					rorg.ConnectionStatus = utils.ConnectionFailure
				} else {
					rorg.Repositories = s.updateRepositoryStatus(glClient, toGitLabProjectResponse(repoList))

					user, _, userErr := glClient.Users.CurrentUser()
					if userErr != nil {
						log.WithFields(f).Errorf("using gitlab client for gitlab org : %s failed : %v", org.OrganizationID, userErr)
						rorg.ConnectionStatus = utils.ConnectionFailure
					} else {
						log.WithFields(f).Debugf("connected to user : %s for gitlab org : %s", user.Name, org.OrganizationID)
						rorg.ConnectionStatus = utils.Connected
					}
				}
			}
		}

		orgMap[org.OrganizationName] = rorg
		out.List = append(out.List, rorg)
	}

	// Sort everything nicely
	sort.Slice(out.List, func(i, j int) bool {
		return strings.ToLower(out.List[i].OrganizationName) < strings.ToLower(out.List[j].OrganizationName)
	})
	for _, orgList := range out.List {
		sort.Slice(orgList.Repositories, func(i, j int) bool {
			return strings.ToLower(orgList.Repositories[i].RepositoryName) < strings.ToLower(orgList.Repositories[j].RepositoryName)
		})
	}

	return out, nil
}

// GetGitLabOrganizationByState returns the GitLab organization by the auth state
func (s *Service) GetGitLabOrganizationByState(ctx context.Context, gitLabOrganizationID, authState string) (*models.GitlabOrganization, error) {
	f := logrus.Fields{
		"functionName":         "v2.gitlab_organizations.service.GetGitLabOrganization",
		utils.XREQUESTID:       ctx.Value(utils.XREQUESTID),
		"gitLabOrganizationID": gitLabOrganizationID,
		"authState":            authState,
	}

	log.WithFields(f).Debugf("fetching gitlab organization for gitlab org id : %s", gitLabOrganizationID)
	dbModel, err := s.repo.GetGitLabOrganization(ctx, gitLabOrganizationID)
	if err != nil {
		return nil, err
	}

	if dbModel.AuthState != authState {
		return nil, fmt.Errorf("auth state doesn't match")
	}

	return common.ToModel(dbModel), nil
}

// UpdateGitLabOrganizationAuth updates the GitLab organization authentication information
func (s *Service) UpdateGitLabOrganizationAuth(ctx context.Context, gitLabOrganizationID string, oauthResp *gitlab_api.OauthSuccessResponse) error {
	f := logrus.Fields{
		"functionName":         "v2.gitlab_organizations.service.UpdateGitLabOrganizationAuth",
		utils.XREQUESTID:       ctx.Value(utils.XREQUESTID),
		"gitLabOrganizationID": gitLabOrganizationID,
	}

	log.WithFields(f).Debugf("updating gitlab org auth")
	authInfoEncrypted, err := gitlab_api.EncryptAuthInfo(oauthResp, s.gitLabApp)
	if err != nil {
		return fmt.Errorf("encrypt failed : %v", err)
	}

	gitLabOrgModel, err := s.GetGitLabOrganizationByID(ctx, gitLabOrganizationID)
	if err != nil {
		return fmt.Errorf("gitlab organization lookup error: %+v", err)
	}

	// Get a reference to the GItLab client
	gitLabClient, err := gitlab_api.NewGitlabOauthClientFromAccessToken(oauthResp.AccessToken)
	if err != nil {
		return fmt.Errorf("initializing gitlab client : %v", err)
	}

	// Query the groups list
	groupsWithMaintainerPerms, groupListErr := gitlab_api.GetGroupsListAll(ctx, gitLabClient, goGitLab.MaintainerPermissions)
	if groupListErr != nil {
		return groupListErr
	}

	for _, g := range groupsWithMaintainerPerms {
		// If we have an external group ID or a full path...
		if (gitLabOrgModel.ExternalGroupID > 0 && g.ID == gitLabOrgModel.ExternalGroupID) ||
			(gitLabOrgModel.OrganizationFullPath != "" && g.FullPath == gitLabOrgModel.OrganizationFullPath) {

			updateGitLabOrgErr := s.repo.UpdateGitLabOrganizationAuth(ctx, gitLabOrganizationID, g.ID, authInfoEncrypted, g.Name, g.FullPath, g.WebURL)
			if updateGitLabOrgErr != nil {
				return updateGitLabOrgErr
			}

			log.WithFields(f).Debugf("fetching updated GitLab group/organization record which should now have all the details")
			updatedOrgDBModel, getErr := s.repo.GetGitLabOrganization(ctx, gitLabOrganizationID)
			if getErr != nil {
				return getErr
			}

			log.WithFields(f).Debugf("adding GitLab repositories for this group/organization")
			_, err = s.v2GitRepoService.GitLabAddRepositoriesByApp(ctx, updatedOrgDBModel)
			if err != nil {
				return err
			}

			return nil
		}
	}

	msg := ""
	if gitLabOrgModel.ExternalGroupID > 0 {
		msg = fmt.Sprintf("external ID: %d", gitLabOrgModel.ExternalGroupID)
	} else if gitLabOrgModel.OrganizationFullPath != "" {
		msg = fmt.Sprintf("full path: '%s'", gitLabOrgModel.OrganizationFullPath)
	}

	return fmt.Errorf("unable to locate the provided GitLab group by %s using the provided permissions - discovered %d groups where user has maintainer or above permissions.",
		msg, len(groupsWithMaintainerPerms))
}

// UpdateGitLabOrganization updates the GitLab organization
func (s *Service) UpdateGitLabOrganization(ctx context.Context, projectSFID string, groupID int64, organizationName, groupFullPath string, autoEnabled bool, autoEnabledClaGroupID string, branchProtectionEnabled bool) error {
	// check if valid cla group id is passed
	if autoEnabledClaGroupID != "" {
		if _, err := s.claGroupRepository.GetCLAGroupNameByID(ctx, autoEnabledClaGroupID); err != nil {
			return err
		}
	}

	return s.repo.UpdateGitLabOrganization(ctx, projectSFID, groupID, organizationName, groupFullPath, autoEnabled, autoEnabledClaGroupID, branchProtectionEnabled, true)
}

// DeleteGitLabOrganizationByFullPath deletes the specified GitLab organization by full path
func (s *Service) DeleteGitLabOrganizationByFullPath(ctx context.Context, projectSFID string, gitLabOrgFullPath string) error {
	f := logrus.Fields{
		"functionName":      "v2.gitlab_organizations.service.DeleteGitLabOrganizationByFullPath",
		utils.XREQUESTID:    ctx.Value(utils.XREQUESTID),
		"projectSFID":       projectSFID,
		"gitLabOrgFullPath": gitLabOrgFullPath,
	}

	// Check for enabled repos...
	repoList, getRepListErr := s.v2GitRepoService.GitLabGetRepositoriesByProjectSFID(ctx, projectSFID)
	if getRepListErr != nil {
		// If nothing to delete...
		if _, ok := getRepListErr.(*utils.GitLabRepositoryNotFound); ok {
			log.WithFields(f).Debugf("no repositories found under GitLab group/organization: %s", gitLabOrgFullPath)
		} else {
			return getRepListErr
		}
	}

	// Check to see if we still have enabled repos belonging to this GitLab organization/group
	var enabledRepoList []string
	if repoList != nil && len(repoList.List) > 0 {
		for _, repo := range repoList.List {
			if strings.HasPrefix(repo.RepositoryName, gitLabOrgFullPath) && repo.Enabled {
				enabledRepoList = append(enabledRepoList, repo.RepositoryName)
			}
		}
	}

	if len(enabledRepoList) > 0 {
		return fmt.Errorf("the following repositories are still enabled under the GitLab Group/Organization: %s - %s", gitLabOrgFullPath, strings.Join(enabledRepoList, ","))
	}

	// First delete the GitLab project/repos
	log.WithFields(f).Debugf("deleting GitLab repos under group: %s", gitLabOrgFullPath)
	repoDeleteErr := s.v2GitRepoService.GitLabDeleteRepositories(ctx, gitLabOrgFullPath)
	if repoDeleteErr != nil {
		log.WithFields(f).WithError(repoDeleteErr).Warnf("problem deleting GitLab repos under group: %s", gitLabOrgFullPath)
		return repoDeleteErr
	}

	return s.repo.DeleteGitLabOrganizationByFullPath(ctx, projectSFID, gitLabOrgFullPath)
}

func buildInstallationURL(gitlabOrgID string, authStateNonce string) *strfmt.URI {
	base := "https://gitlab.com/oauth/authorize"
	c := config.GetConfig()
	state := fmt.Sprintf("%s:%s", gitlabOrgID, authStateNonce)

	params := url.Values{}
	params.Add("client_id", c.Gitlab.AppClientID)
	params.Add("redirect_uri", c.Gitlab.RedirectURI)
	//params.Add("redirect_uri", "http://localhost:8080/v4/gitlab/oauth/callback")
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("scope", "api read_user read_api read_repository write_repository email")

	installationURL := strfmt.URI(base + "?" + params.Encode())
	return &installationURL
}

func toGitLabProjectResponse(gitLabListRepos *models.GitlabRepositoriesList) []*models.GitlabProjectRepository {
	f := logrus.Fields{
		"functionName": "v2.gitlab_organizations.service.toGitLabProjectResponse",
	}

	if gitLabListRepos == nil {
		return []*models.GitlabProjectRepository{}
	}

	var repoList []*models.GitlabProjectRepository
	for _, repo := range gitLabListRepos.List {
		parentProjectSFID, err := project_service.GetClient().GetParentProject(repo.RepositoryProjectSfid)
		if err != nil {
			log.WithFields(f).Warnf("unable to lookup project parent SFID using SFID: %s", repo.RepositoryProjectSfid)
		}

		repoList = append(repoList, &models.GitlabProjectRepository{
			ClaGroupID: repo.RepositoryClaGroupID,
			//ConnectionStatus:   "", // set via another function
			Enabled:            repo.Enabled,
			ParentProjectID:    parentProjectSFID,
			ProjectID:          repo.RepositoryProjectSfid,
			RepositoryGitlabID: repo.RepositoryExternalID,
			RepositoryID:       repo.RepositoryID,
			RepositoryName:     repo.RepositoryName,
			RepositoryFullPath: repo.RepositoryFullPath,
			RepositoryURL:      repo.RepositoryURL,
		})
	}

	return repoList
}

// updateRepositoryStatus is a helper function to set/add the repository connection status
func (s *Service) updateRepositoryStatus(glClient *goGitLab.Client, gitLabProjectRepos []*models.GitlabProjectRepository) []*models.GitlabProjectRepository {
	f := logrus.Fields{
		"functionName": "v2.gitlab_organizations.service.updateRepositoryStatus",
	}

	if gitLabProjectRepos == nil {
		return []*models.GitlabProjectRepository{}
	}

	type responseChannelModel struct {
		RepositoryExternalID int64
		ConnectionStatus     string
		Error                error
	}
	// A channel for the responses from the go routines
	responseChannel := make(chan responseChannelModel)

	opts := &goGitLab.GetProjectOptions{}
	for _, repo := range gitLabProjectRepos {
		// Create a go routine to this concurrently
		go func(glClient *goGitLab.Client, repo *models.GitlabProjectRepository) {
			projectModel, resp, lookupErr := glClient.Projects.GetProject(int(repo.RepositoryGitlabID), opts) // OK to convert int64 to int as it is the ID and probably should be typed as a int anyway
			if lookupErr != nil {
				log.WithFields(f).WithError(lookupErr).Warnf("problem loading GitLab project by external ID: %d, error: %v", repo.RepositoryGitlabID, lookupErr)
				repo.ConnectionStatus = utils.ConnectionFailure
				responseChannel <- responseChannelModel{
					RepositoryExternalID: repo.RepositoryGitlabID,
					ConnectionStatus:     utils.ConnectionFailure,
					Error:                lookupErr,
				}
				return
			}
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				responseBody, readErr := io.ReadAll(resp.Body)
				if readErr != nil {
					log.WithFields(f).WithError(lookupErr).Warnf("problem loading GitLab project by external ID: %d, error: %v", repo.RepositoryGitlabID, readErr)
					responseChannel <- responseChannelModel{
						RepositoryExternalID: repo.RepositoryGitlabID,
						ConnectionStatus:     utils.ConnectionFailure,
						Error:                readErr,
					}
					return
				}
				msg := fmt.Sprintf("problem loading GitLab project by external ID: %d, response code: %d, body: %s", repo.RepositoryGitlabID, resp.StatusCode, responseBody)
				log.WithFields(f).Warnf(msg)
				responseChannel <- responseChannelModel{
					RepositoryExternalID: repo.RepositoryGitlabID,
					ConnectionStatus:     utils.ConnectionFailure,
					Error:                errors.New(msg),
				}
				return
			}

			log.WithFields(f).Debugf("connected to project/repo: %s", projectModel.PathWithNamespace)
			responseChannel <- responseChannelModel{
				RepositoryExternalID: repo.RepositoryGitlabID,
				ConnectionStatus:     utils.Connected,
				Error:                nil,
			}
		}(glClient, repo)
	}

	for i := 0; i < len(gitLabProjectRepos); i++ {
		resp := <-responseChannel
		// Find the matching repo for this response
		for _, repo := range gitLabProjectRepos {
			if repo.RepositoryGitlabID == resp.RepositoryExternalID {
				repo.ConnectionStatus = resp.ConnectionStatus
			}
		}
		if resp.Error != nil {
			log.WithFields(f).Warnf("problem connecting to GitLab repoistory, error: %+v", resp.Error)
		}
	}

	return gitLabProjectRepos
}
