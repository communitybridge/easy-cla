// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package tests

import (
	"testing"

	"github.com/communitybridge/easycla/cla-backend-go/emails"

	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestV2ContributorApprovalRequestTemplate(t *testing.T) {
	params := emails.V2ContributorApprovalRequestTemplateParams{
		CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
			RecipientName: "JohnsClaManager",
			Project:       emails.CLAProjectParams{ExternalProjectName: "JohnsProject"},
			CLAGroupName:  "JohnsCLAGroupName",
			CompanyName:   "JohnsCompany",
		},
		UserDetails:           "UserDetailsValue",
		CorporateConsoleV2URL: "http://CorporateConsoleV2URL.com",
	}

	result, err := emails.RenderTemplate(utils.V1, emails.V2ContributorApprovalRequestTemplateName, emails.V2ContributorApprovalRequestTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the organization JohnsCompany")
	assert.Contains(t, result, "contribution to the CLA Group JohnsCLAGroupName")
	assert.Contains(t, result, "UserDetailsValue")
	assert.Contains(t, result, "Approval can be done at http://CorporateConsoleV2URL.com")

	params.SigningEntityName = "SigningEntityNameValue"

	result, err = emails.RenderTemplate(utils.V1, emails.V2ContributorApprovalRequestTemplateName, emails.V2ContributorApprovalRequestTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the organization JohnsCompany")
	assert.Contains(t, result, "contribution to the CLA Group JohnsCLAGroupName")
	assert.Contains(t, result, "UserDetailsValue")
	assert.Contains(t, result, "Approval can be done at http://CorporateConsoleV2URL.com")
}

func TestV2OrgAdminTemplate(t *testing.T) {
	params := emails.V2OrgAdminTemplateParams{
		CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
			RecipientName: "JohnsClaManager",
			Project: emails.CLAProjectParams{
				ExternalProjectName: "JohnsProject",
				ProjectSFID:         "ProjectSFIDValue",
				FoundationSFID:      "FoundationSFIDValue",
				CorporateConsole:    "http://CorporateConsole.com",
			},
			CLAGroupName: "JohnsCLAGroupName",
			CompanyName:  "JohnsCompany",
		},
		SenderName:       "SenderNameValue",
		SenderEmail:      "SenderEmailValue",
		CorporateConsole: "http://CorporateConsole.com",
	}

	result, err := emails.RenderTemplate(utils.V1, emails.V2OrgAdminTemplateName, emails.V2OrgAdminTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "signing process for the organization JohnsCompany")
	assert.Contains(t, result, "SenderNameValue SenderEmailValue has identified you")
	assert.Contains(t, result, "Corporate CLA in support of the following project(s):")
	assert.Contains(t, result, "<li>JohnsProject</li>")
	assert.Contains(t, result, "can login to this portal (http://CorporateConsole.com)")
	assert.Contains(t, result, `sign the CLA for this project <a href="http://CorporateConsole.com/foundation/FoundationSFIDValue/project/ProjectSFIDValue/cla" target="_blank">JohnsProject</a>`)
}

func TestV2ContributorToOrgAdminTemplate(t *testing.T) {
	params := emails.V2ContributorToOrgAdminTemplateParams{
		CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
			RecipientName: "JohnsClaManager",
			Project:       emails.CLAProjectParams{ExternalProjectName: "JohnsProject"},
			CLAGroupName:  "JohnsCLAGroupName",
			CompanyName:   "JohnsCompany",
		},
		Projects: []emails.CLAProjectParams{
			{ExternalProjectName: "Project1", ProjectSFID: "ProjectSFID1", FoundationSFID: "FoundationSFID1", CorporateConsole: "http://CorporateConsole.com"},
			{ExternalProjectName: "Project2", ProjectSFID: "ProjectSFID2", FoundationSFID: "FoundationSFID2", CorporateConsole: "http://CorporateConsole.com"},
		},
		UserDetails:      "UserDetailsValue",
		CorporateConsole: "http://CorporateConsole.com",
	}

	result, err := emails.RenderTemplate(utils.V1, emails.V2ContributorToOrgAdminTemplateName, emails.V2ContributorToOrgAdminTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project(s) Project1,Project2")
	assert.Contains(t, result, "sign the CLA for the organization: JohnsCompany")
	assert.Contains(t, result, "<p>UserDetailsValue</p>")
	assert.Contains(t, result, "Kindly login to this portal http://CorporateConsole.com")
	assert.Contains(t, result, `CLA for any of the project(s): <a href="http://CorporateConsole.com/foundation/FoundationSFID1/project/ProjectSFID1/cla" target="_blank">Project1</a>,<a href="http://CorporateConsole.com/foundation/FoundationSFID2/project/ProjectSFID2/cla" target="_blank">Project2</a>`)
}

func TestV2CLAManagerDesigneeCorporateTemplate(t *testing.T) {
	params := emails.V2CLAManagerDesigneeCorporateTemplateParams{
		CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
			RecipientName: "JohnsClaManager",
			Project: emails.CLAProjectParams{
				ExternalProjectName: "JohnsProject",
				FoundationSFID:      "FoundationSFIDValue",
				ProjectSFID:         "ProjectSFIDValue",
				CorporateConsole:    "http://CorporateConsole.com",
			},
			CLAGroupName: "JohnsCLAGroupName",
			CompanyName:  "JohnsCompany",
		},
		SenderName:       "SenderNameValue",
		SenderEmail:      "SenderEmailValue",
		CorporateConsole: "http://CorporateConsole.com",
	}

	result, err := emails.RenderTemplate(utils.V1, emails.V2CLAManagerDesigneeCorporateTemplateName, emails.V2CLAManagerDesigneeCorporateTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "CLA setup and signing process for the organization JohnsCompany")
	assert.Contains(t, result, "SenderNameValue SenderEmailValue has identified you")
	assert.Contains(t, result, "Corporate CLA for the organization JohnsCompany")
	assert.Contains(t, result, "<li>JohnsProject</li>")
	assert.Contains(t, result, "can login to this portal (http://CorporateConsole.com)")
	assert.Contains(t, result, `sign the CLA for this project <a href="http://CorporateConsole.com/foundation/FoundationSFIDValue/project/ProjectSFIDValue/cla" target="_blank">JohnsProject</a>`)
}

func TestV2ToCLAManagerDesigneeTemplate(t *testing.T) {
	params := emails.V2ToCLAManagerDesigneeTemplateParams{
		RecipientName: "JohnsClaManager",
		Projects: []emails.CLAProjectParams{
			{ExternalProjectName: "Project1", ProjectSFID: "ProjectSFID1", FoundationSFID: "FoundationSFID1", CorporateConsole: "http://CorporateConsole.com"},
			{ExternalProjectName: "Project2", ProjectSFID: "ProjectSFID2", FoundationSFID: "FoundationSFID2", CorporateConsole: "http://CorporateConsole.com"},
		},
		ContributorID:    "ContributorIDValue",
		ContributorName:  "ContributorNameValue",
		CorporateConsole: "http://CorporateConsole.com",
	}

	result, err := emails.RenderTemplate(utils.V1, emails.V2ToCLAManagerDesigneeTemplateName, emails.V2ToCLAManagerDesigneeTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project(s): Project1, Project2")
	assert.Contains(t, result, "<p> ContributorIDValue (ContributorNameValue) </p>")
	assert.Contains(t, result, "Kindly login to this portal http://CorporateConsole.com")
	assert.Contains(t, result, `CLA for one of the project(s): <a href="http://CorporateConsole.com/foundation/FoundationSFID1/project/ProjectSFID1/cla" target="_blank">Project1</a>,<a href="http://CorporateConsole.com/foundation/FoundationSFID2/project/ProjectSFID2/cla" target="_blank">Project2</a>`)

	params.Projects = []emails.CLAProjectParams{
		{ExternalProjectName: "Project1", ProjectSFID: "ProjectSFID1", FoundationSFID: "FoundationSFID1", CorporateConsole: "http://CorporateConsole.com"},
	}
	result, err = emails.RenderTemplate(utils.V1, emails.V2ToCLAManagerDesigneeTemplateName, emails.V2ToCLAManagerDesigneeTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project(s): Project1")
	assert.Contains(t, result, "<p> ContributorIDValue (ContributorNameValue) </p>")
	assert.Contains(t, result, "Kindly login to this portal http://CorporateConsole.com")
	assert.Contains(t, result, `CLA for one of the project(s): <a href="http://CorporateConsole.com/foundation/FoundationSFID1/project/ProjectSFID1/cla" target="_blank">Project1</a>`)

}

func TestV2DesigneeToUserWithNoLFIDTemplate(t *testing.T) {
	params := emails.V2DesigneeToUserWithNoLFIDTemplateParams{
		CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
			RecipientName: "JohnsClaManager",
			Project: emails.CLAProjectParams{
				ExternalProjectName:     "JohnsProjectExternal",
				CorporateConsole:        "https://corporate.dev.lfcla.com",
				FoundationSFID:          "FoundationSFIDValue",
				SignedAtFoundationLevel: true,
			},
			CLAGroupName: "JohnsCLAGroupName",
			CompanyName:  "JohnsCompany",
		},
		RequesterUserName: "RequesterUserNameValue",
		RequesterEmail:    "RequesterEmailValue",
		CorporateConsole:  "https://corporate.dev.lfcla.com",
	}

	result, err := emails.RenderTemplate(utils.V1, emails.V2DesigneeToUserWithNoLFIDTemplateName, emails.V2DesigneeToUserWithNoLFIDTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager,")
	assert.Contains(t, result, "The following contributor would like to contribute to JohnsProjectExternal on behalf of your organization: JohnsCompany.")
	assert.Contains(t, result, "you will be redirected to this portal https://corporate.dev.lfcla.com ")
	assert.Contains(t, result, `where you can sign the CLA for the project <a href="https://corporate.dev.lfcla.com/foundation/FoundationSFIDValue/cla" target="_blank">JohnsProjectExternal</a>`)
}

func TestV2CLAManagerToUserWithNoLFIDTemplate(t *testing.T) {
	params := emails.V2CLAManagerToUserWithNoLFIDTemplateParams{
		CLAManagerTemplateParams: emails.CLAManagerTemplateParams{
			RecipientName: "JohnsClaManager",
			Project:       emails.CLAProjectParams{ExternalProjectName: "JohnsProjectExternal"},
			CLAGroupName:  "JohnsCLAGroupName",
			CompanyName:   "JohnsCompany",
		},
		RequesterUserName: "RequesterUserNameValue",
		RequesterEmail:    "RequesterEmailValue",
	}

	result, err := emails.RenderTemplate(utils.V1, emails.V2CLAManagerToUserWithNoLFIDTemplateName, emails.V2CLAManagerToUserWithNoLFIDTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the Project JohnsProjectExternal and CLA Group JohnsCLAGroupName")
	assert.Contains(t, result, "User RequesterUserNameValue (RequesterEmailValue) was trying")
	assert.Contains(t, result, "CLA Manager for the Project JohnsProject")
	assert.Contains(t, result, "notify the user RequesterUserNameValue")
}
