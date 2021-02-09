// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package emails

import (
	"testing"

	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestRemovedCLAManagerTemplate(t *testing.T) {
	params := RemovedCLAManagerTemplateParams{
		CLAManagerTemplateParams: CLAManagerTemplateParams{
			RecipientName:       "JohnsClaManager",
			ProjectName:         "JohnsProject",
			ExternalProjectName: "JohnsProjectExternal",
			CompanyName:         "JohnsCompany",
			CLAManagers: []ClaManagerInfoParams{
				{LfUsername: "LFUserName", Email: "LFEmail"},
			},
		},
	}

	result, err := RenderTemplate(utils.V1, RemovedCLAManagerTemplateName, RemovedCLAManagerTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")
	assert.Contains(t, result, "CLA Manager from JohnsCompany for the project JohnsProject")
	assert.Contains(t, result, "<li>LFUserName LFEmail</li>")

	// even if the foundation is set we should show the project name
	// because 0 child projects under the claGroup
	params.CLAManagerTemplateParams.FoundationName = "CNCF"
	result, err = RenderTemplate(utils.V1, RemovedCLAManagerTemplateName, RemovedCLAManagerTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")

	// then we increase the child project count so we should get the FoundationName instead of project name
	params.ChildProjectCount = 2
	result, err = RenderTemplate(utils.V1, RemovedCLAManagerTemplateName, RemovedCLAManagerTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project CNCF")
}

func TestRequestAccessToCLAManagersTemplate(t *testing.T) {
	params := RequestAccessToCLAManagersTemplateParams{
		CLAManagerTemplateParams: CLAManagerTemplateParams{
			RecipientName:       "JohnsClaManager",
			ProjectName:         "JohnsProject",
			ExternalProjectName: "JohnsProjectExternal",
			CompanyName:         "JohnsCompany",
		},
		RequesterName:  "RequesterName",
		RequesterEmail: "RequesterEmail",
		CorporateURL:   "http://CorporateURL.com",
	}

	result, err := RenderTemplate(utils.V1, RequestAccessToCLAManagersTemplateName, RequestAccessToCLAManagersTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")
	assert.Contains(t, result, "from JohnsCompany for the project JohnsProject")
	assert.Contains(t, result, "contribute to JohnsProject")
	assert.Contains(t, result, "CLA Managers for JohnsProject")
	assert.Contains(t, result, "RequesterName (RequesterEmail) has requested")
	assert.Contains(t, result, "another CLA Manager from JohnsCompany for JohnsProject")
	assert.Contains(t, result, "<a href=\"http://CorporateURL.com\" target=\"_blank\">")
	assert.Contains(t, result, "then select the JohnsProject project")

}

func TestRequestApprovedToCLAManagersTemplate(t *testing.T) {
	params := RequestApprovedToCLAManagersTemplateParams{
		CLAManagerTemplateParams: CLAManagerTemplateParams{
			RecipientName:       "JohnsClaManager",
			ProjectName:         "JohnsProject",
			ExternalProjectName: "JohnsProjectExternal",
			CompanyName:         "JohnsCompany",
		},
		RequesterName:  "RequesterName",
		RequesterEmail: "RequesterEmail",
	}

	result, err := RenderTemplate(utils.V1, RequestApprovedToCLAManagersTemplateName, RequestApprovedToCLAManagersTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")
	assert.Contains(t, result, "CLA Manager from JohnsCompany for the project JohnsProject")
	assert.Contains(t, result, "allowed to contribute to JohnsProject")
	assert.Contains(t, result, "CLA Managers for JohnsProject")
	assert.Contains(t, result, "<li>RequesterName (RequesterEmail)</li>")
}

func TestRequestApprovedToRequesterTemplateParams(t *testing.T) {
	params := RequestApprovedToRequesterTemplateParams{
		CLAManagerTemplateParams: CLAManagerTemplateParams{
			RecipientName:       "JohnsClaManager",
			ProjectName:         "JohnsProject",
			ExternalProjectName: "JohnsProjectExternal",
			CompanyName:         "JohnsCompany",
		},
		CorporateURL: "http://CorporateURL.com",
	}

	result, err := RenderTemplate(utils.V1, RequestApprovedToRequesterTemplateName, RequestApprovedToRequesterTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")
	assert.Contains(t, result, "CLA Manager from JohnsCompany for the project JohnsProject")
	assert.Contains(t, result, "allowed to contribute to JohnsProject")
	assert.Contains(t, result, "CLA Managers for JohnsProject")
	assert.Contains(t, result, "<a href=\"http://CorporateURL.com\" target=\"_blank\">")
	assert.Contains(t, result, "and then the project JohnsProject")
}

func TestRequestDeniedToCLAManagersTemplate(t *testing.T) {
	params := RequestDeniedToCLAManagersTemplateParams{
		CLAManagerTemplateParams: CLAManagerTemplateParams{
			RecipientName:       "JohnsClaManager",
			ProjectName:         "JohnsProject",
			ExternalProjectName: "JohnsProjectExternal",
			CompanyName:         "JohnsCompany",
		},
		RequesterName:  "RequesterName",
		RequesterEmail: "RequesterEmail",
	}

	result, err := RenderTemplate(utils.V1, RequestDeniedToCLAManagersTemplateName, RequestDeniedToCLAManagersTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")
	assert.Contains(t, result, "CLA Manager from JohnsCompany for the project JohnsProject")
	assert.Contains(t, result, "allowed to contribute to JohnsProject")
	assert.Contains(t, result, "<li>RequesterName (RequesterEmail)</li>")
}

func TestRequestDeniedToRequesterTemplate(t *testing.T) {
	params := RequestDeniedToRequesterTemplateParams{
		CLAManagerTemplateParams: CLAManagerTemplateParams{
			RecipientName:       "JohnsClaManager",
			ProjectName:         "JohnsProject",
			ExternalProjectName: "JohnsProjectExternal",
			CompanyName:         "JohnsCompany",
		},
	}

	result, err := RenderTemplate(utils.V1, RequestDeniedToRequesterTemplateName, RequestDeniedToRequesterTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")
	assert.Contains(t, result, "CLA Manager from JohnsCompany for the project JohnsProject")
	assert.Contains(t, result, "allowed to contribute to JohnsProject")
}

func TestClaManagerAddedEToUserTemplate(t *testing.T) {
	params := ClaManagerAddedEToUserTemplateParams{
		CLAManagerTemplateParams: CLAManagerTemplateParams{
			RecipientName:       "JohnsClaManager",
			ProjectName:         "JohnsProject",
			ExternalProjectName: "JohnsProjectExternal",
			CompanyName:         "JohnsCompany",
		},
		CorporateURL: "http://CorporateURL.com",
	}

	result, err := RenderTemplate(utils.V1, ClaManagerAddedEToUserTemplateName, ClaManagerAddedEToUserTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")
	assert.Contains(t, result, "CLA Manager from JohnsCompany for the project JohnsProject")
	assert.Contains(t, result, "allowed to contribute to JohnsProject")
	assert.Contains(t, result, "CLA Managers for JohnsProject")
	assert.Contains(t, result, "<a href=\"http://CorporateURL.com\" target=\"_blank\">")
	assert.Contains(t, result, "and then the project JohnsProject")
}

func TestClaManagerAddedToCLAManagersTemplate(t *testing.T) {
	params := ClaManagerAddedToCLAManagersTemplateParams{
		CLAManagerTemplateParams: CLAManagerTemplateParams{
			RecipientName:       "JohnsClaManager",
			ProjectName:         "JohnsProject",
			ExternalProjectName: "JohnsProjectExternal",
			CompanyName:         "JohnsCompany",
		},
		Name:  "John",
		Email: "john@example.com",
	}

	result, err := RenderTemplate(utils.V1, ClaManagerAddedToCLAManagersTemplateName, ClaManagerAddedToCLAManagersTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")
	assert.Contains(t, result, "CLA Manager from JohnsCompany for the project JohnsProject")
	assert.Contains(t, result, "contribute to JohnsProject")
	assert.Contains(t, result, "CLA Managers for JohnsProject")
	assert.Contains(t, result, "<li>John (john@example.com)</li>")

}

func TestClaManagerDeletedToCLAManagersTemplate(t *testing.T) {
	params := ClaManagerDeletedToCLAManagersTemplateParams{
		CLAManagerTemplateParams: CLAManagerTemplateParams{
			RecipientName:       "JohnsClaManager",
			ProjectName:         "JohnsProject",
			ExternalProjectName: "JohnsProjectExternal",
			CompanyName:         "JohnsCompany",
		},
		Name:  "John",
		Email: "john@example.com",
	}

	result, err := RenderTemplate(utils.V1, ClaManagerDeletedToCLAManagersTemplateName, ClaManagerDeletedToCLAManagersTemplate,
		params)
	assert.NoError(t, err)
	assert.Contains(t, result, "Hello JohnsClaManager")
	assert.Contains(t, result, "regarding the project JohnsProject")
	assert.Contains(t, result, "John (john@example.com) has been removed")

}
