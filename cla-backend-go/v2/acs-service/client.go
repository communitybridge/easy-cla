// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package acs_service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/client/role"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/client/object_type"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/sirupsen/logrus"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/token"

	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/client"
	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/client/invite"
	"github.com/communitybridge/easycla/cla-backend-go/v2/acs-service/models"
	runtimeClient "github.com/go-openapi/runtime/client"

	"errors"

	"github.com/go-openapi/strfmt"
)

// Client is client for acs_service
type Client struct {
	apiKey   string
	apiGwURL string
	cl       *client.CentralAuthorizationLayerForTheLFXPlatform
}

var (
	acsServiceClient *Client
)

// errors
var (
	ErrRoleNotFound     = errors.New("role not found")
	ErrProjectIDMissing = errors.New("project ID missing")
)

// InitClient initializes the acs_service client
func InitClient(APIGwURL string, apiKey string) {
	url := strings.ReplaceAll(APIGwURL, "https://", "")
	acsServiceClient = &Client{
		apiKey:   apiKey,
		apiGwURL: APIGwURL,
		cl: client.NewHTTPClientWithConfig(strfmt.Default, &client.TransportConfig{
			Host:     url,
			BasePath: "acs/v1/api",
			Schemes:  []string{"https"},
		}),
	}
}

// GetClient return user_service client
func GetClient() *Client {
	return acsServiceClient
}

// SendUserInvite invites users to the LFX platform
func (ac *Client) SendUserInvite(ctx context.Context, email *string,
	roleName string, scope string, projectID *string, organizationID string, inviteType string, subject *string, emailContent *string, automate bool) error {
	f := logrus.Fields{
		"functionName":   "SendUserInvite",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"roleName":       roleName,
		"scope":          scope,
		"projectID":      utils.StringValue(projectID),
		"organizationID": organizationID,
		"inviteType":     inviteType,
		"subject":        utils.StringValue(subject),
		"automate":       automate,
	}

	tok, err := token.GetToken()
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem obtaining token, error: %+v", err)
		return err
	}

	clientAuth := runtimeClient.BearerToken(tok)
	params := &invite.CreateUserInviteParams{
		SendInvite: &models.CreateInvite{
			Automate: automate,
			Email:    email,
			Scope:    scope,
			RoleName: roleName,
			Type:     inviteType,
		},
		Context: ctx,
	}
	if scope == utils.ProjectOrgScope && projectID == nil {
		log.WithFields(f).Warnf("Project ID required for project|organization scope, error: %+v", ErrProjectIDMissing)
		return ErrProjectIDMissing
	}
	if scope == utils.ProjectOrgScope {
		// Set project|organization scope
		params.SendInvite.ScopeID = fmt.Sprintf("%s|%s", *projectID, organizationID)
	} else {
		params.SendInvite.ScopeID = organizationID
	}
	if subject != nil {
		params.SendInvite.Subject = *subject
	}
	// Pass emailContent if passed in the args
	if emailContent != nil {
		params.SendInvite.Body = *emailContent
	}

	log.WithFields(f).Debugf("Submitting ACS Service CreateUserInvite with payload: %+v", params)
	result, inviteErr := ac.cl.Invite.CreateUserInvite(params, clientAuth)
	if inviteErr != nil {
		log.WithFields(f).WithError(inviteErr).Warnf("ACS Service CreateUserInvite failed with payload: %+v", params)
		return nil
	}

	log.WithFields(f).Debugf("Successfully submitted ACS Service CreatedUserInvite, response: %+v",
		result.Payload)
	return nil
}

// GetRoleID will return roleID for the provided role name
func (ac *Client) GetRoleID(roleName string) (string, error) {
	f := logrus.Fields{
		"functionName": "GetRoleID",
		"roleName":     roleName,
	}

	tok, err := token.GetToken()
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem obtaining token, error: %+v", err)
		return "", err
	}

	rolesParams := &role.GetRolesParams{
		Search:  aws.String(roleName),
		Context: context.Background(),
	}
	clientAuth := runtimeClient.BearerToken(tok)
	response, err := ac.cl.Role.GetRoles(rolesParams, clientAuth)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem fetching GetRole, error: %+v", err)
		return "", err
	}

	for _, theRole := range response.Payload {
		if theRole.RoleName == roleName {
			return theRole.RoleID, nil
		}
	}

	return "", ErrRoleNotFound
}

// GetObjectTypeIDByName will return object type ID for the provided role name
func (ac *Client) GetObjectTypeIDByName(objectType string) (int, error) {
	f := logrus.Fields{
		"functionName": "GetObjectTypeID",
		"objectType":   objectType,
	}

	tok, err := token.GetToken()
	if err != nil {
		log.WithFields(f).Warnf("problem obtaining token, error: %+v", err)
		return 0, err
	}

	objectTypeListParams := &object_type.GetObjectTypeListParams{
		Context: context.Background(),
	}
	clientAuth := runtimeClient.BearerToken(tok)
	response, err := ac.cl.ObjectType.GetObjectTypeList(objectTypeListParams, clientAuth)
	if err != nil {
		log.WithFields(f).Warnf("problem fetching GetObjectTypeList, error: %+v", err)
		return 0, err
	}

	for _, entry := range response.Payload {
		log.WithFields(f).Debugf("Checking entry with name: %s against input: %s", entry.Name, objectType)
		if entry.Name == objectType {
			log.WithFields(f).Debugf("Found match: %s == %s, entry.TypeID: %d", entry.Name, objectType, int(entry.TypeID))
			return int(entry.TypeID), nil
		}
	}

	return 0, ErrRoleNotFound
}

// GetAssignedRoles will return assigned roles based on the roleName, project and organization SFID
func (ac *Client) GetAssignedRoles(roleName, projectSFID, organizationSFID string) (*models.ObjectRoleScope, error) {
	f := logrus.Fields{
		"functionName":     "GetAssignedRole",
		"roleName":         roleName,
		"projectSFID":      projectSFID,
		"organizationSFID": organizationSFID,
	}

	tok, err := token.GetToken()
	if err != nil {
		log.WithFields(f).Warnf("problem obtaining token, error: %+v", err)
		return nil, err
	}

	// Lookup the Project|Organization type
	objectTypeID, err := ac.GetObjectTypeIDByName(utils.ProjectOrgScope)
	if err != nil {
		log.WithFields(f).Warnf("problem obtaining token, error: %+v", err)
		return nil, err
	}

	objectID := fmt.Sprintf("%s|%s", projectSFID, organizationSFID)
	objectTypeParams := &object_type.GetObjectTypeRoleListParams{
		ID:       strconv.Itoa(objectTypeID),
		Objectid: aws.String(objectID),
		Context:  context.Background(),
	}
	clientAuth := runtimeClient.BearerToken(tok)
	log.WithFields(f).Debugf("querying for object type role list: %s with %s", strconv.Itoa(objectTypeID), objectID)
	response, err := ac.cl.ObjectType.GetObjectTypeRoleList(objectTypeParams, clientAuth)
	if err != nil {
		log.WithFields(f).Warnf("problem fetching GetObjectTypeRoleList, error: %+v", err)
		return nil, err
	}

	return response.Payload, nil
}

// DeleteRoleByID will delete the specified role by ID
func (ac *Client) DeleteRoleByID(roleID string) error {
	f := logrus.Fields{
		"functionName": "DeleteRoleByID",
		"roleID":       roleID,
	}

	if roleID == "" {
		log.WithFields(f).Warn("unable to delete role by ID - role ID is empty")
		return errors.New("empty role ID")
	}

	tok, err := token.GetToken()
	if err != nil {
		log.WithFields(f).Warnf("problem obtaining token, error: %+v", err)
		return err
	}

	roleParams := &role.DeleteRoleParams{
		ID:      strfmt.UUID(roleID),
		Context: context.Background(),
	}
	clientAuth := runtimeClient.BearerToken(tok)
	_, err = ac.cl.Role.DeleteRole(roleParams, clientAuth) // nolint
	if err != nil {
		log.WithFields(f).Warnf("problem with DeleteRole using roleID: %s, error: %+v", roleID, err)
		return err
	}

	return nil
}

// RemoveCLAUserRolesByProject will remove user CLA roles for the specified project SFID
func (ac *Client) RemoveCLAUserRolesByProject(projectSFID string, roleNames []string) error {
	f := logrus.Fields{
		"functionName": "DeleteRolesByObjectType",
		"projectSFID":  projectSFID,
		"roleNames":    strings.Join(roleNames, ","),
	}

	if projectSFID == "" {
		log.WithFields(f).Warn("unable to delete roles by project SFID - projectSFID is empty")
		return errors.New("empty project SFID")
	}

	if len(roleNames) == 0 {
		log.WithFields(f).Warn("unable to delete roles by project SFID with empty roleName list")
		return errors.New("empty role name list")
	}

	tok, err := token.GetToken()
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem obtaining token, error: %+v", err)
		return err
	}

	objectTypeID, err := ac.GetObjectTypeIDByName(utils.ProjectOrgScope)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem obtaining object type id by name: %s, error: %+v",
			utils.ProjectOrgScope, err)
		return err
	}

	params := &object_type.DeleteRolesByObjectTypeParams{
		ID:        int64(objectTypeID),
		Objectid:  projectSFID + "|*", // wildcard for organization scope - applies to all
		Rolenames: strings.Join(roleNames, ","),
		Context:   context.Background(),
	}

	_, err = ac.cl.ObjectType.DeleteRolesByObjectType(params, runtimeClient.BearerToken(tok)) // nolint
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem with DeleteRolesByObjectType using objectTypeID: %d and projectSFID: %s, error: %+v", objectTypeID, projectSFID, err)
		return err
	}

	return nil
}

// RemoveCLAUserRolesByProjectOrganization will remove user CLA roles for the specified project SFID
func (ac *Client) RemoveCLAUserRolesByProjectOrganization(projectSFID, organizationSFID string, roleNames []string) error {
	f := logrus.Fields{
		"functionName":     "RemoveCLAUserRolesByProjectOrganization",
		"projectSFID":      projectSFID,
		"organizationSFID": organizationSFID,
		"roleNames":        strings.Join(roleNames, ","),
	}

	if projectSFID == "" {
		log.WithFields(f).Warn("unable to delete roles by project SFID - projectSFID is empty")
		return errors.New("empty project SFID")
	}

	if organizationSFID == "" {
		log.WithFields(f).Warn("unable to delete roles by organization SFID - organizationSFID is empty")
		return errors.New("empty organization SFID")
	}

	if len(roleNames) == 0 {
		log.WithFields(f).Warn("unable to delete roles by project SFID and organization SFID with empty roleName list")
		return errors.New("empty role name list")
	}

	tok, err := token.GetToken()
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem obtaining token, error: %+v", err)
		return err
	}

	objectTypeID, err := ac.GetObjectTypeIDByName(utils.ProjectOrgScope)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem obtaining object type id by name: %s, error: %+v",
			utils.ProjectOrgScope, err)
		return err
	}

	params := &object_type.DeleteRolesByObjectTypeParams{
		ID:        int64(objectTypeID),
		Objectid:  projectSFID + "|" + organizationSFID,
		Rolenames: strings.Join(roleNames, ","),
		Context:   context.Background(),
	}

	_, err = ac.cl.ObjectType.DeleteRolesByObjectType(params, runtimeClient.BearerToken(tok)) // nolint
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("problem with DeleteRolesByObjectType using objectTypeID: %d, error: %+v", objectTypeID, err)
		return err
	}

	return nil
}

//UserScope entity representative of project and org for given role scope
type UserScope struct {
	Username string
	Email    string
	ObjectID string
}

// GetProjectRoleUsersScopes gets list of users with given role
func (ac *Client) GetProjectRoleUsersScopes(projectSFID, roleName string) ([]UserScope, error) {
	f := logrus.Fields{
		"functionName": "GetProjectRoleUsersScopes",
		"projectSFID":  projectSFID,
		"roleName":     roleName,
	}
	var userScopes []UserScope
	tok, err := token.GetToken()
	if err != nil {
		log.WithFields(f).Warnf("problem obtaining token, error: %+v", err)
		return nil, err
	}

	objectTypeID, err := ac.GetObjectTypeIDByName(utils.ProjectOrgScope)
	if err != nil {
		log.WithFields(f).Warnf("problem getting objectType ID for objectName :%s ", utils.ProjectOrgScope)
		return nil, err
	}
	log.WithFields(f).Debugf("Found objectID : %s for objectName: %s ", strconv.Itoa(objectTypeID), utils.ProjectOrgScope)
	clientAuth := runtimeClient.BearerToken(tok)

	log.WithFields(f).Debugf("Get objectRoleList with objectTypeID: %s ", strconv.Itoa(objectTypeID))
	params := &object_type.GetObjectTypeRoleListParams{
		ID:      strconv.Itoa(objectTypeID),
		Context: context.Background(),
	}

	response, err := ac.cl.ObjectType.GetObjectTypeRoleList(params, clientAuth)
	if err != nil {
		log.WithFields(f).Warnf("problem getting objectTypeRoleList for objectID: %s, error: %s ", strconv.Itoa(objectTypeID), err)
		return nil, err
	}

	for _, objectScope := range response.Payload.Data {
		for _, role := range objectScope.Roles {
			if role.Name == roleName {
				for _, scope := range role.Scopes {
					log.WithFields(f).Debugf("Identified user: %s , role: %s , objectID: %s", objectScope.User.Username, role.Name, scope.ObjectID)
					if strings.Split(scope.ObjectID, "|")[0] == projectSFID {
						log.WithFields(f).Debugf("Project: %s found in scope: %s ", projectSFID, scope.ObjectID)
						userScopes = append(userScopes, UserScope{
							Username: objectScope.User.Username,
							Email:    objectScope.User.Email,
							ObjectID: scope.ObjectID,
						})
					}
				}
			}
		}
	}

	return userScopes, nil
}
