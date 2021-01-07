// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package events

import "github.com/communitybridge/easycla/cla-backend-go/gen/models"

//IndividualSignedEvent represntative of ICLA signatures
const IndividualSignedEvent = "IndividualSignatureSigned"

// Event data model
type Event struct {
	EventID                string `dynamodbav:"event_id"`
	EventType              string `dynamodbav:"event_type"`
	EventUserID            string `dynamodbav:"event_user_id"`
	EventUserName          string `dynamodbav:"event_user_name"`
	EventLfUsername        string `dynamodbav:"event_lf_username"`
	EventProjectID         string `dynamodbav:"event_project_id"`
	EventProjectExternalID string `dynamodbav:"event_project_external_id"`
	EventProjectName       string `dynamodbav:"event_project_name"`
	EventCompanyID         string `dynamodbav:"event_company_id"`
	EventCompanyName       string `dynamodbav:"event_company_name"`
	EventTime              string `dynamodbav:"event_time"`
	EventTimeEpoch         int64  `dynamodbav:"event_time_epoch"`
	EventData              string `dynamodbav:"event_data"`
	EventSummary           string `dynamodbav:"event_summary"`
	EventFoundationSFID    string `dynamodbav:"event_foundation_sfid"`
	EventSFProjectName     string `dynamodbav:"event_sf_project_name"`
	EventProjectSFID       string `dynamodbav:"event_project_sfid"`
	EventCompanySFID       string `dynamodbav:"event_company_sfid"`
}

// DBUser data model
type DBUser struct {
	UserID             string   `json:"user_id"`
	UserExternalID     string   `json:"user_external_id"`
	LFEmail            string   `json:"lf_email"`
	Admin              bool     `json:"admin"`
	LFUsername         string   `json:"lf_username"`
	DateCreated        string   `json:"date_created"`
	DateModified       string   `json:"date_modified"`
	UserName           string   `json:"user_name"`
	Version            string   `json:"version"`
	UserEmails         []string `json:"user_emails"`
	UserGithubID       string   `json:"user_github_id"`
	UserCompanyID      string   `json:"user_company_id"`
	UserGithubUsername string   `json:"user_github_username"`
	Note               string   `json:"note"`
}

func (e *Event) toEvent() *models.Event { //nolint
	event := &models.Event{
		EventData:              e.EventData,
		EventSummary:           e.EventSummary,
		EventID:                e.EventID,
		EventProjectID:         e.EventProjectID,
		EventProjectExternalID: e.EventProjectExternalID,
		EventProjectName:       e.EventProjectName,
		EventTime:              e.EventTime,
		EventType:              e.EventType,
		UserID:                 e.EventUserID,
		UserName:               e.EventUserName,
		LfUsername:             e.EventLfUsername,
		EventTimeEpoch:         e.EventTimeEpoch,
		EventFoundationSFID:    e.EventFoundationSFID,
		EventProjectSFID:       e.EventProjectSFID,
		EventProjectSFName:     e.EventSFProjectName,
		EventCompanySFID:       e.EventCompanySFID,
	}
	// Disregard Company details for ICLA event
	if event.EventType != IndividualSignedEvent {
		event.EventCompanyID = e.EventCompanyID
		event.EventCompanyName = e.EventCompanyName
	}

	return event
}

// DBProjectModel data model
type DBProjectModel struct {
	DateCreated                      string                   `dynamodbav:"date_created"`
	DateModified                     string                   `dynamodbav:"date_modified"`
	ProjectExternalID                string                   `dynamodbav:"project_external_id"`
	ProjectID                        string                   `dynamodbav:"project_id"`
	ProjectName                      string                   `dynamodbav:"project_name"`
	Version                          string                   `dynamodbav:"version"`
	ProjectCclaEnabled               bool                     `dynamodbav:"project_ccla_enabled"`
	ProjectCclaRequiresIclaSignature bool                     `dynamodbav:"project_ccla_requires_icla_signature"`
	ProjectIclaEnabled               bool                     `dynamodbav:"project_icla_enabled"`
	ProjectCorporateDocuments        []DBProjectDocumentModel `dynamodbav:"project_corporate_documents"`
	ProjectIndividualDocuments       []DBProjectDocumentModel `dynamodbav:"project_individual_documents"`
	ProjectMemberDocuments           []DBProjectDocumentModel `dynamodbav:"project_member_documents"`
	ProjectACL                       []string                 `dynamodbav:"project_acl"`
}

// DBProjectDocumentModel is a data model for the CLA Group Project documents
type DBProjectDocumentModel struct {
	DocumentName            string `dynamodbav:"document_name"`
	DocumentFileID          string `dynamodbav:"document_file_id"`
	DocumentPreamble        string `dynamodbav:"document_preamble"`
	DocumentLegalEntityName string `dynamodbav:"document_legal_entity_name"`
	DocumentAuthorName      string `dynamodbav:"document_author_name"`
	DocumentContentType     string `dynamodbav:"document_content_type"`
	DocumentS3URL           string `dynamodbav:"document_s3_url"`
	DocumentMajorVersion    string `dynamodbav:"document_major_version"`
	DocumentMinorVersion    string `dynamodbav:"document_minor_version"`
	DocumentCreationDate    string `dynamodbav:"document_creation_date"`
}
