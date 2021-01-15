// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/sirupsen/logrus"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/google/go-github/v33/github"
)

// errors
var (
	ErrGithubOrganizationNotFound = errors.New("github organization name not found")
)

// GetOrganization gets github organization
func GetOrganization(ctx context.Context, organizationName string) (*github.Organization, error) {
	f := logrus.Fields{
		"functionName":     "GetOrganization",
		utils.XREQUESTID:   ctx.Value(utils.XREQUESTID),
		"organizationName": organizationName,
	}

	client := NewGithubOauthClient()
	org, resp, err := client.Organizations.Get(ctx, organizationName)
	if err != nil {
		log.WithFields(f).Warnf("GetOrganization %s failed. error = %s", organizationName, err.Error())
		if resp != nil && resp.StatusCode == 404 {
			return nil, ErrGithubOrganizationNotFound
		}
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := fmt.Sprintf("GetOrganization %s failed with no success response code %d. error = %s", organizationName, resp.StatusCode, err.Error())
		log.WithFields(f).Warnf(msg)
		return nil, errors.New(msg)
	}
	return org, nil
}
