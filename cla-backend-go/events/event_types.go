// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package events

// event types
const (

	// CreateTemplate event type
	CreateTemplate = "Create Template"

	// AddGithubOrgToWL event type
	AddGithubOrgToWL = "Add GH Org To WL"
	// DeleteGithubOrgFromWL event type
	DeleteGithubOrgFromWL = "Delete GH Org From WL"

	// CreateCCLAWhitelistRequest event type
	CreateCCLAWhitelistRequest = "Create CCLA WL Request"
	// DeleteCCLAWhitelistRequest event type
	DeleteCCLAWhitelistRequest = "Delete CCLA WL Request"

	// AddUserToCompanyACL event type
	// DeleteUserFromCompanyACL event type
	//DeleteUserFromCompanyACL = "Delete User From Company ACL"

	// DeletePendingInvite event type
	AddGithubOrganization    = "Add Github Organization"
	DeleteGithubOrganization = "Delete Github Organization"
)

// events
// naming convention : <resource>.<action>
const (
	CLATemplateCreated = "cla_template.created"
	UserCreated        = "user.created"
	UserUpdated        = "user.updated"
	UserDeleted        = "user.deleted"

	GithubRepositoryAdded   = "github_repository.added"
	GithubRepositoryDeleted = "github_repository.deleted"

	GerritRepositoryAdded   = "gerrit_repository.added"
	GerritRepositoryDeleted = "gerrit_repository.deleted"

	GithubOrganizationAdded   = "github_organization.added"
	GithubOrganizationDeleted = "github_organization.deleted"

	CompanyACLUserAdded       = "company_acl.user_added"
	CompanyACLRequestAdded    = "company_acl.request_added"
	CompanyACLRequestApproved = "company_acl.request_approved"
	CompanyACLRequestDenied   = "company_acl.request_denied"

	CCLAWhitelistRequestCreated  = "ccla_whitelist_request.created"
	CCLAWhitelistRequestApproved = "ccla_whitelist_request.approved"
	CCLAWhitelistRequestRejected = "ccla_whitelist_request.rejected"

	WhitelistGithubOrganizationAdded   = "whitelist.github_organization_added"
	WhitelistGithubOrganizationDeleted = "whitelist.github_organization_deleted"

	ClaManagerAccessRequestCreated  = "cla_manager.access_request_created"
	ClaManagerAccessRequestApproved = "cla_manager.access_request_approved"
	ClaManagerAccessRequestDenied   = "cla_manager.access_request_denied"

	ProjectCreated = "project.created"
	ProjectUpdated = "project.updated"
	ProjectDeleted = "project.deleted"

	InvalidatedSignature = "signature.invalidated"
)
