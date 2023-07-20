// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package gitlab_activity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/communitybridge/easycla/cla-backend-go/config"

	"github.com/communitybridge/easycla/cla-backend-go/company"
	signatures1 "github.com/communitybridge/easycla/cla-backend-go/gen/v1/restapi/operations/signatures"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v1/models"
	v2Models "github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	gitlab_api "github.com/communitybridge/easycla/cla-backend-go/gitlab_api"
	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"
	"github.com/communitybridge/easycla/cla-backend-go/repositories"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	signatureModels "github.com/communitybridge/easycla/cla-backend-go/signatures/models"
	"github.com/communitybridge/easycla/cla-backend-go/users"
	"github.com/communitybridge/easycla/cla-backend-go/v2/common"
	gitlabModels "github.com/communitybridge/easycla/cla-backend-go/v2/gitlab-activity/models"
	"github.com/communitybridge/easycla/cla-backend-go/v2/gitlab_organizations"
	gitV2Repositories "github.com/communitybridge/easycla/cla-backend-go/v2/repositories"
	"github.com/communitybridge/easycla/cla-backend-go/v2/store"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var (
	missingID                 = errors.New("user missing in easyCLA records")
	missingCompanyAffiliation = errors.New("must confirm affiliation with their company")
	missingCompanyApproval    = errors.New("missing in company approval lists")
	secretTokenMismatch       = errors.New("secret token mismatch")
)

type Service interface {
	ProcessMergeCommentActivity(ctx context.Context, secretToken string, commentEvent *gitlab.MergeEvent) error
	ProcessMergeOpenedActivity(ctx context.Context, secretToken string, mergeEvent *gitlab.MergeEvent) error
	ProcessMergeActivity(ctx context.Context, secretToken string, input *gitlabModels.ProcessMergeActivityInput) error
	IsUserApprovedForSignature(ctx context.Context, f logrus.Fields, corporateSignature *models.Signature, user *models.User, gitlabUser *gitlab.User) bool
	HasUserSigned(ctx context.Context, claGroupID string, gitlabUser *gitlab.User) (bool, error)
}

type service struct {
	usersRepository             users.UserRepository
	gitlabOrgService            gitlab_organizations.ServiceInterface
	gitRepository               repositories.RepositoryInterface
	gitV2Repository             gitV2Repositories.RepositoryInterface
	signaturesRepository        signatures.SignatureRepository
	projectsCLAGroupsRepository projects_cla_groups.Repository
	companyRepository           company.IRepository
	signatureRepository         signatures.SignatureRepository
	gitLabApp                   *gitlab_api.App
	storeRepository             store.Repository
}

func NewService(gitRepository repositories.RepositoryInterface, gitV2Repository gitV2Repositories.RepositoryInterface, usersRepository users.UserRepository, signaturesRepository signatures.SignatureRepository, projectsCLAGroupsRepository projects_cla_groups.Repository,
	companyRepository company.IRepository, signatureRepository signatures.SignatureRepository, gitlabOrgService gitlab_organizations.ServiceInterface, storeRepository store.Repository) Service {
	return &service{
		gitRepository:               gitRepository,
		gitV2Repository:             gitV2Repository,
		usersRepository:             usersRepository,
		signaturesRepository:        signaturesRepository,
		projectsCLAGroupsRepository: projectsCLAGroupsRepository,
		companyRepository:           companyRepository,
		signatureRepository:         signatureRepository,
		gitLabApp:                   gitlab_api.Init(config.GetConfig().Gitlab.AppClientID, config.GetConfig().Gitlab.AppClientSecret, config.GetConfig().Gitlab.AppPrivateKey),
		gitlabOrgService:            gitlabOrgService,
		storeRepository:             storeRepository,
	}
}

func (s *service) ProcessMergeOpenedActivity(ctx context.Context, secretToken string, mergeEvent *gitlab.MergeEvent) error {
	projectName := mergeEvent.Project.Name
	projectPath := mergeEvent.Project.PathWithNamespace
	projectNamespace := mergeEvent.Project.Namespace
	projectID := mergeEvent.Project.ID
	mergeID := mergeEvent.ObjectAttributes.IID
	repositoryPath := mergeEvent.Project.PathWithNamespace
	lastCommitSha := mergeEvent.ObjectAttributes.LastCommit.ID

	input := &gitlabModels.ProcessMergeActivityInput{
		ProjectName:      projectName,
		ProjectPath:      projectPath,
		ProjectNamespace: projectNamespace,
		ProjectID:        projectID,
		MergeID:          mergeID,
		RepositoryPath:   repositoryPath,
		LastCommitSha:    lastCommitSha,
	}

	return s.ProcessMergeActivity(ctx, secretToken, input)

}

func (s *service) ProcessMergeCommentActivity(ctx context.Context, secretToken string, commentEvent *gitlab.MergeEvent) error {
	f := logrus.Fields{
		"functionName":      "ProcessMergeCommentActivity",
		utils.XREQUESTID:    ctx.Value(utils.XREQUESTID),
		"gitlabProjectPath": commentEvent.Project.PathWithNamespace,
		"projectID":         commentEvent.Project.ID,
		"projectPath":       commentEvent.Project.PathWithNamespace,
		"projectName":       commentEvent.Project.Name,
		"repositoryPath":    commentEvent.Project.PathWithNamespace,
		"commitSha":         commentEvent.ObjectAttributes.LastCommit.ID,
	}

	// Since we cant fetch the mergeID for comment event, we need to parse it from the URL
	urlPathList := strings.Split(commentEvent.ObjectAttributes.URL, "/")
	mergeID := strings.Split(urlPathList[len(urlPathList)-1], "#")[0]
	if mergeID == "" {
		return fmt.Errorf("merge ID not found in URL: %s", commentEvent.ObjectAttributes.URL)
	}
	mergeIDInt, err := strconv.Atoi(mergeID)
	if err != nil {
		return fmt.Errorf("unable to convert merge ID to int: %s, error: %v", mergeID, err)
	}

	f["mergeID"] = mergeIDInt

	projectName := commentEvent.Project.Name
	projectPath := commentEvent.Project.PathWithNamespace
	projectNamespace := commentEvent.Project.Namespace
	projectID := commentEvent.Project.ID
	repositoryPath := commentEvent.Project.PathWithNamespace
	authorUserName := commentEvent.User.Username
	authorEmail := commentEvent.User.Email

	input := &gitlabModels.ProcessMergeActivityInput{
		ProjectName:      projectName,
		ProjectPath:      projectPath,
		ProjectNamespace: projectNamespace,
		ProjectID:        projectID,
		MergeID:          mergeIDInt,
		RepositoryPath:   repositoryPath,
		LastCommitSha:    commentEvent.ObjectAttributes.LastCommit.ID,
		AuthorUserName:   authorUserName,
		AuthorEmail:      authorEmail,
	}

	return s.ProcessMergeActivity(ctx, secretToken, input)
}

func (s *service) ProcessMergeActivity(ctx context.Context, secretToken string, input *gitlabModels.ProcessMergeActivityInput) error {
	projectName := input.ProjectName
	projectPath := input.ProjectPath
	projectNamespace := input.ProjectNamespace
	projectID := input.ProjectID
	mergeID := input.MergeID
	repositoryPath := input.RepositoryPath
	lastCommitSha := input.LastCommitSha

	f := logrus.Fields{
		"functionName":           "ProcessMergeActivity",
		utils.XREQUESTID:         ctx.Value(utils.XREQUESTID),
		"gitlabProjectPath":      projectPath,
		"gitlabProjectName":      projectName,
		"gitlabProjectID":        projectID,
		"gitlabProjectNamespace": projectNamespace,
		"mergeID":                mergeID,
		"repositoryName":         repositoryPath,
	}

	log.WithFields(f).Debugf("looking up for gitlab org in easycla records ...")
	gitlabOrg, err := s.getGitlabOrganizationFromProjectPath(ctx, projectPath, projectNamespace)
	if err != nil {
		return fmt.Errorf("fetching internal gitlab org for following path : %s failed : %v", repositoryPath, err)
	}

	// log.WithFields(f).Debugf("checking gitlab org : %s auth state agains the webhook secret token", gitlabOrg.OrganizationName)
	// if gitlabOrg.AuthState != secretToken {
	// 	return secretTokenMismatch
	// }

	log.WithFields(f).Debugf("internal gitlab org : %s:%s is associated with external path : %s", gitlabOrg.OrganizationID, gitlabOrg.OrganizationName, repositoryPath)

	// fetch updated token info
	log.WithFields(f).Debugf("refreshing gitlab org : %s:%s auth info", gitlabOrg.OrganizationID, gitlabOrg.OrganizationName)
	oauthResponse, err := s.gitlabOrgService.RefreshGitLabOrganizationAuth(ctx, common.ToCommonModel(gitlabOrg))
	if err != nil {
		return fmt.Errorf("refreshing gitlab org auth info failed : %v", err)
	}

	gitlabClient, err := gitlab_api.NewGitlabOauthClient(*oauthResponse, s.gitLabApp)
	if err != nil {
		return fmt.Errorf("initializing gitlab client : %v", err)
	}

	if lastCommitSha == "" {
		log.WithFields(f).Debugf("loading GitLab merge request info for merge request: %d", mergeID)
		lastSha, err := gitlab_api.GetLatestCommit(gitlabClient, projectID, mergeID)
		if err != nil {
			return fmt.Errorf("fetching info for mr : %d and project : %d: %s, failed : %v", mergeID, projectID, projectName, err)
		}
		lastCommitSha = lastSha.ID
	}

	f["lastCommitSha"] = lastCommitSha
	log.WithFields(f).Debugf("last commit sha for merge request: %d is %s", mergeID, lastCommitSha)

	_, err = gitlab_api.FetchMrInfo(gitlabClient, projectID, mergeID)
	if err != nil {
		return fmt.Errorf("fetching info for mr : %d and project : %d: %s, failed : %v", mergeID, projectID, projectName, err)
	}

	// try to find the repository via the external id
	gitlabRepo, err := s.getGitlabRepoByName(ctx, repositoryPath)
	if err != nil {
		return fmt.Errorf("finding internal repository for gitlab org name failed : %v", err)
	}

	log.WithFields(f).Debugf("loading GitLab merge request participatants for merge request: %d", mergeID)
	participants, err := gitlab_api.FetchMrParticipants(gitlabClient, projectID, mergeID)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem loading GitLab merge request participants for merge request: %d", mergeID)
		return fmt.Errorf("problem loading GitLab merge request participants for merge request: %d - error: %+v", mergeID, err)
	}

	if len(participants) == 0 {
		return fmt.Errorf("no participants found in GitLab mr : %d, and gitlab project : %d", mergeID, projectID)
	}

	claGroup, err := s.projectsCLAGroupsRepository.GetClaGroupIDForProject(ctx, gitlabOrg.ProjectSfid)
	if err != nil {
		return fmt.Errorf("fetching claGroup id for gitlabOrg project sfid : %s, failed : %v", gitlabOrg.ProjectSfid, err)
	}
	claGroupID := claGroup.ClaGroupID
	log.WithFields(f).Debugf("gitlabOrg : %s is associated with cla group id : %s", gitlabOrg.OrganizationName, claGroupID)

	log.WithFields(f).Debugf("found %d participants for the MR ", len(participants))
	missingCLAMsg := "Missing CLA Authorization"
	signedCLAMsg := "EasyCLA check passed. You are authorized to contribute."

	var missingUsers []*utils.GatedGitlabUser
	var signedUsers []*gitlab.User
	for _, gitlabUser := range participants {
		log.WithFields(f).Debugf("checking if GitLab user: %s (%d) with email: %s has signed", gitlabUser.Username, gitlabUser.ID, gitlabUser.Email)
		userSigned, signedCheckErr := s.hasUserSigned(ctx, claGroupID, gitlabUser)
		if signedCheckErr != nil {
			log.WithFields(f).WithError(signedCheckErr).Warnf("problem checking if user : %s (%d) has signed - assuming not signed", gitlabUser.Username, gitlabUser.ID)
			missingUsers = append(missingUsers, &utils.GatedGitlabUser{
				User: gitlabUser,
				Err:  err,
			})
			continue
		}

		if userSigned {
			log.WithFields(f).Infof("gitlabUser: %s (%d) has signed", gitlabUser.Username, gitlabUser.ID)
			signedUsers = append(signedUsers, gitlabUser)
		} else {
			log.WithFields(f).Infof("gitlabUser: %s (%d) has NOT signed", gitlabUser.Username, gitlabUser.ID)
			missingUsers = append(missingUsers, &utils.GatedGitlabUser{
				User: gitlabUser,
				Err:  err,
			})
		}
	}

	err = s.setActiveMrMetadata(ctx, input.AuthorUserName, input.AuthorEmail, claGroupID, strconv.Itoa(int(gitlabRepo.RepositoryExternalID)), strconv.Itoa(mergeID), strconv.Itoa(projectID))
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem setting active merge request for merge id: %d", mergeID)
		return fmt.Errorf("setting commit status failed : %v", err)
	}
	signURL := utils.GetFullSignURL(gitlabOrg.OrganizationID, strconv.Itoa(int(gitlabRepo.RepositoryExternalID)), strconv.Itoa(mergeID))
	mrCommentContent := utils.PrepareMrCommentContent(missingUsers, signedUsers, signURL)
	if len(missingUsers) > 0 {
		log.WithFields(f).Errorf("merge request faild with 1 or more users not passing authorization - failed users : %+v", missingUsers)
		if statusErr := gitlab_api.SetCommitStatus(gitlabClient, projectID, lastCommitSha, gitlab.Failed, missingCLAMsg, signURL); statusErr != nil {
			log.WithFields(f).WithError(statusErr).Warnf("problem setting the commit status for merge request ID: %d, sha: %s", mergeID, lastCommitSha)
			return fmt.Errorf("setting commit status failed : %v", statusErr)
		}

		if mrCommentErr := gitlab_api.SetMrComment(gitlabClient, projectID, mergeID, mrCommentContent); mrCommentErr != nil {
			log.WithFields(f).WithError(mrCommentErr).Warnf("problem setting the commit merge request comment for merge request ID: %d", mergeID)
			return fmt.Errorf("setting comment failed : %v", mrCommentErr)
		}

		return nil
	}

	commitStatusErr := gitlab_api.SetCommitStatus(gitlabClient, projectID, lastCommitSha, gitlab.Success, signedCLAMsg, "")
	if commitStatusErr != nil {
		log.WithFields(f).WithError(commitStatusErr).Warnf("problem setting the commit status for merge request ID: %d, sha: %s", mergeID, lastCommitSha)
		return fmt.Errorf("setting commit status failed : %v", commitStatusErr)
	}

	if mrCommentErr := gitlab_api.SetMrComment(gitlabClient, projectID, mergeID, mrCommentContent); mrCommentErr != nil {
		log.WithFields(f).WithError(mrCommentErr).Warnf("problem setting the commit merge request comment for merge request ID: %d", mergeID)
		return fmt.Errorf("setting comment failed : %v", mrCommentErr)
	}

	return nil
}

// setActiveMrMetadata set active gitlab merge request metadata
func (s *service) setActiveMrMetadata(ctx context.Context, authorUserName string, authorEmail string, claGroupID string, repositoryID string, mrID string, projectID string) error {
	f := logrus.Fields{
		"functionName": "setActiveMrMetadata",
		"authorEmail":  authorEmail,
		"claGroupID":   claGroupID,
		"repositoryID": repositoryID,
		"mrID":         mrID,
		"projectID":    projectID,
	}

	type StoreValue struct {
		GitLabAuthorUserName string `json:"gitlab_author_username"`
		GitLabAuthorEmail    string `json:"gitlab_author_email"`
		ClaGroupID           string `json:"cla_group_id"`
		RepositoryID         string `json:"repository_id"`
		MergeRequestID       string `json:"merge_request_id"`
		ProjectID            string `json:"project_id"`
	}

	log.WithFields(f).Debugf("setting active merge request metadata")
	// set active signature metadata to track the user signing process

	storeValue := StoreValue{
		GitLabAuthorUserName: authorUserName,
		GitLabAuthorEmail:    authorEmail,
		ClaGroupID:           claGroupID,
		RepositoryID:         repositoryID,
		MergeRequestID:       mrID,
		ProjectID:            projectID,
	}

	json_data, err := json.Marshal(storeValue)
	if err != nil {
		msg := fmt.Sprintf("unable to marshall storeValue object: %+v", storeValue)
		log.WithFields(f).Warn(msg)
		return err
	}
	expire := time.Now().AddDate(0, 0, 7).Unix()
	log.WithFields(f).Debugf("setting expiry for active merge request data to : %d", expire)
	log.WithFields(f).Debugf("json data: %s", string(json_data))
	keyUName := fmt.Sprintf("active_mr:u:%s", authorUserName)
	err = s.storeRepository.SetActiveSignatureMetaData(ctx, keyUName, expire, string(json_data))
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to save merge request metadata")
		return err
	}
	if authorEmail != "" && authorEmail != "[REDACTED]" {
		keyEmail := fmt.Sprintf("active_mr:e:%s", authorEmail)
		err = s.storeRepository.SetActiveSignatureMetaData(ctx, keyEmail, expire, string(json_data))
		if err != nil {
			log.WithFields(f).WithError(err).Warn("unable to save merge request metadata")
			return err
		}
	}
	return nil
}

func (s *service) getGitlabOrganizationFromProjectPath(ctx context.Context, projectPath, projectNameSpace string) (*v2Models.GitlabOrganization, error) {
	parts := strings.Split(projectPath, "/")
	organizationName := parts[0]
	f := logrus.Fields{
		"functionName":     "getGitlabOrganizationFromProjectPath",
		"projectPath":      projectPath,
		"projectNameSpace": projectNameSpace,
		"organizationName": organizationName,
	}

	log.WithFields(f).Debug("getting gitlab org from project path")
	gitlabOrg, err := s.gitlabOrgService.GetGitLabOrganizationByFullPath(ctx, organizationName)
	if err != nil || gitlabOrg == nil {
		// try getting it with project name as well
		log.WithFields(f).Debugf("getting gitlab org with project name : %s", projectNameSpace)
		gitlabOrg, err = s.gitlabOrgService.GetGitLabOrganizationByFullPath(ctx, projectNameSpace)
		if err != nil || gitlabOrg == nil {
			return nil, fmt.Errorf("gitlab org : %s doesn't exist : %v", organizationName, err)
		}
	}

	gitlabOrg, err = s.gitlabOrgService.GetGitLabOrganization(ctx, gitlabOrg.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("fetching gitlab org : %s failed : %v", gitlabOrg.OrganizationID, err)
	}

	return gitlabOrg, nil
}

func (s *service) getGitlabRepoByName(ctx context.Context, repoNameWithPath string) (*models.GithubRepository, error) {
	gitlabRepo, err := s.gitV2Repository.GitLabGetRepositoryByName(ctx, repoNameWithPath)
	if err != nil || gitlabRepo == nil {
		return nil, fmt.Errorf("unable to locate GitLab repo for repoNameWithPath : %s, failed : %v", repoNameWithPath, err)
	}

	return gitlabRepo.ToGitHubModel(), nil
}

type UserSigned struct {
	signed bool
	err    error
}

func (s *service) HasUserSigned(ctx context.Context, claGroupID string, gitlabUser *gitlab.User) (bool, error) {
	return s.hasUserSigned(ctx, claGroupID, gitlabUser)
}

func (s *service) hasUserSigned(ctx context.Context, claGroupID string, gitlabUser *gitlab.User) (bool, error) {
	f := logrus.Fields{
		"functionName":    "v2.gitlab-activity.service.hasUserSigned",
		utils.XREQUESTID:  ctx.Value(utils.XREQUESTID),
		"gitlabUserID":    gitlabUser.ID,
		"gitlabUserName":  gitlabUser.Username,
		"gitlabUserEmail": gitlabUser.Email,
	}

	userModels, lookUpErr := s.findUserModelForGitlabUser(f, gitlabUser)
	if lookUpErr != nil {
		log.WithFields(f).WithError(lookUpErr).Warnf("unable to find user model for gitlab user: %v", gitlabUser)
		return false, lookUpErr
	}

	if len(userModels) == 0 {
		log.WithFields(f).Warnf("gitlab user: %s (%d) not found in easycla records", gitlabUser.Username, gitlabUser.ID)
		return false, missingID
	}

	for _, userModel := range userModels {
		signed, err := s.isSigned(ctx, userModel, claGroupID, gitlabUser)
		if err != nil {
			log.WithFields(f).Debugf("error checking if user is signed, error: %v", err)
			continue
		}
		if signed {
			log.WithFields(f).Debugf("found signed user for clagroupID: %s, userID: %s", claGroupID, userModel.UserID)
			return true, nil
		} else {
			log.WithFields(f).Debugf("user is not signed for claGroupID: %s, userID: %s", claGroupID, userModel.UserID)
		}
	}

	return false, nil
}

func (s *service) isSigned(ctx context.Context, userModel *models.User, claGroupID string, gitlabUser *gitlab.User) (bool, error) {
	f := logrus.Fields{
		"functionName":    "v2.gitlab-activity.service.isSigned",
		utils.XREQUESTID:  ctx.Value(utils.XREQUESTID),
		"gitlabUserID":    gitlabUser.ID,
		"gitlabUserName":  gitlabUser.Username,
		"gitlabUserEmail": gitlabUser.Email,
	}

	// First check for an ICLA signature
	icla, err := s.signaturesRepository.GetIndividualSignature(ctx, claGroupID, userModel.UserID, aws.Bool(true), aws.Bool(true))
	if err != nil {
		log.WithFields(f).Warnf("fetching ICLA for gitlab user : %d:%s failed : %v", gitlabUser.ID, gitlabUser.Username, err)
		return false, err
	}

	if icla != nil {
		log.WithFields(f).Infof("user has signed the following signature (ICLA): %s, passing", icla.SignatureID)
		return true, nil
	}

	if userModel.CompanyID == "" {
		log.WithFields(f).Debugf("user does not have association with any company, can't confirm employee acknoledgement")
		return false, fmt.Errorf("user hasn't signed yet")
	}

	companyID := userModel.CompanyID
	_, err = s.companyRepository.GetCompany(ctx, companyID)
	if err != nil {
		msg := fmt.Sprintf("can't load company record: %s for user: %s (%s), error: %v", companyID, userModel.Username, userModel.UserID, err)
		log.WithFields(f).Errorf(msg)
		return false, fmt.Errorf(msg)
	}

	corporateSignature, err := s.signatureRepository.GetCorporateSignature(ctx, claGroupID, companyID, aws.Bool(true), aws.Bool(true))
	if err != nil {
		msg := fmt.Sprintf("can't load company signature record for company: %s for user : %s (%s), error : %v", companyID, userModel.Username, userModel.UserID, err)
		log.WithFields(f).Errorf(msg)
		return false, fmt.Errorf(msg)
	}

	if corporateSignature == nil {
		msg := fmt.Sprintf("no corporate signature (CCLA) record found for company : %s ", companyID)
		log.WithFields(f).Errorf(msg)
		return false, fmt.Errorf(msg)
	}

	log.WithFields(f).Debugf("loaded corporate signature id: %s for claGroupID: %s and companyID: %s", corporateSignature.SignatureID, claGroupID, companyID)

	approvalCriteria := &signatureModels.ApprovalCriteria{}
	if gitlabUser.Email != "" {
		approvalCriteria.UserEmail = gitlabUser.Email
	} else if gitlabUser.Username != "" {
		approvalCriteria.GitlabUsername = gitlabUser.Username
	} else {
		msg := fmt.Sprintf("gitlabUser model doesn't have enough information to fetch the employee signatures for user : %s", userModel.UserID)
		log.WithFields(f).Errorf(msg)
		return false, fmt.Errorf(msg)
	}

	if !s.IsUserApprovedForSignature(ctx, f, corporateSignature, userModel, gitlabUser) {
		log.WithFields(f).Debugf("user is not approved in signature : %s", corporateSignature.SignatureID)
		return false, fmt.Errorf("user is not approved in signature : %s", corporateSignature.SignatureID)
	}

	employeeSignatures, err := s.signaturesRepository.GetProjectCompanyEmployeeSignatures(ctx, signatures1.GetProjectCompanyEmployeeSignaturesParams{
		CompanyID: companyID,
		ProjectID: claGroupID,
		PageSize:  utils.Int64(100),
	}, approvalCriteria)

	if err != nil {
		msg := fmt.Sprintf("can't load employee signature records : %s for user : %s association : %v", companyID, userModel.UserID, err)
		log.WithFields(f).Errorf(msg)
		return false, fmt.Errorf(msg)
	}

	if len(employeeSignatures.Signatures) == 0 {
		msg := fmt.Sprintf("no employee signature records found for company : %s user : %s association", companyID, userModel.UserID)
		log.WithFields(f).Errorf(msg)
		return false, fmt.Errorf(msg)
	}

	log.WithFields(f).Warnf("is in signature approval list : %s and has employee signature", corporateSignature.SignatureID)
	return true, nil
}

// findUserModelForGitlabUser locates the user model in our users table for the given GitLab user (by GitLab ID, GitLab username, or email)
func (s *service) findUserModelForGitlabUser(f logrus.Fields, gitlabUser *gitlab.User) ([]*models.User, error) {

	if gitlabUser.ID != 0 {
		log.WithFields(f).Debugf("Looking up GitLab user via ID: %d", gitlabUser.ID)
		userModel, lookupErr := s.usersRepository.GetUserByGitlabID(gitlabUser.ID)
		if lookupErr != nil {
			log.WithFields(f).WithError(lookupErr).Warnf("problem locating GitLab user via GitLab ID : %d", gitlabUser.ID)
		} else if userModel != nil {
			log.WithFields(f).Debugf("located GitLab user via ID: %d", gitlabUser.ID)
			return []*models.User{userModel}, nil
		}
	}

	if gitlabUser.Username != "" {
		log.WithFields(f).Debugf("Looking up GitLab user via username: %s", gitlabUser.Username)
		userModel, lookupErr := s.usersRepository.GetUserByGitLabUsername(gitlabUser.Username)
		if lookupErr != nil {
			log.WithFields(f).WithError(lookupErr).Warnf("problem locating GitLab user via GitLab username : %s", gitlabUser.Username)
		} else if userModel != nil {
			log.WithFields(f).Debugf("located GitLab user via username: %s", gitlabUser.Username)
			return []*models.User{userModel}, nil
		}
	}

	if gitlabUser.Email != "" {
		gitlabUsers := make([]*models.User, 0)
		log.WithFields(f).Debugf("Looking up GitLab user via user email: %s", gitlabUser.Email)
		// previously search was done by lf_email, now we are searching by email #3816
		users, lookupErr := s.usersRepository.GetUsersByEmail(gitlabUser.Email)
		if lookupErr != nil {
			log.WithFields(f).WithError(lookupErr).Warnf("problem locating GitLab user via GitLab username : %s", gitlabUser.Username)
		} else if len(users) > 0 {
			log.WithFields(f).Debugf("located GitLab user via email: %s", gitlabUser.Email)
			gitlabUsers = append(gitlabUsers, users...)
			return gitlabUsers, nil
		}
	}

	// Didn't find it
	return nil, nil
}

func (s *service) IsUserApprovedForSignature(ctx context.Context, f logrus.Fields, corporateSignature *models.Signature, user *models.User, gitlabUser *gitlab.User) bool {
	log.WithFields(f).Debugf("checking if user : %s is approved for corporate signature : %s", user.UserID, corporateSignature.SignatureID)
	userEmails := user.Emails
	if string(user.LfEmail) != "" {
		userEmails = append(userEmails, string(user.LfEmail))
	}

	emailApprovalList := corporateSignature.EmailApprovalList
	domainApprovalList := corporateSignature.DomainApprovalList
	log.WithFields(f).Debugf("checking if user : %s is approved for corporate signature : %s, email approval list : %+v", user.UserID, corporateSignature.SignatureID, emailApprovalList)

	if len(userEmails) > 0 && len(emailApprovalList) > 0 {
		for _, email := range userEmails {
			for _, approvalEmail := range emailApprovalList {
				if email == approvalEmail {
					log.WithFields(f).Debugf("found user email : %s in email approval list ", email)
					return true
				}
			}
		}
	} else {
		log.WithFields(f).Warnf("no match for user in signature email approval list")
	}

	if len(domainApprovalList) > 0 && len(userEmails) > 0 {
		log.WithFields(f).Debugf("checking if emails : %+v are approved for corporate signature : %s, domain approval list : %+v", userEmails, corporateSignature.SignatureID, domainApprovalList)
		for _, userEmail := range userEmails {
			for _, domainApprovalPattern := range domainApprovalList {
				if strings.HasPrefix(domainApprovalPattern, "*.") {
					domainApprovalPattern = strings.Replace(domainApprovalPattern, "*.", ".*", 1)
				} else if strings.HasPrefix(domainApprovalPattern, "*") {
					domainApprovalPattern = strings.Replace(domainApprovalPattern, "*", ".*", 1)
				} else if strings.HasPrefix(domainApprovalPattern, ".") {
					domainApprovalPattern = strings.Replace(domainApprovalPattern, ".", ".*", 1)
				}
				regexpApprovalPattern := "^.*@" + domainApprovalPattern + "$"
				if ok, err := regexp.MatchString(regexpApprovalPattern, userEmail); ok && err == nil {
					log.WithFields(f).Debugf("found user email : %s in email approval list : %s", userEmail, domainApprovalPattern)
					return true
				}
			}
		}
	}

	gitlabUserName := gitlabUser.Username
	gitlabUsernameApprovalList := corporateSignature.GitlabUsernameApprovalList
	if gitlabUserName != "" && len(gitlabUsernameApprovalList) > 0 {
		log.WithFields(f).Debugf("checking gitlab username : %s for gitlab approval list : %+v", gitlabUserName, gitlabUsernameApprovalList)
		for _, gitlabApproval := range gitlabUsernameApprovalList {
			if gitlabApproval == gitlabUserName {
				log.WithFields(f).Debugf("found gitlab username : %s in gitlab approval list ", gitlabUserName)
				return true
			}
		}

	} else {
		log.WithFields(f).Warnf("no match found for gitlabUser : %s in gitlab approval list : %+v", gitlabUserName, gitlabUsernameApprovalList)
	}

	gitlabGroupApprovalList := corporateSignature.GitlabOrgApprovalList
	if gitlabUserName != "" && len(gitlabGroupApprovalList) > 0 {
		log.WithFields(f).Debugf("checking gitlab username : %s for gitlab org approval list : %+v ", gitlabUserName, gitlabGroupApprovalList)

		for _, gitlabGroupApproval := range gitlabGroupApprovalList {
			isApproved, err := s.checkGitLabGroupApproval(ctx, gitlabUserName, gitlabGroupApproval)
			if err != nil {
				log.WithFields(f).WithError(err).Warn("unable to get username")
				break
			}
			if isApproved == true {
				log.WithFields(f).Debugf(" found gitlab username : %s in gitlab org approval list : %+v", gitlabUserName, gitlabGroupApprovalList)
				return true
			}
		}
	}

	log.WithFields(f).Errorf("unable to find user in any approval list")
	return false

}

/**
 * Parses url with the given regular expression and returns the
 * group values defined in the expression.
 *
 */
func getParams(regEx, url string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

func (s *service) checkGitLabGroupApproval(ctx context.Context, userName, URL string) (bool, error) {
	f := logrus.Fields{
		"functionName": "checkGitLabGroupApproval",
		"userName":     userName,
		"group_url":    URL,
	}

	log.WithFields(f).Debugf("checking approval list gitlab org criteria : %s for user: %s ", URL, userName)
	var searchURL = URL
	params := getParams(`(?P<base>\bhttps://gitlab.com/\b)(?P<group>\bgroups\/\b)?(?P<name>\w+)`, URL)
	if params[`group`] == "" {
		params[`group`] = "groups/"
		updated := fmt.Sprintf("%s%s%s", params[`base`], params[`group`], params[`name`])
		log.WithFields(f).Debugf("updating url : %s to %s for easycla search purporses ", searchURL, updated)
		searchURL = updated
	}
	gitlabOrg, _ := s.gitlabOrgService.GetGitLabOrganizationByURL(ctx, searchURL)
	if gitlabOrg != nil {
		oauthResponse, err := s.gitlabOrgService.RefreshGitLabOrganizationAuth(ctx, common.ToCommonModel(gitlabOrg))
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("problem refreshing gitlab auth for org: %s ", gitlabOrg.OrganizationName)
			return false, err
		}

		gitlabClient, clientErr := gitlab_api.NewGitlabOauthClient(*oauthResponse, s.gitLabApp)
		if clientErr != nil {
			log.WithFields(f).WithError(clientErr).Warnf("problem getting gitLabClient for org: %s ", gitlabOrg.OrganizationName)
			return false, clientErr
		}
		members, err := gitlab_api.ListGroupMembers(ctx, gitlabClient, int(gitlabOrg.OrganizationExternalID))
		if err != nil {
			log.WithFields(f).WithError(err).Warn("problem getting gitlab group members")
			return false, err
		}
		for _, member := range members {
			if userName == member.Username {
				log.WithFields(f).Debugf("%s is a member of group: %s ", userName, URL)
				return true, nil
			}
		}
	}

	return false, nil
}
