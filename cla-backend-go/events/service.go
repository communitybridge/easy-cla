// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package events

import (
	"context"
	"errors"
	"fmt"

	project_service "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"

	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	eventOps "github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/events"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
)

// constants
const (
	ReturnAllEvents     = true
	LoadRepoDetails     = true
	DontLoadRepoDetails = false
)

// Service interface defines methods of event service
type Service interface {
	LogEvent(args *LogEventArgs)
	LogEventWithContext(ctx context.Context, args *LogEventArgs)
	SearchEvents(params *eventOps.SearchEventsParams) (*models.EventList, error)
	GetRecentEvents(paramPageSize *int64) (*models.EventList, error)

	GetFoundationEvents(foundationSFID string, nextKey *string, paramPageSize *int64, all bool, searchTerm *string) (*models.EventList, error)
	GetClaGroupEvents(claGroupID string, nextKey *string, paramPageSize *int64, all bool, searchTerm *string) (*models.EventList, error)
	GetCompanyFoundationEvents(companySFID, companyID, foundationSFID string, nextKey *string, paramPageSize *int64, all bool) (*models.EventList, error)
	GetCompanyClaGroupEvents(companySFID, companyID, claGroupID string, nextKey *string, paramPageSize *int64, all bool) (*models.EventList, error)
	GetCompanyEvents(companyID, eventType string, nextKey *string, paramPageSize *int64, all bool) (*models.EventList, error)
}

// CombinedRepo contains the various methods of other repositories
type CombinedRepo interface {
	GetCLAGroupByID(ctx context.Context, claGroupID string, loadRepoDetails bool) (*models.ClaGroup, error)
	GetCompany(ctx context.Context, companyID string) (*models.Company, error)
	GetUserByUserName(userName string, fullMatch bool) (*models.User, error)
	GetUser(userID string) (*models.User, error)
}

type service struct {
	repo         Repository
	combinedRepo CombinedRepo
}

// NewService creates new instance of event service
func NewService(repo Repository, combinedRepo CombinedRepo) Service {
	return &service{
		repo:         repo,
		combinedRepo: combinedRepo,
	}
}

func (s *service) CreateEvent(event models.Event) error {
	return s.repo.CreateEvent(&event)
}

// SearchEvents service definition
func (s *service) SearchEvents(params *eventOps.SearchEventsParams) (*models.EventList, error) {
	const defaultPageSize int64 = 50
	var pageSize = defaultPageSize
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	return s.repo.SearchEvents(params, pageSize)
}

// GetRecentEvents returns event list of recent events
func (s *service) GetRecentEvents(paramPageSize *int64) (*models.EventList, error) {
	const defaultPageSize int64 = 10
	var pageSize = defaultPageSize
	if paramPageSize != nil {
		pageSize = *paramPageSize
	}
	return s.repo.GetRecentEvents(pageSize)
}

// GetFoundationEvents returns the list of foundation events
func (s *service) GetFoundationEvents(foundationSFID string, nextKey *string, paramPageSize *int64, all bool, searchTerm *string) (*models.EventList, error) {
	return s.repo.GetFoundationEvents(foundationSFID, nextKey, paramPageSize, all, searchTerm)
}

// GetClaGroupEvents returns the list of project events
func (s *service) GetClaGroupEvents(projectSFDC string, nextKey *string, paramPageSize *int64, all bool, searchTerm *string) (*models.EventList, error) {
	return s.repo.GetClaGroupEvents(projectSFDC, nextKey, paramPageSize, all, searchTerm)
}

// GetCompanyFoundationEvents returns list of events for company and foundation
func (s *service) GetCompanyFoundationEvents(companySFID, companyID, foundationSFID string, nextKey *string, paramPageSize *int64, all bool) (*models.EventList, error) {
	return s.repo.GetCompanyFoundationEvents(companySFID, companyID, foundationSFID, nextKey, paramPageSize, all)
}

// GetCompanyClaGroupEvents returns list of events for company and cla group
func (s *service) GetCompanyClaGroupEvents(companySFID, companyID, claGroupID string, nextKey *string, paramPageSize *int64, all bool) (*models.EventList, error) {
	return s.repo.GetCompanyClaGroupEvents(companySFID, companyID, claGroupID, nextKey, paramPageSize, all)
}

func (s *service) GetCompanyEvents(companyID, eventType string, nextKey *string, paramPageSize *int64, all bool) (*models.EventList, error) {
	return s.repo.GetCompanyEvents(companyID, eventType, nextKey, paramPageSize, all)
}

// LogEventArgs is argument to LogEvent function
// EventType, EventData are compulsory.
// One of LfUsername, UserID must be present
type LogEventArgs struct {
	EventType string

	UserID     string
	LfUsername string
	UserName   string
	UserModel  *models.User

	CLAGroupID    string
	CLAGroupName  string
	ClaGroupModel *models.ClaGroup

	ProjectID         string // Should just use CLA GroupID
	ProjectSFID       string
	ProjectName       string
	ParentProjectSFID string
	ParentProjectName string

	CompanyID    string
	CompanyName  string
	CompanySFID  string
	CompanyModel *models.Company

	EventData EventData
}

func (s *service) loadCompany(ctx context.Context, args *LogEventArgs) error {
	f := logrus.Fields{
		"functionName":   "loadCompany",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
	}

	if args.CompanyModel != nil {
		args.CompanyName = args.CompanyModel.CompanyName
		args.CompanyID = args.CompanyModel.CompanyID
		args.CompanySFID = args.CompanyModel.CompanyExternalID
		return nil
	} else if args.CompanyID != "" {
		companyModel, err := s.combinedRepo.GetCompany(ctx, args.CompanyID)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("failed to load company record ID: %s", args.CompanyID)
			return err
		}
		args.CompanyModel = companyModel
		args.CompanyName = companyModel.CompanyName
		args.CompanySFID = companyModel.CompanyExternalID
	}

	return nil
}

func (s *service) loadCLAGroup(ctx context.Context, args *LogEventArgs) error {
	f := logrus.Fields{
		"functionName":   "loadCLAGroup",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
	}

	// First, attempt to user the CLA Group model that was provided...
	if args.ClaGroupModel != nil {
		args.CLAGroupID = args.ClaGroupModel.ProjectID
		args.ProjectName = args.ClaGroupModel.ProjectName
		args.ProjectSFID = args.ClaGroupModel.ProjectExternalID
	} else {
		// Did they set the CLA Group ID?
		claGroupID := ""
		if args.CLAGroupID != "" {
			claGroupID = args.CLAGroupID
		} else if args.ProjectID != "" && utils.IsUUIDv4(args.ProjectID) { // legacy parameter
			claGroupID = args.ProjectID
		}

		// Load the CLA Group ID if set...
		if claGroupID != "" {
			claGroupModel, err := s.combinedRepo.GetCLAGroupByID(ctx, claGroupID, DontLoadRepoDetails)
			if err != nil {
				log.WithFields(f).WithError(err).Warnf("failed to load CLA Group by ID: %s", claGroupID)
				return err
			}
			args.ClaGroupModel = claGroupModel
			args.CLAGroupName = claGroupModel.ProjectName
			args.ProjectSFID = claGroupModel.ProjectExternalID
		}
	}

	return nil
}

func (s *service) loadSFProject(ctx context.Context, args *LogEventArgs) error {
	f := logrus.Fields{
		"functionName":   "loadSFProject",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
	}

	if utils.IsUUIDv4(args.ProjectID) { // internal CLA Group ID
		return s.loadCLAGroup(ctx, args)
	} else if utils.IsSalesForceID(args.ProjectID) { // external SF project ID
		args.ProjectSFID = args.ProjectID
		// Check if project exists in platform project service
		log.WithFields(f).Debugf("loading salesforce project by ID: %s...", args.ProjectSFID)
		project, projectErr := project_service.GetClient().GetProject(args.ProjectSFID)
		if projectErr != nil || project == nil {
			log.WithFields(f).Warnf("failed to load salesforce project by ID: %s", args.ProjectSFID)
			return nil
		}
		log.WithFields(f).Debugf("loaded salesforce project by ID: %s", args.ProjectSFID)
		args.ProjectName = project.Name

		// Try to load and set the parent information
		if project.Parent != "" {
			log.WithFields(f).Debugf("loading salesforce project parent by ID: %s...", project.Parent)
			parentProject, parentProjectErr := project_service.GetClient().GetProject(project.Parent)
			if parentProjectErr != nil || parentProject == nil {
				log.WithFields(f).Warnf("failed to load salesforce project parent by ID: %s", project.Parent)
				return nil
			}
			log.WithFields(f).Debugf("loaded salesforce project by parent ID: %s", project.Parent)
			args.ParentProjectSFID = parentProject.ID
			args.ParentProjectName = parentProject.Name
		}

		return nil
	}

	return nil
}

func (s *service) loadUser(ctx context.Context, args *LogEventArgs) error {
	f := logrus.Fields{
		"functionName":   "loadUser",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
	}

	if args.UserModel != nil {
		args.UserName = args.UserModel.Username
		args.UserID = args.UserModel.UserID
		args.LfUsername = args.UserModel.LfUsername
		return nil
	}
	if args.UserID == "" && args.LfUsername == "" {
		log.WithFields(f).Warn("failed to load user for event log - user ID and username were not set")
		return errors.New("require userID or LfUsername")
	}
	var userModel *models.User
	var err error
	if args.LfUsername != "" {
		userModel, err = s.combinedRepo.GetUserByUserName(args.LfUsername, true)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("failed to load user by username: %s", args.LfUsername)
			return err
		}
	}
	if args.UserID != "" {
		userModel, err = s.combinedRepo.GetUser(args.UserID)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("failed to load user by ID: %s", args.UserID)
			return err
		}
	}

	if userModel != nil {
		args.UserModel = userModel
		args.UserName = userModel.Username
		args.UserID = userModel.UserID
		args.LfUsername = userModel.LfUsername
	} else {
		log.WithFields(f).Warnf("unable to set user information for event log entry")
	}

	return nil
}

// loadDetails fetches and sets additional information into the data model required to fill out the event log entry
func (s *service) loadDetails(ctx context.Context, args *LogEventArgs) error {
	err := s.loadCompany(ctx, args)
	if err != nil {
		return err
	}

	err = s.loadCLAGroup(ctx, args)
	if err != nil {
		return err
	}

	err = s.loadSFProject(ctx, args)
	if err != nil {
		return err
	}

	err = s.loadUser(ctx, args)
	if err != nil {
		return err
	}

	return nil
}

// LogEventWithContext logs the event in database
func (s *service) LogEventWithContext(ctx context.Context, args *LogEventArgs) {
	f := logrus.Fields{
		"functionName":   "events.service.LogEvent",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
	}

	defer func() {
		if r := recover(); r != nil {
			log.WithFields(f).Error("panic occurred in CreateEvent", fmt.Errorf("%v", r))
		}
	}()

	if args == nil || args.EventType == "" || args.EventData == nil || (args.UserID == "" && args.LfUsername == "") {
		log.WithFields(f).Warnf("invalid arguments to LogEvent, missing one or more required values. args %#v", args)
		return
	}

	err := s.loadDetails(ctx, args)
	if err != nil {
		log.WithFields(f).Error("unable to load details for event", err)
		return
	}

	eventData, containsPII := args.EventData.GetEventDetailsString(args)
	eventSummary, _ := args.EventData.GetEventSummaryString(args)
	event := models.Event{
		EventType: args.EventType,

		UserID:     args.UserID,
		UserName:   args.UserName,
		LfUsername: args.LfUsername,

		EventCLAGroupID:   args.CLAGroupID,
		EventCLAGroupName: args.ProjectName,

		EventCompanyID:   args.CompanyID,
		EventCompanySFID: args.CompanySFID,
		EventCompanyName: args.CompanyName,

		EventProjectID:         args.ProjectID,
		EventProjectSFID:       args.ProjectSFID,
		EventProjectName:       args.ProjectName,
		EventParentProjectSFID: args.ParentProjectSFID,
		EventParentProjectName: args.ParentProjectName,

		EventData:    eventData,
		EventSummary: eventSummary,

		ContainsPII: containsPII,
	}
	err = s.repo.CreateEvent(&event)
	if err != nil {
		log.WithFields(f).Error(fmt.Sprintf("unable to create event for args %#v", args), err)
	}
}

// LogEvent logs the event in database
func (s *service) LogEvent(args *LogEventArgs) {
	s.LogEventWithContext(utils.NewContext(), args)
}
