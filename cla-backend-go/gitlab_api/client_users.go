// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package gitlab

import (
	"context"
	"errors"
	"fmt"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/sirupsen/logrus"
	goGitLab "github.com/xanzy/go-gitlab"
)

// GetUserByName gets a gitlab user object by the given name
func GetUserByName(ctx context.Context, client GitLabClient, name string) (*goGitLab.User, error) {
	f := logrus.Fields{
		"functionName":   "gitlab_api.client_users.GetUserByName",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"name":           name,
	}

	users, err := client.ListUsers(&goGitLab.ListUsersOptions{
		ListOptions: goGitLab.ListOptions{
			Page:    0,
			PerPage: 10,
		},
		Username: utils.StringRef(name),
	})

	if err != nil {
		msg := fmt.Sprintf("problem fetching users, error: %+v", err)
		log.WithFields(f).WithError(err).Warn(msg)
		return nil, errors.New(msg)
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}
