// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package gitlab

import (
	"errors"
	"fmt"

	"github.com/communitybridge/easycla/cla-backend-go/config"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// FetchOauthCredentials is responsible for fetching the credentials from gitlab for alredy started Oauth process (access_token, refresh_token)
func FetchOauthCredentials(code string) (*OauthSuccessResponse, error) {
	gitLabConfig := config.GetConfig().Gitlab
	f := logrus.Fields{
		"functionName": "gitlab.auth.FetchOauthCredentials",
		"code":         code,
		"redirectURI":  config.GetConfig().Gitlab.RedirectURI,
	}

	if len(gitLabConfig.AppID) > 4 {
		f["gitLabClientID"] = fmt.Sprintf("%s...%s", gitLabConfig.AppID[0:4], gitLabConfig.AppID[len(gitLabConfig.AppID)-4:])
	} else {
		return nil, errors.New("gitlab application client ID value is not set - value is empty or malformed")
	}
	if len(gitLabConfig.ClientSecret) > 4 {
		f["gitLabClientSecret"] = fmt.Sprintf("%s...%s", gitLabConfig.ClientSecret[0:4], gitLabConfig.ClientSecret[len(gitLabConfig.ClientSecret)-4:])
	} else {
		return nil, errors.New("gitlab application client secret value is not set - value is empty or malformed")
	}

	// For info on this authorization flow, see: https://docs.gitlab.com/ee/api/oauth2.html#authorization-code-flow
	client := resty.New()
	params := map[string]string{
		"client_id":     gitLabConfig.AppID,
		"client_secret": gitLabConfig.ClientSecret,
		"code":          code,
		"grant_type":    "authorization_code",
		"redirect_uri":  gitLabConfig.RedirectURI,
		//"redirect_uri": "http://localhost:8080/v4/gitlab/oauth/callback",
	}

	url := "https://gitlab.com/oauth/token"
	resp, err := client.R().
		SetQueryParams(params).
		SetResult(&OauthSuccessResponse{}).
		Post(url)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem invoking GitLab auth token exchange to: %s", url)
		return nil, err
	}

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		msg := fmt.Sprintf("problem invoking GitLab auth token exchange to: %s with status code: %d, response: %s", url, resp.StatusCode(), string(resp.Body()))
		log.WithFields(f).Warn(msg)
		return nil, errors.New(msg)
	}

	return resp.Result().(*OauthSuccessResponse), nil
}
