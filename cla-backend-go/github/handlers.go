// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/LF-Engineering/cla-monorepo/cla-backend-go/gen/restapi/operations"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gofrs/uuid"
	"github.com/savaki/dynastore"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const (
	SessionStoreKey = "cla-github"
)

func Configure(api *operations.ClaAPI, clientID, clientSecret string, sessionStore *dynastore.Store) {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes: []string{
			"read:org",
		},
		Endpoint: github.Endpoint,
	}

	api.GithubLoginHandler = operations.GithubLoginHandlerFunc(func(params operations.GithubLoginParams) middleware.Responder {
		return middleware.ResponderFunc(
			func(w http.ResponseWriter, pr runtime.Producer) {
				session, err := sessionStore.Get(params.HTTPRequest, SessionStoreKey)
				if err != nil {
					fmt.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// Store the callback url so we can redirect back to it once logged in.
				session.Values["callback"] = params.Callback

				// Generate a csrf token to send
				state, err := uuid.NewV4()
				if err != nil {
					fmt.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				session.Values["state"] = state.String()

				err = session.Save(params.HTTPRequest, w)
				if err != nil {
					fmt.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.Redirect(w, params.HTTPRequest, oauthConfig.AuthCodeURL(state.String()), http.StatusFound)
			})
	})

	api.GithubRedirectHandler = operations.GithubRedirectHandlerFunc(func(params operations.GithubRedirectParams) middleware.Responder {
		return middleware.ResponderFunc(
			func(w http.ResponseWriter, pr runtime.Producer) {
				// Verify csrf token
				session, err := sessionStore.Get(params.HTTPRequest, SessionStoreKey)
				if err != nil {
					fmt.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				persistedState, ok := session.Values["state"].(string)
				if !ok {
					fmt.Println("no session state")
					http.Error(w, "no session state", http.StatusInternalServerError)
					return
				}

				if params.State != persistedState {
					fmt.Println("mismatch state")
					http.Error(w, "mismatch state", http.StatusInternalServerError)
					return
				}

				// trade temporary code for access token
				token, err := oauthConfig.Exchange(context.TODO(), params.Code)
				if err != nil {
					fmt.Println("unable to exchange code")
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// persist access token
				session.Values["github_access_token"] = token.AccessToken

				err = session.Save(params.HTTPRequest, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				callback, ok := session.Values["callback"].(string)
				if !ok {
					fmt.Println("unable to find callback to redirect to")
					http.Error(w, "unable to find callback to redirect to", http.StatusInternalServerError)
					return
				}

				http.Redirect(w, params.HTTPRequest, callback, http.StatusFound)
			})
	})
}
