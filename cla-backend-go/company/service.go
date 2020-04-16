// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package company

import (
	"errors"
	"fmt"
	"strings"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/user"
)

type service struct {
	repo                CompanyRepository
	userDynamoRepo      user.RepositoryService
	corporateConsoleURL string
}

const (
	// StatusPending indicates the invitation status is pending
	StatusPending = "Pending Approval"
)

// Service interface defining the public functions
type Service interface { // nolint
	GetCompanies() (*models.Companies, error)
	GetCompany(companyID string) (*models.Company, error)
	SearchCompanyByName(companyName string, nextKey string) (*models.Companies, error)
	GetCompaniesByUserManager(userID string) (*models.Companies, error)
	GetCompaniesByUserManagerWithInvites(userID string) (*models.CompaniesWithInvites, error)

	AddPendingCompanyInviteRequest(companyID string, userID string) error
	GetCompanyInviteRequests(companyID string) ([]models.CompanyInviteUser, error)
	GetCompanyUserInviteRequests(companyID string, userID string) (*models.CompanyInviteUser, error)
	RejectCompanyInviteRequest(companyID string, userID string) error
	DeletePendingCompanyInviteRequest(CompanyID string, InviteID string, lfID string) error

	AddUserToCompanyAccessList(companyID string, inviteID string, lfid string) error
	SendRequestAccessEmail(companyID string, user *user.CLAUser) error
	//sendRejectionEmail(company *models.Company, recipientAddress string, rejectedUser *user.CLAUser) error
}

// NewService creates a new company service object
func NewService(repo CompanyRepository, corporateConsoleURL string, userDynamoRepo user.RepositoryService) Service {
	return service{
		repo:                repo,
		userDynamoRepo:      userDynamoRepo,
		corporateConsoleURL: corporateConsoleURL,
	}
}

// GetCompanies returns all the companies
func (s service) GetCompanies() (*models.Companies, error) {
	return s.repo.GetCompanies()
}

// GetCompany returns the company associated with the company ID
func (s service) GetCompany(companyID string) (*models.Company, error) {
	return s.repo.GetCompany(companyID)
}

// SearchCompanyByName locates companies by the matching name and return any potential matches
func (s service) SearchCompanyByName(companyName string, nextKey string) (*models.Companies, error) {
	companies, err := s.repo.SearchCompanyByName(companyName, nextKey)
	if err != nil {
		log.Warnf("Error searching company by company name: %s, error: %v", companyName, err)
		return nil, err
	}

	return companies, nil
}

// GetCompanyUserManager the get a list of companies when provided the company id and user manager
func (s service) GetCompaniesByUserManager(userID string) (*models.Companies, error) {
	userModel, err := s.userDynamoRepo.GetUser(userID)
	if err != nil {
		log.Warnf("Unable to lookup user by user id: %s, error: %v", userID, err)
		return nil, err
	}

	return s.repo.GetCompaniesByUserManager(userID, userModel)
}

// GetCompanyUserManagerWithInvites the get a list of companies including status when provided the company id and user manager
func (s service) GetCompaniesByUserManagerWithInvites(userID string) (*models.CompaniesWithInvites, error) {
	userModel, err := s.userDynamoRepo.GetUser(userID)
	if err != nil {
		log.Warnf("Unable to lookup user by user id: %s, error: %v", userID, err)
		return nil, err
	}

	return s.repo.GetCompaniesByUserManagerWithInvites(userID, userModel)
}

// AddPendingCompanyInviteRequest adds a new company invite request
func (s service) AddPendingCompanyInviteRequest(companyID string, userID string) error {
	return s.repo.AddPendingCompanyInviteRequest(companyID, userID)
}

// GetCompanyInviteRequests returns a list of company invites when provided the company ID
func (s service) GetCompanyInviteRequests(companyID string) ([]models.CompanyInviteUser, error) {
	companyInvites, err := s.repo.GetCompanyInviteRequests(companyID)
	if err != nil {
		return nil, err
	}

	var users []models.CompanyInviteUser
	for _, invite := range companyInvites {

		dbUserModel, err := s.userDynamoRepo.GetUser(invite.UserID)
		if err != nil {
			log.Warnf("Error fetching user with userID: %s, error: %v", invite.UserID, err)
			continue
		}

		// Default status is pending if there's a record but no status
		if invite.Status == "" {
			invite.Status = StatusPending
		}

		users = append(users, models.CompanyInviteUser{
			InviteID:  invite.CompanyInviteID,
			UserName:  dbUserModel.UserName,
			UserEmail: dbUserModel.LFEmail,
			UserLFID:  dbUserModel.LFUsername,
			Status:    invite.Status,
		})
	}

	return users, nil

}

// GetCompanyUserInviteRequests returns a list of company invites when provided the company ID
func (s service) GetCompanyUserInviteRequests(companyID string, userID string) (*models.CompanyInviteUser, error) {
	invite, err := s.repo.GetCompanyUserInviteRequests(companyID, userID)
	if err != nil {
		return nil, err
	}

	if invite == nil {
		return nil, nil
	}

	//var users []models.CompanyInviteUser

	dbUserModel, err := s.userDynamoRepo.GetUser(invite.UserID)
	if err != nil {
		log.Warnf("Error fetching company invite user with company id: %s and user id: %s, error: %v",
			companyID, userID, err)
		return nil, err
	}

	// Default status is pending if there's a record but no status
	if invite.Status == "" {
		invite.Status = StatusPending
	}

	// Let's do a company lookup so we can grab the company name
	company, err := s.repo.GetCompany(companyID)
	if err != nil {
		log.Warnf("Error fetching company with company id: %s, error: %v", companyID, err)
		return nil, err
	}

	return &models.CompanyInviteUser{
		InviteID:    invite.CompanyInviteID,
		UserName:    dbUserModel.UserName,
		UserEmail:   dbUserModel.LFEmail,
		UserLFID:    dbUserModel.LFUsername,
		Status:      invite.Status,
		CompanyName: company.CompanyName,
	}, nil
}

// RejectCompanyInviteRequest updates the invite with the rejection status
func (s service) RejectCompanyInviteRequest(companyID string, userID string) error {
	return s.repo.RejectCompanyInviteRequest(companyID, userID)
}

// DeletePendingCompanyInviteRequest deletes the pending company invite request when provided the invite ID
func (s service) DeletePendingCompanyInviteRequest(companyID string, inviteID string, lfID string) error {
	// When a CLA Manager Declines a pending invite, remove the invite from the table
	company, err := s.repo.GetCompany(companyID)
	if err != nil {
		log.Warnf("Error retrieving company by company ID: %s, error: %v", companyID, err)
		return err
	}
	log.Debugf("Deleting Company Invite Request inviteID : %s", inviteID)

	userProfile, err := s.userDynamoRepo.GetUserAndProfilesByLFID(lfID)
	if err != nil {
		log.Warnf("Error getting user profile by LFID: %s, error: %v", lfID, err)
		return nil
	}

	recipientEmailAddress := userProfile.LFEmail

	err = s.repo.DeletePendingCompanyInviteRequest(inviteID)
	if err != nil {
		log.Warnf("Error deleting the pending company invite with invite ID: %s, error: %v", inviteID, err)
		return err
	}

	err = s.sendRejectionEmail(company, recipientEmailAddress, &userProfile)
	if err != nil {
		return errors.New("failed to send notification email")
	}

	return nil
}

// AddUserToCompanyAccessList adds a user to the specified company
func (s service) AddUserToCompanyAccessList(companyID string, inviteID string, lfid string) error {
	// call the get company function
	company, err := s.repo.GetCompany(companyID)
	if err != nil {
		log.Warnf("Error retrieving company by company ID: %s, error: %v", companyID, err)
		return err
	}

	// perform ACL check
	// check if user already exists in the company acl
	for _, acl := range company.CompanyACL {
		if acl == lfid {
			log.Warnf(fmt.Sprintf("User %s has already been added to the company acl", lfid))
			err = s.repo.DeletePendingCompanyInviteRequest(inviteID)
			if err != nil {
				log.Warnf("Error deleting pending company invite request with inviteID: %s, error: %v", inviteID, err)
				return fmt.Errorf("failed to delete pending invite")
			}
			return nil
		}
	}
	// add user to string set
	company.CompanyACL = append(company.CompanyACL, lfid)

	err = s.repo.UpdateCompanyAccessList(companyID, company.CompanyACL)
	if err != nil {
		log.Warnf("Error updating company access list with company ID: %s, company ACL: %v, error: %v", companyID, company.CompanyACL, err)
		return err
	}

	userProfile, err := s.userDynamoRepo.GetUserAndProfilesByLFID(lfid)
	if err != nil {
		log.Warnf("Error getting user profile by LFID: %s, error: %v", lfid, err)
		return nil
	}

	recipientEmailAddress := userProfile.LFEmail

	err = sendApprovalEmail(company.CompanyName, recipientEmailAddress, &userProfile)
	if err != nil {
		return errors.New("failed to send notification email")
	}

	// Remove pending invite ID once approval emails are sent
	err = s.repo.DeletePendingCompanyInviteRequest(inviteID)
	if err != nil {
		return fmt.Errorf("failed to delete pending invite")
	}

	return nil
}

// sendApprovalEmail sends the approval email when provided the company name, address and user object
func sendApprovalEmail(companyName string, recipientAddress string, user *user.CLAUser) error {
	var (
		Recipient = recipientAddress
		Subject   = "CLA: Approval of Access for Corporate CLA"

		//The email body for recipients with non-HTML email clients.
		TextBody = fmt.Sprintf(`Hello %s,

You have now been granted access to the organization: %s

	%s <%s>

- Linux Foundation CLA System`, user.Name, companyName, user.LFUsername, user.LFEmail)
		// The character encoding for the email.
	)

	err := utils.SendEmail(Subject, TextBody, []string{Recipient})
	if err != nil {
		log.Warnf("Error sending mail, error: %v", err)
		return err
	}
	log.Debugf("Sent '%s' email to: %s", Subject, Recipient)

	return nil
}

// sendRejectionEmail sends the rejection email
func (s service) sendRejectionEmail(company *models.Company, recipientAddress string, rejectedUser *user.CLAUser) error {
	// Get CLAUser admin list
	log.Debugf("Processing rejection email for User: %s for Company: %s ", rejectedUser.LFUsername, company.CompanyName)
	var admins []*user.CLAUser
	for _, acl := range company.CompanyACL {
		admin, err := s.userDynamoRepo.GetUserAndProfilesByLFID(acl)
		if err != nil {
			log.Warnf("Error fetching user profile using admin: %s, error: %v", admin, err)
			continue
		}
		admins = append(admins, &admin)
	}

	// String builder to return 'Manager <email>' list
	var sb strings.Builder
	for _, admin := range admins {
		sb.WriteString(fmt.Sprintf("- %s <%s>\n", admin.Name, admin.LFEmail))
	}

	var (
		Recipient = recipientAddress
		Subject   = "CLA: Denial of Access for Corporate CLA "
		TextBody  = fmt.Sprintf(` Hello %s,
		Your request to become a CLA Manager for the organization: %s was denied.
		If you have further questions, contact one of the existing CLA Managers :
		%s 
		
		- Linux Foundation CLA System`, rejectedUser.Name, company.CompanyName, sb.String())
	)
	log.Debugf("acls : %v", company.CompanyACL)
	err := utils.SendEmail(Subject, TextBody, []string{Recipient})
	if err != nil {
		log.Warnf("Error sending mail, error: %v", err)
		return err
	}
	log.Debugf("Send '%s' email to: %s", Subject, Recipient)

	return nil

}

// SendRequestAccessEmail sends the request access e-mail when provided the company ID and user object
func (s service) SendRequestAccessEmail(companyID string, user *user.CLAUser) error {

	log.Debugf("Processing send invite access email for company ID: %s, user: %+v", companyID, user)

	// Get Company
	company, err := s.repo.GetCompany(companyID)
	if err != nil {
		log.Warnf("Error fetching company by company ID: %s, error: %v", companyID, err)
		return err
	}

	// Add a pending request to the company-invites table
	err = s.repo.AddPendingCompanyInviteRequest(companyID, user.UserID)
	if err != nil {
		log.Warnf("Error adding pending company invite request using company ID: %s, user ID: %s, error: %v", companyID, user.UserID, err)
		return err
	}

	// Send Email to every CLA Manager in the Company ACL
	Subject := "CLA: Request of Access for Corporate CLA Manager"

	for _, admin := range company.CompanyACL {
		// Retrieve admin's user profile for email and name
		adminUser, err := s.userDynamoRepo.GetUserAndProfilesByLFID(admin)
		if err != nil {
			log.Warnf("Error fetching user profile using admin: %s, error: %v", admin, err)
			return err
		}

		TextBody := fmt.Sprintf(`Hello %s, 

The following user is requesting access to your organization: %s

	%s <%s>

Please navigate to the Corporate Console using the link below, where you can approve this user's request.

%s

- Linux Foundation CLA System`, adminUser.Name, company.CompanyName, user.LFUsername, user.LFEmail, s.corporateConsoleURL)

		err = utils.SendEmail(Subject, TextBody, []string{adminUser.LFEmail})
		if err != nil {
			log.Warnf("Error sending mail, error: %v", err)
			return err
		}
		log.Debugf("Sent '%s' email to: %s", Subject, adminUser.LFEmail)
	}

	return nil
}
