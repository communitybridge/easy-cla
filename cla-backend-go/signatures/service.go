// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package signatures

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/events"

	"github.com/communitybridge/easycla/cla-backend-go/users"

	"github.com/LF-Engineering/lfx-kit/auth"
	"github.com/communitybridge/easycla/cla-backend-go/company"
	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/signatures"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	githubpkg "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// SignatureService interface
type SignatureService interface {
	GetSignature(signatureID string) (*models.Signature, error)
	GetIndividualSignature(claGroupID, userID string) (*models.Signature, error)
	GetCorporateSignature(claGroupID, companyID string) (*models.Signature, error)
	GetProjectSignatures(params signatures.GetProjectSignaturesParams) (*models.Signatures, error)
	GetProjectCompanySignature(companyID, projectID string, signed, approved *bool, nextKey *string, pageSize *int64) (*models.Signature, error)
	GetProjectCompanySignatures(params signatures.GetProjectCompanySignaturesParams) (*models.Signatures, error)
	GetProjectCompanyEmployeeSignatures(params signatures.GetProjectCompanyEmployeeSignaturesParams) (*models.Signatures, error)
	GetCompanySignatures(params signatures.GetCompanySignaturesParams) (*models.Signatures, error)
	GetCompanyIDsWithSignedCorporateSignatures(claGroupID string) ([]SignatureCompanyID, error)
	GetUserSignatures(params signatures.GetUserSignaturesParams) (*models.Signatures, error)
	InvalidateProjectRecords(projectID string, projectName string) (int, error)

	GetGithubOrganizationsFromWhitelist(signatureID string, githubAccessToken string) ([]models.GithubOrg, error)
	AddGithubOrganizationToWhitelist(signatureID string, whiteListParams models.GhOrgWhitelist, githubAccessToken string) ([]models.GithubOrg, error)
	DeleteGithubOrganizationFromWhitelist(signatureID string, whiteListParams models.GhOrgWhitelist, githubAccessToken string) ([]models.GithubOrg, error)
	UpdateApprovalList(authUser *auth.User, projectModel *models.Project, companyModel *models.Company, claGroupID string, params *models.ApprovalList) (*models.Signature, error)

	AddCLAManager(signatureID, claManagerID string) (*models.Signature, error)
	RemoveCLAManager(signatureID, claManagerID string) (*models.Signature, error)

	GetClaGroupICLASignatures(claGroupID string, searchTerm *string) (*models.IclaSignatures, error)
	GetClaGroupCorporateContributors(claGroupID string, companyID *string, searchTerm *string) (*models.CorporateContributorList, error)
}

type service struct {
	repo                SignatureRepository
	companyService      company.IService
	usersService        users.Service
	eventsService       events.Service
	githubOrgValidation bool
}

// NewService creates a new whitelist service
func NewService(repo SignatureRepository, companyService company.IService, usersService users.Service, eventsService events.Service, githubOrgValidation bool) SignatureService {
	return service{
		repo,
		companyService,
		usersService,
		eventsService,
		githubOrgValidation,
	}
}

// GetSignature returns the signature associated with the specified signature ID
func (s service) GetSignature(signatureID string) (*models.Signature, error) {
	return s.repo.GetSignature(signatureID)
}

// GetIndividualSignature returns the signature associated with the specified CLA Group and User ID
func (s service) GetIndividualSignature(claGroupID, userID string) (*models.Signature, error) {
	return s.repo.GetIndividualSignature(claGroupID, userID)
}

// GetCorporateSignature returns the signature associated with the specified CLA Group and Company ID
func (s service) GetCorporateSignature(claGroupID, companyID string) (*models.Signature, error) {
	return s.repo.GetCorporateSignature(claGroupID, companyID)
}

// GetProjectSignatures returns the list of signatures associated with the specified project
func (s service) GetProjectSignatures(params signatures.GetProjectSignaturesParams) (*models.Signatures, error) {

	const defaultPageSize int64 = 10
	var pageSize = defaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	projectSignatures, err := s.repo.GetProjectSignatures(params, pageSize)
	if err != nil {
		return nil, err
	}

	return projectSignatures, nil
}

// GetProjectCompanySignature returns the signature associated with the specified project and company
func (s service) GetProjectCompanySignature(companyID, projectID string, signed, approved *bool, nextKey *string, pageSize *int64) (*models.Signature, error) {
	return s.repo.GetProjectCompanySignature(companyID, projectID, signed, approved, nextKey, pageSize)
}

// GetProjectCompanySignatures returns the list of signatures associated with the specified project
func (s service) GetProjectCompanySignatures(params signatures.GetProjectCompanySignaturesParams) (*models.Signatures, error) {

	const defaultPageSize int64 = 10
	var pageSize = defaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	signed := true
	approved := true

	projectSignatures, err := s.repo.GetProjectCompanySignatures(
		params.CompanyID, params.ProjectID, &signed, &approved, params.NextKey, &pageSize)
	if err != nil {
		return nil, err
	}

	return projectSignatures, nil
}

// GetProjectCompanyEmployeeSignatures returns the list of employee signatures associated with the specified project
func (s service) GetProjectCompanyEmployeeSignatures(params signatures.GetProjectCompanyEmployeeSignaturesParams) (*models.Signatures, error) {

	const defaultPageSize int64 = 10
	var pageSize = defaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	projectSignatures, err := s.repo.GetProjectCompanyEmployeeSignatures(params, pageSize)
	if err != nil {
		return nil, err
	}

	return projectSignatures, nil
}

// GetCompanySignatures returns the list of signatures associated with the specified company
func (s service) GetCompanySignatures(params signatures.GetCompanySignaturesParams) (*models.Signatures, error) {

	const defaultPageSize int64 = 50
	var pageSize = defaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	companySignatures, err := s.repo.GetCompanySignatures(params, pageSize, LoadACLDetails)
	if err != nil {
		return nil, err
	}

	return companySignatures, nil
}

// GetCompanyIDsWithSignedCorporateSignatures returns a list of company IDs that have signed a CLA agreement
func (s service) GetCompanyIDsWithSignedCorporateSignatures(claGroupID string) ([]SignatureCompanyID, error) {
	return s.repo.GetCompanyIDsWithSignedCorporateSignatures(claGroupID)
}

// GetUserSignatures returns the list of user signatures associated with the specified user
func (s service) GetUserSignatures(params signatures.GetUserSignaturesParams) (*models.Signatures, error) {

	const defaultPageSize int64 = 10
	var pageSize = defaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	userSignatures, err := s.repo.GetUserSignatures(params, pageSize)
	if err != nil {
		return nil, err
	}

	return userSignatures, nil
}

// GetGithubOrganizationsFromWhitelist retrieves the organization from the whitelist
func (s service) GetGithubOrganizationsFromWhitelist(signatureID string, githubAccessToken string) ([]models.GithubOrg, error) {

	if signatureID == "" {
		msg := "unable to get GitHub organizations whitelist - signature ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	orgIds, err := s.repo.GetGithubOrganizationsFromWhitelist(signatureID)
	if err != nil {
		log.Warnf("error loading github organization from whitelist using signatureID: %s, error: %v",
			signatureID, err)
		return nil, err
	}

	if githubAccessToken != "" {
		log.Debugf("already authenticated with github - scanning for user's orgs...")

		selectedOrgs := make(map[string]struct{}, len(orgIds))
		for _, selectedOrg := range orgIds {
			selectedOrgs[*selectedOrg.ID] = struct{}{}
		}

		// Since we're logged into github, lets get the list of organization we can add.
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubAccessToken},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client := githubpkg.NewClient(tc)

		opt := &githubpkg.ListOptions{
			PerPage: 100,
		}

		orgs, _, err := client.Organizations.List(context.Background(), "", opt)
		if err != nil {
			return nil, err
		}

		for _, org := range orgs {
			_, ok := selectedOrgs[*org.Login]
			if ok {
				continue
			}

			orgIds = append(orgIds, models.GithubOrg{ID: org.Login})
		}
	}

	return orgIds, nil
}

// AddGithubOrganizationToWhitelist adds the GH organization to the whitelist
func (s service) AddGithubOrganizationToWhitelist(signatureID string, whiteListParams models.GhOrgWhitelist, githubAccessToken string) ([]models.GithubOrg, error) {
	organizationID := whiteListParams.OrganizationID

	if signatureID == "" {
		msg := "unable to add GitHub organization from whitelist - signature ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	if organizationID == nil {
		msg := "unable to add GitHub organization from whitelist - organization ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	// GH_ORG_VALIDATION environment - set to false to test locally which will by-pass the GH auth checks and
	// allow functional tests (e.g. with curl or postmon) - default is enabled

	if s.githubOrgValidation {
		// Verify the authenticated github user has access to the github organization being added.
		if githubAccessToken == "" {
			msg := fmt.Sprintf("unable to add github organization, not logged in using "+
				"signatureID: %s, github organization id: %s",
				signatureID, *organizationID)
			log.Warn(msg)
			return nil, errors.New(msg)
		}

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubAccessToken},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client := githubpkg.NewClient(tc)

		opt := &githubpkg.ListOptions{
			PerPage: 100,
		}

		log.Debugf("querying for user's github organizations...")
		orgs, _, err := client.Organizations.List(context.Background(), "", opt)
		if err != nil {
			return nil, err
		}

		found := false
		for _, org := range orgs {
			if *org.Login == *organizationID {
				found = true
				break
			}
		}

		if !found {
			msg := fmt.Sprintf("user is not authorized for github organization id: %s", *organizationID)
			log.Warnf(msg)
			return nil, errors.New(msg)
		}
	}

	gitHubWhiteList, err := s.repo.AddGithubOrganizationToWhitelist(signatureID, *organizationID)
	if err != nil {
		log.Warnf("issue adding github organization to white list using signatureID: %s, gh org id: %s, error: %v",
			signatureID, *organizationID, err)
		return nil, err
	}

	return gitHubWhiteList, nil
}

// DeleteGithubOrganizationFromWhitelist deletes the specified GH organization from the whitelist
func (s service) DeleteGithubOrganizationFromWhitelist(signatureID string, whiteListParams models.GhOrgWhitelist, githubAccessToken string) ([]models.GithubOrg, error) {

	// Extract the payload values
	organizationID := whiteListParams.OrganizationID

	if signatureID == "" {
		msg := "unable to delete GitHub organization from whitelist - signature ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	if organizationID == nil {
		msg := "unable to delete GitHub organization from whitelist - organization ID is nil"
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	// GH_ORG_VALIDATION environment - set to false to test locally which will by-pass the GH auth checks and
	// allow functional tests (e.g. with curl or postmon) - default is enabled

	if s.githubOrgValidation {
		// Verify the authenticated github user has access to the github organization being added.
		if githubAccessToken == "" {
			msg := fmt.Sprintf("unable to delete github organization, not logged in using "+
				"signatureID: %s, github organization id: %s",
				signatureID, *organizationID)
			log.Warn(msg)
			return nil, errors.New(msg)
		}

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubAccessToken},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client := githubpkg.NewClient(tc)

		opt := &githubpkg.ListOptions{
			PerPage: 100,
		}

		log.Debugf("querying for user's github organizations...")
		orgs, _, err := client.Organizations.List(context.Background(), "", opt)
		if err != nil {
			return nil, err
		}

		found := false
		for _, org := range orgs {
			if *org.Login == *organizationID {
				found = true
				break
			}
		}

		if !found {
			msg := fmt.Sprintf("user is not authorized for github organization id: %s", *organizationID)
			log.Warnf(msg)
			return nil, errors.New(msg)
		}
	}

	gitHubWhiteList, err := s.repo.DeleteGithubOrganizationFromWhitelist(signatureID, *organizationID)
	if err != nil {
		return nil, err
	}

	return gitHubWhiteList, nil
}

// UpdateApprovalList service method
func (s service) UpdateApprovalList(authUser *auth.User, projectModel *models.Project, companyModel *models.Company, claGroupID string, params *models.ApprovalList) (*models.Signature, error) {
	pageSize := int64(1)
	signed, approved := true, true
	sigModel, sigErr := s.GetProjectCompanySignature(companyModel.CompanyID, claGroupID, &signed, &approved, nil, &pageSize)
	if sigErr != nil {
		msg := fmt.Sprintf("unable to locate project company signature by Company ID: %s, Project ID: %s, CLA Group ID: %s, error: %+v",
			companyModel.CompanyID, projectModel.ProjectID, claGroupID, sigErr)
		log.Warn(msg)
		return nil, NewBadRequestError(msg)
	}
	if sigModel == nil {
		msg := fmt.Sprintf("unable to locate signature for company ID: %s CLA Group ID: %s, type: ccla, signed: %t, approved: %t",
			companyModel.CompanyID, claGroupID, signed, approved)
		log.Warn(msg)
		return nil, NewBadRequestError(msg)
	}

	// Ensure current user is in the Signature ACL
	claManagers := sigModel.SignatureACL
	if !utils.CurrentUserInACL(authUser, claManagers) {
		msg := fmt.Sprintf("EasyCLA - 403 Forbidden - CLA Manager %s / %s is not authorized to approve request for company ID: %s / %s / %s, project ID: %s / %s / %s",
			authUser.UserName, authUser.Email,
			companyModel.CompanyName, companyModel.CompanyExternalID, companyModel.CompanyID,
			projectModel.ProjectName, projectModel.ProjectExternalID, projectModel.ProjectID)
		return nil, NewForbiddenError(msg)
	}

	// Lookup the user making the request
	userModel, userErr := s.usersService.GetUserByUserName(authUser.UserName, true)
	if userErr != nil {
		return nil, userErr
	}

	updatedSig, err := s.repo.UpdateApprovalList(projectModel.ProjectID, companyModel.CompanyID, params)
	if err != nil {
		return updatedSig, err
	}

	// Log Events
	s.createEventLogEntries(companyModel, projectModel, userModel, params)

	// Send an email to the CLA Managers
	for _, claManager := range claManagers {
		claManagerEmail := getBestEmail(claManager)
		s.sendApprovalListUpdateEmailToCLAManagers(companyModel, projectModel, claManager.Username, claManagerEmail, params)
	}

	// Send emails to contributors if email or GH username as added/removed
	s.sendRequestAccessEmailToContributors(authUser, companyModel, projectModel, params)

	return updatedSig, nil
}

// Disassociate project signatures
func (s service) InvalidateProjectRecords(projectID string, projectName string) (int, error) {
	f := logrus.Fields{
		"functionName": "InvalidateProjectRecords",
		"projectID":    projectID,
		"projectName":  projectName}

	result, err := s.repo.ProjectSignatures(projectID)
	if err != nil {
		log.WithFields(f).Warnf(fmt.Sprintf("Unable to get signatures for project: %s", projectID))
		return 0, err
	}

	if len(result.Signatures) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(result.Signatures))
		log.WithFields(f).Debugf(fmt.Sprintf("Invalidating %d signatures for project: %s ",
			len(result.Signatures), projectID))
		for _, signature := range result.Signatures {
			// Do this in parallel, as we could have a lot to invalidate
			go func(sigID, projName string) {
				defer wg.Done()
				updateErr := s.repo.InvalidateProjectRecord(sigID, projName)
				if updateErr != nil {
					log.WithFields(f).Warnf("Unable to update signature: %s with project name: %s, error: %v",
						sigID, projName, updateErr)
				}
			}(signature.SignatureID, projectName)
		}

		// Wait until all the workers are done
		wg.Wait()
	}

	return len(result.Signatures), nil
}

// AddCLAManager adds the specified manager to the signature ACL list
func (s service) AddCLAManager(signatureID, claManagerID string) (*models.Signature, error) {
	return s.repo.AddCLAManager(signatureID, claManagerID)
}

// RemoveCLAManager removes the specified manager from the signature ACL list
func (s service) RemoveCLAManager(signatureID, claManagerID string) (*models.Signature, error) {
	return s.repo.RemoveCLAManager(signatureID, claManagerID)
}

// appendList is a helper function to generate the email content of the Approval List changes
func appendList(approvalList []string, message string) string {
	approvalListSummary := ""

	if len(approvalList) > 0 {
		for _, value := range approvalList {
			approvalListSummary += fmt.Sprintf("<li>%s %s</li>", message, value)
		}
	}

	return approvalListSummary
}

// buildApprovalListSummary is a helper function to generate the email content of the Approval List changes
func buildApprovalListSummary(approvalListChanges *models.ApprovalList) string {
	approvalListSummary := "<ul>"
	approvalListSummary += appendList(approvalListChanges.AddEmailApprovalList, "Added Email:")
	approvalListSummary += appendList(approvalListChanges.RemoveEmailApprovalList, "Removed Email:")
	approvalListSummary += appendList(approvalListChanges.AddDomainApprovalList, "Added Domain:")
	approvalListSummary += appendList(approvalListChanges.RemoveDomainApprovalList, "Removed Domain:")
	approvalListSummary += appendList(approvalListChanges.AddGithubUsernameApprovalList, "Added GithHub User:")
	approvalListSummary += appendList(approvalListChanges.RemoveGithubUsernameApprovalList, "Removed GitHub User:")
	approvalListSummary += appendList(approvalListChanges.AddGithubOrgApprovalList, "Added GithHub Organization:")
	approvalListSummary += appendList(approvalListChanges.RemoveGithubOrgApprovalList, "Removed GitHub Organization:")
	approvalListSummary += "</ul>"
	return approvalListSummary
}

// sendRequestAccessEmailToCLAManagers sends the request access email to the specified CLA Managers
func (s service) sendApprovalListUpdateEmailToCLAManagers(companyModel *models.Company, projectModel *models.Project, recipientName, recipientAddress string, approvalListChanges *models.ApprovalList) {
	f := logrus.Fields{
		"function":          "sendApprovalListUpdateEmailToCLAManagers",
		"projectName":       projectModel.ProjectName,
		"projectExternalID": projectModel.ProjectExternalID,
		"foundationSFID":    projectModel.FoundationSFID,
		"companyName":       companyModel.CompanyName,
		"companyExternalID": companyModel.CompanyExternalID,
		"recipientName":     recipientName,
		"recipientAddress":  recipientAddress}

	companyName := companyModel.CompanyName
	projectName := projectModel.ProjectName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: Approval List Update for %s on %s", companyName, projectName)
	recipients := []string{recipientAddress}
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the project %s.</p>
<p>The EasyCLA approval list for %s for project %s was modified.</p>
<p>The modification was as follows:</p>
%s
<p>Contributors with previously failed pull requests to %s can close and re-open the pull request to force a recheck by
the EasyCLA system.</p>
%s
%s`,
		recipientName, projectName, companyName, projectName, buildApprovalListSummary(approvalListChanges), projectName,
		utils.GetEmailHelpContent(projectModel.Version == utils.V2), utils.GetEmailSignOffContent())

	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.WithFields(f).Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.WithFields(f).Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

// getAddEmailContributors is a helper function to lookup the contributors impacted by the Approval List update
func (s service) getAddEmailContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.AddEmailApprovalList {
		userModel, err := s.usersService.GetUserByEmail(value)
		if err != nil {
			log.Warnf("unable to lookup user by LF email: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}

// getRemoveEmailContributors is a helper function to lookup the contributors impacted by the Approval List update
func (s service) getRemoveEmailContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.RemoveEmailApprovalList {
		userModel, err := s.usersService.GetUserByEmail(value)
		if err != nil {
			log.Warnf("unable to lookup user by LF email: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}

// getAddGitHubContributors is a helper function to lookup the contributors impacted by the Approval List update
func (s service) getAddGitHubContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.AddGithubUsernameApprovalList {
		userModel, err := s.usersService.GetUserByGitHubUsername(value)
		if err != nil {
			log.Warnf("unable to lookup user by GitHub username: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}

// getRemoveGitHubContributors is a helper function to lookup the contributors impacted by the Approval List update
func (s service) getRemoveGitHubContributors(approvalList *models.ApprovalList) []*models.User {
	var userModelList []*models.User
	for _, value := range approvalList.RemoveGithubUsernameApprovalList {
		userModel, err := s.usersService.GetUserByGitHubUsername(value)
		if err != nil {
			log.Warnf("unable to lookup user by GitHub username: %s, error: %+v", value, err)
		} else {
			userModelList = append(userModelList, userModel)
		}
	}

	return userModelList
}
func (s service) sendRequestAccessEmailToContributors(authUser *auth.User, companyModel *models.Company, projectModel *models.Project, approvalList *models.ApprovalList) {
	addEmailUsers := s.getAddEmailContributors(approvalList)
	for _, user := range addEmailUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, projectModel, user.Username, user.LfEmail, "added", "to", "you are authorized to contribute to")
	}
	removeEmailUsers := s.getRemoveEmailContributors(approvalList)
	for _, user := range removeEmailUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, projectModel, user.Username, user.LfEmail, "removed", "from", "you are no longer authorized to contribute to")
	}
	addGitHubUsers := s.getAddGitHubContributors(approvalList)
	for _, user := range addGitHubUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, projectModel, user.Username, user.LfEmail, "added", "to", "you are authorized to contribute to")
	}
	removeGitHubUsers := s.getRemoveGitHubContributors(approvalList)
	for _, user := range removeGitHubUsers {
		sendRequestAccessEmailToContributorRecipient(authUser, companyModel, projectModel, user.Username, user.LfEmail, "removed", "from", "you are no longer authorized to contribute to")
	}
}

func (s service) createEventLogEntries(companyModel *models.Company, projectModel *models.Project, userModel *models.User, approvalList *models.ApprovalList) {
	for _, value := range approvalList.AddEmailApprovalList {
		// Send an event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:         events.ClaApprovalListUpdated,
			ProjectID:         projectModel.ProjectID,
			ProjectModel:      projectModel,
			CompanyID:         companyModel.CompanyID,
			CompanyModel:      companyModel,
			LfUsername:        userModel.LfUsername,
			UserID:            userModel.UserID,
			UserModel:         userModel,
			ExternalProjectID: projectModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddEmailData{
				UserName:          userModel.LfUsername,
				UserEmail:         userModel.LfEmail,
				UserLFID:          userModel.UserID,
				ApprovalListEmail: value,
			},
		})
	}
	for _, value := range approvalList.RemoveEmailApprovalList {
		// Send an event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:         events.ClaApprovalListUpdated,
			ProjectID:         projectModel.ProjectID,
			ProjectModel:      projectModel,
			CompanyID:         companyModel.CompanyID,
			CompanyModel:      companyModel,
			LfUsername:        userModel.LfUsername,
			UserID:            userModel.UserID,
			UserModel:         userModel,
			ExternalProjectID: projectModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveEmailData{
				UserName:          userModel.LfUsername,
				UserEmail:         userModel.LfEmail,
				UserLFID:          userModel.UserID,
				ApprovalListEmail: value,
			},
		})
	}
	for _, value := range approvalList.AddDomainApprovalList {
		// Send an event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:         events.ClaApprovalListUpdated,
			ProjectID:         projectModel.ProjectID,
			ProjectModel:      projectModel,
			CompanyID:         companyModel.CompanyID,
			CompanyModel:      companyModel,
			LfUsername:        userModel.LfUsername,
			UserID:            userModel.UserID,
			UserModel:         userModel,
			ExternalProjectID: projectModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddDomainData{
				UserName:           userModel.LfUsername,
				UserEmail:          userModel.LfEmail,
				UserLFID:           userModel.UserID,
				ApprovalListDomain: value,
			},
		})
	}
	for _, value := range approvalList.RemoveDomainApprovalList {
		// Send an event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:         events.ClaApprovalListUpdated,
			ProjectID:         projectModel.ProjectID,
			ProjectModel:      projectModel,
			CompanyID:         companyModel.CompanyID,
			CompanyModel:      companyModel,
			LfUsername:        userModel.LfUsername,
			UserID:            userModel.UserID,
			UserModel:         userModel,
			ExternalProjectID: projectModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveDomainData{
				UserName:           userModel.LfUsername,
				UserEmail:          userModel.LfEmail,
				UserLFID:           userModel.UserID,
				ApprovalListDomain: value,
			},
		})
	}
	for _, value := range approvalList.AddGithubUsernameApprovalList {
		// Send an event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:         events.ClaApprovalListUpdated,
			ProjectID:         projectModel.ProjectID,
			ProjectModel:      projectModel,
			CompanyID:         companyModel.CompanyID,
			CompanyModel:      companyModel,
			LfUsername:        userModel.LfUsername,
			UserID:            userModel.UserID,
			UserModel:         userModel,
			ExternalProjectID: projectModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddGitHubUsernameData{
				UserName:                   userModel.LfUsername,
				UserEmail:                  userModel.LfEmail,
				UserLFID:                   userModel.UserID,
				ApprovalListGitHubUsername: value,
			},
		})
	}
	for _, value := range approvalList.RemoveGithubUsernameApprovalList {
		// Send an event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:         events.ClaApprovalListUpdated,
			ProjectID:         projectModel.ProjectID,
			ProjectModel:      projectModel,
			CompanyID:         companyModel.CompanyID,
			CompanyModel:      companyModel,
			LfUsername:        userModel.LfUsername,
			UserID:            userModel.UserID,
			UserModel:         userModel,
			ExternalProjectID: projectModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveGitHubUsernameData{
				UserName:                   userModel.LfUsername,
				UserEmail:                  userModel.LfEmail,
				UserLFID:                   userModel.UserID,
				ApprovalListGitHubUsername: value,
			},
		})
	}
	for _, value := range approvalList.AddGithubOrgApprovalList {
		// Send an event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:         events.ClaApprovalListUpdated,
			ProjectID:         projectModel.ProjectID,
			ProjectModel:      projectModel,
			CompanyID:         companyModel.CompanyID,
			CompanyModel:      companyModel,
			LfUsername:        userModel.LfUsername,
			UserID:            userModel.UserID,
			UserModel:         userModel,
			ExternalProjectID: projectModel.ProjectExternalID,
			EventData: &events.CLAApprovalListAddGitHubOrgData{
				UserName:              userModel.LfUsername,
				UserEmail:             userModel.LfEmail,
				UserLFID:              userModel.UserID,
				ApprovalListGitHubOrg: value,
			},
		})
	}
	for _, value := range approvalList.RemoveGithubOrgApprovalList {
		// Send an event
		s.eventsService.LogEvent(&events.LogEventArgs{
			EventType:         events.ClaApprovalListUpdated,
			ProjectID:         projectModel.ProjectID,
			ProjectModel:      projectModel,
			CompanyID:         companyModel.CompanyID,
			CompanyModel:      companyModel,
			LfUsername:        userModel.LfUsername,
			UserID:            userModel.UserID,
			UserModel:         userModel,
			ExternalProjectID: projectModel.ProjectExternalID,
			EventData: &events.CLAApprovalListRemoveGitHubOrgData{
				UserName:              userModel.LfUsername,
				UserEmail:             userModel.LfEmail,
				UserLFID:              userModel.UserID,
				ApprovalListGitHubOrg: value,
			},
		})
	}
}

func (s service) GetClaGroupICLASignatures(claGroupID string, searchTerm *string) (*models.IclaSignatures, error) {
	return s.repo.GetClaGroupICLASignatures(claGroupID, searchTerm)
}

func (s service) GetClaGroupCorporateContributors(claGroupID string, companyID *string, searchTerm *string) (*models.CorporateContributorList, error) {
	return s.repo.GetClaGroupCorporateContributors(claGroupID, companyID, searchTerm)
}

// sendRequestAccessEmailToContributors sends the request access email to the specified contributors
func sendRequestAccessEmailToContributorRecipient(authUser *auth.User, companyModel *models.Company, projectModel *models.Project, recipientName, recipientAddress, addRemove, toFrom, authorizedString string) {
	companyName := companyModel.CompanyName
	projectName := projectModel.ProjectName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: Approval List Update for %s on %s", companyName, projectName)
	recipients := []string{recipientAddress}
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the project %s.</p>
<p>You have been %s %s the Approval List of %s for %s by CLA Manager %s. This means that %s on behalf of %s.</p>
<p>If you had previously submitted one or more pull requests to %s that had failed, you should 
close and re-open the pull request to force a recheck by the EasyCLA system.</p>
%s
%s`,
		recipientName, projectName, addRemove, toFrom,
		companyName, projectName, authUser.UserName, authorizedString, projectName, projectName,
		utils.GetEmailHelpContent(projectModel.Version == utils.V2), utils.GetEmailSignOffContent())

	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

// getBestEmail is a helper function to return the best email address for the user model
func getBestEmail(claManager models.User) string {
	if claManager.LfEmail != "" {
		return claManager.LfEmail
	}

	for _, email := range claManager.Emails {
		if email != "" {
			return email
		}
	}

	return ""
}
