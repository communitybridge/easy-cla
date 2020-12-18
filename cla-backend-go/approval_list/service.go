// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package approval_list

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/utils"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/communitybridge/easycla/cla-backend-go/company"
	"github.com/communitybridge/easycla/cla-backend-go/project"
	"github.com/communitybridge/easycla/cla-backend-go/users"

	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
)

// errors
var (
	ErrCclaApprovalRequestAlreadyExists = errors.New("approval request already exist")
)

// constants
const (
	DontLoadRepoDetails = true
)

// IService interface defines the service methods/functions
type IService interface {
	AddCclaWhitelistRequest(ctx context.Context, companyID string, claGroupID string, args models.CclaWhitelistRequestInput) (string, error)
	ApproveCclaWhitelistRequest(ctx context.Context, companyID, claGroupID, requestID string) error
	RejectCclaWhitelistRequest(ctx context.Context, companyID, claGroupID, requestID string) error
	ListCclaWhitelistRequest(companyID string, claGroupID, status *string) (*models.CclaWhitelistRequestList, error)
	ListCclaWhitelistRequestByCompanyProjectUser(companyID string, claGroupID, status, userID *string) (*models.CclaWhitelistRequestList, error)
}

type service struct {
	repo           IRepository
	userRepo       users.UserRepository
	companyRepo    company.IRepository
	projectRepo    project.ProjectRepository
	signatureRepo  signatures.SignatureRepository
	corpConsoleURL string
	httpClient     *http.Client
}

// NewService creates a new whitelist service
func NewService(repo IRepository, userRepo users.UserRepository, companyRepo company.IRepository, projectRepo project.ProjectRepository, signatureRepo signatures.SignatureRepository, corpConsoleURL string, httpClient *http.Client) IService {
	return service{
		repo:           repo,
		userRepo:       userRepo,
		companyRepo:    companyRepo,
		projectRepo:    projectRepo,
		signatureRepo:  signatureRepo,
		corpConsoleURL: corpConsoleURL,
		httpClient:     httpClient,
	}
}

func (s service) AddCclaWhitelistRequest(ctx context.Context, companyID string, claGroupID string, args models.CclaWhitelistRequestInput) (string, error) {
	list, err := s.ListCclaWhitelistRequestByCompanyProjectUser(companyID, &claGroupID, nil, &args.ContributorID)
	if err != nil {
		log.Warnf("AddCclaWhitelistRequest - error looking up existing contributor invite requests for company: %s, project: %s, user by id: %s with name: %s, email: %s, error: %+v",
			companyID, claGroupID, args.ContributorID, args.ContributorName, args.ContributorEmail, err)
		return "", err
	}
	for _, item := range list.List {
		if item.RequestStatus == "pending" || item.RequestStatus == "approved" {
			log.Warnf("AddCclaWhitelistRequest - found existing contributor invite - id: %s, request for company: %s, project: %s, user by id: %s with name: %s, email: %s",
				list.List[0].RequestID, companyID, claGroupID, args.ContributorID, args.ContributorName, args.ContributorEmail)
			return "", ErrCclaApprovalRequestAlreadyExists
		}
	}
	companyModel, err := s.companyRepo.GetCompany(ctx, companyID)
	if err != nil {
		log.Warnf("AddCclaWhitelistRequest - unable to lookup company by id: %s, error: %+v", companyID, err)
		return "", err
	}
	claGroupModel, err := s.projectRepo.GetCLAGroupByID(ctx, claGroupID, DontLoadRepoDetails)
	if err != nil {
		log.Warnf("AddCclaWhitelistRequest - unable to lookup project by id: %s, error: %+v", claGroupID, err)
		return "", err
	}
	userModel, err := s.userRepo.GetUser(args.ContributorID)
	if err != nil {
		log.Warnf("AddCclaWhitelistRequest - unable to lookup user by id: %s with name: %s, email: %s, error: %+v",
			args.ContributorID, args.ContributorName, args.ContributorEmail, err)
		return "", err
	}
	if userModel == nil {
		log.Warnf("AddCclaWhitelistRequest - unable to lookup user by id: %s with name: %s, email: %s, error: user object not found",
			args.ContributorID, args.ContributorName, args.ContributorEmail)
		return "", errors.New("invalid user")
	}

	signed, approved := true, true
	sortOrder := utils.SortOrderAscending
	pageSize := int64(5)
	sig, sigErr := s.signatureRepo.GetProjectCompanySignatures(ctx, companyID, claGroupID, &signed, &approved, nil, &sortOrder, &pageSize)
	if sigErr != nil || sig == nil || sig.Signatures == nil {
		log.Warnf("AddCclaWhitelistRequest - unable to lookup signature by company id: %s project id: %s - (or no managers), sig: %+v, error: %+v",
			companyID, claGroupID, sig, err)
		return "", err
	}

	requestID, addErr := s.repo.AddCclaWhitelistRequest(companyModel, claGroupModel, userModel, args.ContributorName, args.ContributorEmail)
	if addErr != nil {
		log.Warnf("AddCclaWhitelistRequest - unable to add Approval Request for id: %s with name: %s, email: %s, error: %+v",
			args.ContributorID, args.ContributorName, args.ContributorEmail, addErr)
	}

	// Send the emails to the CLA managers for this CCLA Signature which includes the managers in the ACL list
	s.sendRequestSentEmail(companyModel, claGroupModel, sig.Signatures[0], args.ContributorName, args.ContributorEmail, args.RecipientName, args.RecipientEmail, args.Message)

	return requestID, nil
}

// ApproveCclaWhitelistRequest is the handler for the approve CLA request
func (s service) ApproveCclaWhitelistRequest(ctx context.Context, companyID, claGroupID, requestID string) error {
	err := s.repo.ApproveCclaWhitelistRequest(requestID)
	if err != nil {
		log.Warnf("ApproveCclaWhitelistRequest - problem updating approved list with 'approved' status for request: %s, error: %+v",
			requestID, err)
		return err
	}

	requestModel, err := s.repo.GetCclaWhitelistRequest(requestID)
	if err != nil {
		log.Warnf("ApproveCclaWhitelistRequest - unable to lookup request by id: %s, error: %+v", requestID, err)
		return err
	}

	companyModel, err := s.companyRepo.GetCompany(ctx, companyID)
	if err != nil {
		log.Warnf("ApproveCclaWhitelistRequest - unable to lookup company by id: %s, error: %+v", companyID, err)
		return err
	}
	claGroupModel, err := s.projectRepo.GetCLAGroupByID(ctx, claGroupID, DontLoadRepoDetails)
	if err != nil {
		log.Warnf("ApproveCclaWhitelistRequest - unable to lookup project by id: %s, error: %+v", claGroupID, err)
		return err
	}

	if requestModel.UserEmails == nil {
		msg := fmt.Sprintf("ApproveCclaWhitelistRequest - unable to send approval email - email missing for request: %+v, error: %+v",
			requestModel, err)
		log.Warnf(msg)
		return errors.New(msg)
	}

	// Send the email
	sendRequestApprovedEmailToRecipient(companyModel, claGroupModel, requestModel.UserName, requestModel.UserEmails[0])

	return nil
}

// RejectCclaWhitelistRequest is the handler for the decline CLA request
func (s service) RejectCclaWhitelistRequest(ctx context.Context, companyID, claGroupID, requestID string) error {
	err := s.repo.RejectCclaWhitelistRequest(requestID)
	if err != nil {
		log.Warnf("RejectCclaWhitelistRequest - problem updating approved list with 'rejected' status for request: %s, error: %+v", requestID, err)
		return err
	}

	requestModel, err := s.repo.GetCclaWhitelistRequest(requestID)
	if err != nil {
		log.Warnf("RejectCclaWhitelistRequest - unable to lookup request by id: %s, error: %+v", requestID, err)
		return err
	}

	companyModel, err := s.companyRepo.GetCompany(ctx, companyID)
	if err != nil {
		log.Warnf("RejectCclaWhitelistRequest - unable to lookup company by id: %s, error: %+v", companyID, err)
		return err
	}
	claGroupModel, err := s.projectRepo.GetCLAGroupByID(ctx, claGroupID, DontLoadRepoDetails)
	if err != nil {
		log.Warnf("RejectCclaWhitelistRequest - unable to lookup project by id: %s, error: %+v", claGroupID, err)
		return err
	}

	signed, approved := true, true
	sortOrder := utils.SortOrderAscending
	pageSize := int64(5)
	sig, sigErr := s.signatureRepo.GetProjectCompanySignatures(ctx, companyID, claGroupID, &signed, &approved, nil, &sortOrder, &pageSize)
	if sigErr != nil || sig == nil || sig.Signatures == nil {
		log.Warnf("RejectCclaWhitelistRequest - unable to lookup signature by company id: %s project id: %s - (or no managers), sig: %+v, error: %+v",
			companyID, claGroupID, sig, err)
		return err
	}

	if requestModel.UserEmails == nil {
		msg := fmt.Sprintf("RejectCclaWhitelistRequest - unable to send approval email - email missing for request: %+v, error: %+v",
			requestModel, err)
		log.Warnf(msg)
		return errors.New(msg)
	}

	// Send the email
	s.sendRequestRejectedEmailToRecipient(companyModel, claGroupModel, sig.Signatures[0], requestModel.UserName, requestModel.UserEmails[0])

	return nil
}

// ListCclaWhitelistRequest is the handler for the list CLA request
func (s service) ListCclaWhitelistRequest(companyID string, claGroupID, status *string) (*models.CclaWhitelistRequestList, error) {
	return s.repo.ListCclaWhitelistRequest(companyID, claGroupID, status, nil)
}

// ListCclaWhitelistRequestByCompanyProjectUser is the handler for the list CLA request
func (s service) ListCclaWhitelistRequestByCompanyProjectUser(companyID string, claGroupID, status, userID *string) (*models.CclaWhitelistRequestList, error) {
	return s.repo.ListCclaWhitelistRequest(companyID, claGroupID, status, userID)
}

// sendRequestSentEmail sends emails to the CLA managers specified in the signature record
func (s service) sendRequestSentEmail(companyModel *models.Company, claGroupModel *models.ClaGroup, signature *models.Signature, contributorName, contributorEmail, recipientName, recipientEmail, message string) {

	// If we have an override name and email from the request - possibly from the web form where the user selected the
	// CLA Manager Name/Email from a list, send this to this recipient (CLA Manager) - otherwise we will send to all
	// CLA Managers on the Signature ACL
	if recipientName != "" && recipientEmail != "" {
		s.sendRequestEmailToRecipient(companyModel, claGroupModel, contributorName, contributorEmail, recipientName, recipientEmail, message)
		return
	}

	// Send an email to each manager
	for _, manager := range signature.SignatureACL {

		// Need to determine which email...
		var whichEmail = ""
		if manager.LfEmail != "" {
			whichEmail = manager.LfEmail
		}

		// If no LF Email try to grab the first other email in their email list
		if manager.LfEmail == "" && manager.Emails != nil {
			whichEmail = manager.Emails[0]
		}
		if whichEmail == "" {
			log.Warnf("unable to send email to manager: %+v - no email on file...", manager)
		} else {
			// Send the email
			s.sendRequestEmailToRecipient(companyModel, claGroupModel, contributorName, contributorEmail, manager.Username, whichEmail, message)
		}
	}
}

// sendRequestEmailToRecipient generates and sends an email to the specified recipient
func (s service) sendRequestEmailToRecipient(companyModel *models.Company, claGroupModel *models.ClaGroup, contributorName, contributorEmail, recipientName, recipientAddress, message string) {
	companyName := companyModel.CompanyName
	projectName := claGroupModel.ProjectName

	var optionalMessage = ""
	if message != "" {
		optionalMessage = fmt.Sprintf("<p>%s included the following message in the request:</p>", contributorName)
		optionalMessage += fmt.Sprintf("<br/><p>%s</p><br/", message)
	}

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: Request to Authorize %s for %s", contributorName, projectName)
	recipients := []string{recipientAddress}
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the project %s.</p>
<p>%s (%s) has requested to be added to the Allow List as an authorized contributor from
%s to the project %s. You are receiving this message as a CLA Manager from %s for
%s.</p>
%s
<p>If you want to add them to the Allow List, please
<a href="https://%s#/company/%s" target="_blank">log into the EasyCLA Corporate
Console</a>, where you can approve this user's request by selecting the 'Manage Approved List' and adding the
contributor's email, the contributor's entire email domain, their GitHub ID or the entire GitHub Organization for the
repository. This will permit them to begin contributing to %s on behalf of %s.</p>
<p>If you are not certain whether to add them to the Allow List, please reach out to them directly to discuss.</p>
%s
%s`,
		recipientName, projectName, contributorName, contributorEmail,
		companyName, projectName, companyName, projectName,
		optionalMessage, s.corpConsoleURL,
		companyModel.CompanyID, projectName, companyName,
		utils.GetEmailHelpContent(claGroupModel.Version == utils.V2), utils.GetEmailSignOffContent())

	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

// sendRequestRejectedEmailToRecipient generates and sends an email to the specified recipient
func (s service) sendRequestRejectedEmailToRecipient(companyModel *models.Company, claGroupModel *models.ClaGroup, signature *models.Signature, recipientName, recipientAddress string) {
	companyName := companyModel.CompanyName
	projectName := claGroupModel.ProjectName

	var claManagerText = ""
	claManagerText += "<ul>"

	// Build a fancy text string with CLA Manager name <email> as an HTML unordered list
	for _, manager := range signature.SignatureACL {

		// Need to determine which email...
		var whichEmail = ""
		if manager.LfEmail != "" {
			whichEmail = manager.LfEmail
		}

		// If no LF Email try to grab the first other email in their email list
		if manager.LfEmail == "" && manager.Emails != nil {
			whichEmail = manager.Emails[0]
		}
		if whichEmail == "" {
			log.Warnf("unable to send email to manager: %+v - no email on file...", manager)
		} else {
			claManagerText += fmt.Sprintf("<li>%s <%s></li>", manager.Username, whichEmail)
		}
	}
	claManagerText += "</ul>"

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: Approval List Request Denied for Project %s", projectName)
	recipients := []string{recipientAddress}
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the project %s.</p>
<p>Your request to get added to the approval list from %s for %s was denied by one of the existing CLA Managers.
If you have further questions about this denial, please contact one of the existing CLA Managers from
%s for %s:</p>
%s
%s
%s`,
		recipientName, projectName,
		companyName, projectName, companyName, projectName,
		claManagerText,
		utils.GetEmailHelpContent(claGroupModel.Version == utils.V2), utils.GetEmailSignOffContent())

	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}

func requestApprovedEmailToRecipientContent(companyModel *models.Company, claGroupModel *models.ClaGroup, recipientName, recipientAddress string) (string, string, []string) {
	companyName := companyModel.CompanyName

	// subject string, body string, recipients []string
	subject := fmt.Sprintf("EasyCLA: Approved List Request Accepted for %s", companyName)
	recipients := []string{recipientAddress}
	body := fmt.Sprintf(`
<p>Hello %s,</p>
<p>This is a notification email from EasyCLA regarding the company %s.</p>
<p>You have now been added to the approval list for %s. </p>
<p>To get started, please navigate back to GitHub or Gerrit and start the authorization process. Once you select the
authorization link, you will be directed to the EasyCLA Contributor Console. GitHub users will need to authorize the
tool to see your GitHub user name and email. Gerrit users will first need to log in with their LF Account. On the
console landing page, select the corporate agreement option. To finish, search and select your company to acknowledge
your association with your company. This will complete the authorization process. For GitHub users, your pull request
will refresh and confirm that you are authorized. For Gerrit users, please log out of the UI and back in to complete the
authorization. </p>
%s
%s`,
		recipientName, companyName,
		companyName,
		utils.GetEmailHelpContent(claGroupModel.Version == utils.V2),
		utils.GetEmailSignOffContent())

	return subject, body, recipients
}

// sendRequestApprovedEmailToRecipient generates and sends an email to the specified recipient
func sendRequestApprovedEmailToRecipient(companyModel *models.Company, claGroupModel *models.ClaGroup, recipientName, recipientAddress string) {
	subject, body, recipients := requestApprovedEmailToRecipientContent(companyModel, claGroupModel, recipientName, recipientAddress)
	err := utils.SendEmail(subject, body, recipients)
	if err != nil {
		log.Warnf("problem sending email with subject: %s to recipients: %+v, error: %+v", subject, recipients, err)
	} else {
		log.Debugf("sent email with subject: %s to recipients: %+v", subject, recipients)
	}
}
