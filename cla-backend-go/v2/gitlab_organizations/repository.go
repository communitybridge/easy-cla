// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package gitlab_organizations

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/communitybridge/easycla/cla-backend-go/v2/common"

	"github.com/gofrs/uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	models2 "github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/sirupsen/logrus"
)

// indexes
const (
	// GitlabOrgSFIDIndex the index for the SFID
	GitlabOrgSFIDIndex = "gitlab-org-sfid-index"
	// GitlabOrgLowerNameIndex the index for the group/org naem in lower case
	GitlabOrgLowerNameIndex = "gitlab-organization-name-lower-search-index"
	// GitLabExternalIDIndex the index for the external ID
	GitLabExternalIDIndex = "gitlab-external-group-id-index"
	// GitLabFullPathIndex the index for the full path
	GitLabFullPathIndex = "gitlab-full-path-index"
)

// RepositoryInterface is interface for gitlab org data model
type RepositoryInterface interface {
	AddGitLabOrganization(ctx context.Context, parentProjectSFID string, projectSFID string, groupID int64, organizationName, groupFullPath string, autoEnabled bool, autoEnabledClaGroupID string, branchProtectionEnabled bool, enabled bool) (*models2.GitlabOrganization, error)
	GetGitLabOrganizations(ctx context.Context, projectSFID string) (*models2.GitlabOrganizations, error)
	GetGitLabOrganization(ctx context.Context, gitlabOrganizationID string) (*common.GitLabOrganization, error)
	GetGitLabOrganizationByName(ctx context.Context, gitLabOrganizationName string) (*common.GitLabOrganization, error)
	GetGitLabOrganizationByExternalID(ctx context.Context, gitLabGroupID int64) (*common.GitLabOrganization, error)
	GetGitLabOrganizationByFullPath(ctx context.Context, groupFullPath string) (*common.GitLabOrganization, error)
	UpdateGitLabOrganizationAuth(ctx context.Context, organizationID string, gitLabGroupID int, authInfo, organizationFullPath, organizationURL string) error
	UpdateGitLabOrganization(ctx context.Context, projectSFID string, organizationName string, autoEnabled bool, autoEnabledClaGroupID string, branchProtectionEnabled bool, enabled bool) error
	UpdateGitLabOrganizationByExternalID(ctx context.Context, projectSFID string, groupID int64, organizationName, groupFullPath string, autoEnabled bool, autoEnabledClaGroupID string, branchProtectionEnabled bool, enabled bool) error
	DeleteGitLabOrganization(ctx context.Context, projectSFID, gitlabOrgName string) error
}

// Repository object/struct
type Repository struct {
	stage              string
	dynamoDBClient     *dynamodb.DynamoDB
	gitlabOrgTableName string
}

// NewRepository creates a new instance of the gitlabOrganizations repository
func NewRepository(awsSession *session.Session, stage string) RepositoryInterface {
	return &Repository{
		stage:              stage,
		dynamoDBClient:     dynamodb.New(awsSession),
		gitlabOrgTableName: fmt.Sprintf("cla-%s-gitlab-orgs", stage),
	}
}

func (repo *Repository) AddGitLabOrganization(ctx context.Context, parentProjectSFID string, projectSFID string, groupID int64, organizationName, groupFullPath string, autoEnabled bool, autoEnabledClaGroupID string, branchProtectionEnabled bool, enabled bool) (*models2.GitlabOrganization, error) {
	f := logrus.Fields{
		"functionName":            "v2.gitlab_organizations.repository.AddGitLabOrganization",
		utils.XREQUESTID:          ctx.Value(utils.XREQUESTID),
		"parentProjectSFID":       parentProjectSFID,
		"projectSFID":             projectSFID,
		"groupID":                 groupID,
		"organizationName":        organizationName,
		"groupFullPath":           groupFullPath,
		"autoEnabled":             autoEnabled,
		"autoEnabledClaGroupID":   autoEnabledClaGroupID,
		"branchProtectionEnabled": branchProtectionEnabled,
		"enabled":                 enabled,
	}

	var existingRecord *common.GitLabOrganization
	var getErr error
	if groupID != 0 {
		// First, let's check to see if we have an existing gitlab organization with the same name
		existingRecord, getErr = repo.GetGitLabOrganizationByExternalID(ctx, groupID)
		if getErr != nil {
			log.WithFields(f).WithError(getErr).Debugf("unable to locate existing GitLab group by name %d - ok to create a new record", groupID)
		}
	} else if groupFullPath != "" {
		// First, let's check to see if we have an existing gitlab organization with the same name
		existingRecord, getErr = repo.GetGitLabOrganizationByFullPath(ctx, groupFullPath)
		if getErr != nil {
			log.WithFields(f).WithError(getErr).Debugf("unable to locate existing GitLab group by full path: %s - ok to create a new record", groupFullPath)
		}
	}

	if existingRecord != nil {
		log.WithFields(f).Debugf("An existing GitLab organization with name %d exists in our database", groupID)
		// If everything matches...
		if projectSFID == existingRecord.ProjectSFID {
			log.WithFields(f).Debug("Existing GitLab organization with same SFID - should be able to update it")
			updateErr := repo.UpdateGitLabOrganizationByExternalID(ctx, projectSFID, groupID, organizationName, groupFullPath,
				autoEnabled, autoEnabledClaGroupID, branchProtectionEnabled, enabled)
			if updateErr != nil {
				return nil, updateErr
			}

			// Return the updated record
			if gitlabOrg, err := repo.GetGitLabOrganizationByExternalID(ctx, groupID); err != nil {
				return nil, err
			} else {
				return common.ToModel(gitlabOrg), nil
			}
		}

		log.WithFields(f).Debug("Existing GitLab organization with different project SFID - won't be able to update it - will return conflict")
		return nil, fmt.Errorf("record already exists")
	}

	// No existing records - create one
	_, currentTime := utils.CurrentTime()
	organizationID, err := uuid.NewV4()
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("Unable to generate a UUID for gitlab org, error: %v", err)
		return nil, err
	}

	authStateNonce, err := uuid.NewV4()
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("Unable to generate a auth nonce UUID for gitlab org, error: %v", err)
		return nil, err
	}

	gitlabOrg := &common.GitLabOrganization{
		OrganizationID:          organizationID.String(),
		DateCreated:             currentTime,
		DateModified:            currentTime,
		OrganizationName:        organizationName,
		OrganizationNameLower:   strings.ToLower(organizationName),
		ExternalGroupID:         int(groupID),
		OrganizationSFID:        parentProjectSFID,
		ProjectSFID:             projectSFID,
		Enabled:                 enabled,
		AutoEnabled:             autoEnabled,
		AutoEnabledClaGroupID:   autoEnabledClaGroupID,
		BranchProtectionEnabled: branchProtectionEnabled,
		AuthState:               authStateNonce.String(),
		Version:                 "v1",
		// OrganizationURL:         set later when we can authenticate to the API
	}

	log.WithFields(f).Debug("Encoding GitLab organization record for adding to the database...")
	av, err := dynamodbattribute.MarshalMap(gitlabOrg)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to marshall request for query")
		return nil, err
	}

	log.WithFields(f).Debug("Adding gitlab organization record to the database...")
	_, err = repo.dynamoDBClient.PutItem(&dynamodb.PutItemInput{
		Item:                av,
		TableName:           aws.String(repo.gitlabOrgTableName),
		ConditionExpression: aws.String("attribute_not_exists(organization_name)"),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.WithFields(f).WithError(err).Warn("gitlab organization already exists")
				return nil, fmt.Errorf("gitlab organization already exists")
			}
		}
		log.WithFields(f).WithError(err).Warn("cannot put gitlab organization in dynamodb")
		return nil, err
	}

	return common.ToModel(gitlabOrg), nil
}

// GetGitLabOrganizations get GitLab organizations based on the project SFID
func (repo *Repository) GetGitLabOrganizations(ctx context.Context, projectSFID string) (*models2.GitlabOrganizations, error) {
	f := logrus.Fields{
		"functionName":   "v2.gitlab_organizations.repository.GetGitLabOrganizations",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
	}

	condition := expression.Key(GitLabOrganizationsOrganizationSFIDColumn).Equal(expression.Value(projectSFID))
	builder := expression.NewBuilder().WithKeyCondition(condition)

	filter := expression.Name("enabled").Equal(expression.Value(true))
	builder = builder.WithFilter(filter)

	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		log.WithFields(f).Warnf("problem building query expression, error: %+v", err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(repo.gitlabOrgTableName),
		IndexName:                 aws.String(GitlabOrgSFIDIndex),
	}

	results, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.WithFields(f).Warnf("error retrieving gitlab_organizations using project_sfid = %s. error = %s", projectSFID, err.Error())
		return nil, err
	}

	if len(results.Items) == 0 {
		log.WithFields(f).Debug("no results from query")
		return &models2.GitlabOrganizations{
			List: []*models2.GitlabOrganization{},
		}, nil
	}

	var resultOutput []*common.GitLabOrganization
	err = dynamodbattribute.UnmarshalListOfMaps(results.Items, &resultOutput)
	if err != nil {
		return nil, err
	}

	log.WithFields(f).Debug("building response model...")
	gitlabOrgList := buildGitlabOrganizationListModels(ctx, resultOutput)
	return &models2.GitlabOrganizations{List: gitlabOrgList}, nil
}

// GetGitLabOrganizationByName get GitLab organization by name
func (repo *Repository) GetGitLabOrganizationByName(ctx context.Context, gitLabOrganizationName string) (*common.GitLabOrganization, error) {
	f := logrus.Fields{
		"functionName":           "v1.gitlab_organizations.repository.GetGitLabOrganizationByName",
		utils.XREQUESTID:         ctx.Value(utils.XREQUESTID),
		"gitLabOrganizationName": gitLabOrganizationName,
	}

	gitLabOrganizationName = strings.ToLower(gitLabOrganizationName)

	condition := expression.Key(GitLabOrganizationsOrganizationNameLowerColumn).Equal(expression.Value(strings.ToLower(gitLabOrganizationName)))
	builder := expression.NewBuilder().WithKeyCondition(condition)
	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		return nil, err
	}
	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(repo.gitlabOrgTableName),
		IndexName:                 aws.String(GitlabOrgLowerNameIndex),
	}

	log.WithFields(f).Debugf("querying for GitLab organization by name using organization_name_lower=%s...", strings.ToLower(gitLabOrganizationName))
	results, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("error retrieving gitlab_organizations using gitLabOrganizationName = %s", gitLabOrganizationName)
		return nil, err
	}
	if len(results.Items) == 0 {
		log.WithFields(f).Debug("Unable to find GitLab organization by name - no results")
		return nil, nil
	}

	var resultOutput []*common.GitLabOrganization
	err = dynamodbattribute.UnmarshalListOfMaps(results.Items, &resultOutput)
	if err != nil {
		log.WithFields(f).Warnf("problem decoding database results, error: %+v", err)
		return nil, err
	}

	return resultOutput[0], nil
}

func (repo *Repository) GetGitLabOrganizationByExternalID(ctx context.Context, gitLabGroupID int64) (*common.GitLabOrganization, error) {
	f := logrus.Fields{
		"functionName":   "v1.gitlab_organizations.repository.GetGitLabOrganizationByExternalID",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"gitLabGroupID":  gitLabGroupID,
	}

	condition := expression.Key(GitLabOrganizationsExternalGitLabGroupIDColumn).Equal(expression.Value(gitLabGroupID))
	builder := expression.NewBuilder().WithKeyCondition(condition)
	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(repo.gitlabOrgTableName),
		IndexName:                 aws.String(GitLabExternalIDIndex),
	}

	log.WithFields(f).Debugf("querying for GitLab organization by external group ID: %d...", gitLabGroupID)
	results, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("error retrieving gitlab_organizations using external ID = %d", gitLabGroupID)
		return nil, err
	}
	if len(results.Items) == 0 {
		log.WithFields(f).Debugf("Unable to find GitLab organization by group ID: %d - no results", gitLabGroupID)
		return nil, nil
	}

	var resultOutput []*common.GitLabOrganization
	err = dynamodbattribute.UnmarshalListOfMaps(results.Items, &resultOutput)
	if err != nil {
		log.WithFields(f).Warnf("problem decoding database results, error: %+v", err)
		return nil, err
	}

	return resultOutput[0], nil
}

// GetGitlabOrganizationByFullPath loads the organization based on the full path value
func (repo *Repository) GetGitLabOrganizationByFullPath(ctx context.Context, groupFullPath string) (*common.GitLabOrganization, error) {
	f := logrus.Fields{
		"functionName":   "v1.gitlab_organizations.repository.GetGitLabOrganizationByFullPath",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"groupFullPath":  groupFullPath,
	}

	condition := expression.Key(GitLabOrganizationsOrganizationFullPathColumn).Equal(expression.Value(groupFullPath))
	builder := expression.NewBuilder().WithKeyCondition(condition)
	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(repo.gitlabOrgTableName),
		IndexName:                 aws.String(GitLabFullPathIndex),
	}

	log.WithFields(f).Debugf("querying for GitLab group by full path: %s...", groupFullPath)
	results, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("error retrieving GitLab group by full path: %s", groupFullPath)
		return nil, err
	}
	if len(results.Items) == 0 {
		log.WithFields(f).Debugf("Unable to find GitLab group by full path: %s - no results", groupFullPath)
		return nil, nil
	}

	var resultOutput []*common.GitLabOrganization
	err = dynamodbattribute.UnmarshalListOfMaps(results.Items, &resultOutput)
	if err != nil {
		log.WithFields(f).Warnf("problem decoding database results, error: %+v", err)
		return nil, err
	}

	return resultOutput[0], nil
}

// GetGitLabOrganization by organization name
func (repo *Repository) GetGitLabOrganization(ctx context.Context, gitLabOrganizationID string) (*common.GitLabOrganization, error) {
	f := logrus.Fields{
		"functionName":         "gitlab_organizations.repository.GetGitLabOrganization",
		utils.XREQUESTID:       ctx.Value(utils.XREQUESTID),
		"gitLabOrganizationID": gitLabOrganizationID,
	}

	log.WithFields(f).Debugf("Querying for GitLab organization by ID: %s", gitLabOrganizationID)
	result, err := repo.dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			GitLabOrganizationsOrganizationIDColumn: {
				S: aws.String(gitLabOrganizationID),
			},
		},
		TableName: aws.String(repo.gitlabOrgTableName),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Item) == 0 {
		log.WithFields(f).Debugf("Unable to find GitLab organization by ID: %s - no results", gitLabOrganizationID)
		return nil, nil
	}

	var org common.GitLabOrganization
	err = dynamodbattribute.UnmarshalMap(result.Item, &org)
	if err != nil {
		log.WithFields(f).Warnf("error unmarshalling organization table data, error: %v", err)
		return nil, err
	}
	return &org, nil
}

// UpdateGitLabOrganizationAuth updates the specified Gitlab organization oauth info
func (repo *Repository) UpdateGitLabOrganizationAuth(ctx context.Context, organizationID string, gitLabGroupID int, authInfo, organizationFullPath, organizationURL string) error {
	f := logrus.Fields{
		"functionName":         "gitlab_organizations.repository.UpdateGitLabOrganizationAuth",
		utils.XREQUESTID:       ctx.Value(utils.XREQUESTID),
		"organizationID":       organizationID,
		"organizationFullPath": organizationFullPath,
		"organizationURL":      organizationURL,
		"tableName":            repo.gitlabOrgTableName,
	}

	_, currentTime := utils.CurrentTime()
	gitlabOrg, lookupErr := repo.GetGitLabOrganization(ctx, organizationID)
	if lookupErr != nil || gitlabOrg == nil {
		log.WithFields(f).Warnf("error looking up Gitlab organization by id: %s, error: %+v", organizationID, lookupErr)
		return lookupErr
	}

	expressionAttributeNames := map[string]*string{
		"#A":  aws.String(GitLabOrganizationsAuthInfoColumn),
		"#U":  aws.String(GitLabOrganizationsOrganizationURLColumn),
		"#FP": aws.String(GitLabOrganizationsOrganizationFullPathColumn),
		"#M":  aws.String(GitLabOrganizationsDateModifiedColumn),
		"#P":  aws.String(GitLabOrganizationsExternalGitLabGroupIDColumn),
	}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":a": {
			S: aws.String(authInfo),
		},
		":u": {
			S: aws.String(organizationURL),
		},
		":fp": {
			S: aws.String(organizationFullPath),
		},
		":m": {
			S: aws.String(currentTime),
		},
		":p": {
			N: aws.String(strconv.Itoa(gitLabGroupID)),
		},
	}

	updateExpression := "SET #A = :a, #U = :u, #FP = :fp, #M = :m, #P = :p"

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			GitLabOrganizationsOrganizationIDColumn: {
				S: aws.String(gitlabOrg.OrganizationID),
			},
		},
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		UpdateExpression:          &updateExpression,
		TableName:                 aws.String(repo.gitlabOrgTableName),
	}

	log.WithFields(f).Debug("updating gitlab organization record...")
	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.WithFields(f).Warnf("unable to update Gitlab organization record, error: %+v", updateErr)
		return updateErr
	}

	return nil
}

// UpdateGitLabOrganization updates the GitLab group based on the specified values
func (repo *Repository) UpdateGitLabOrganization(ctx context.Context, projectSFID string, organizationName string, autoEnabled bool, autoEnabledClaGroupID string, branchProtectionEnabled bool, enabled bool) error {
	f := logrus.Fields{
		"functionName":            "gitlab_organizations.repository.UpdateGitLabOrganization",
		utils.XREQUESTID:          ctx.Value(utils.XREQUESTID),
		"projectSFID":             projectSFID,
		"organizationName":        organizationName,
		"autoEnabled":             autoEnabled,
		"autoEnabledClaGroupID":   autoEnabledClaGroupID,
		"branchProtectionEnabled": branchProtectionEnabled,
		"tableName":               repo.gitlabOrgTableName,
	}

	_, currentTime := utils.CurrentTime()
	gitlabOrg, lookupErr := repo.GetGitLabOrganizationByName(ctx, organizationName)
	if lookupErr != nil {
		log.WithFields(f).Warnf("error looking up Gitlab organization by name, error: %+v", lookupErr)
		return lookupErr
	}
	if gitlabOrg == nil {
		log.WithFields(f).Warn("error looking up Gitlab organization - no results")
		return errors.New("unable to lookup Gitlab organization by name")
	}

	expressionAttributeNames := map[string]*string{
		"#A": aws.String(GitLabOrganizationsAutoEnabledColumn),
		"#C": aws.String(GitLabOrganizationsAutoEnabledCLAGroupIDColumn),
		"#B": aws.String(GitLabOrganizationsBranchProtectionEnabledColumn),
		"#M": aws.String(GitLabOrganizationsDateModifiedColumn),
		"#E": aws.String(GitLabOrganizationsEnabledColumn),
	}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":a": {
			BOOL: aws.Bool(autoEnabled),
		},
		":c": {
			S: aws.String(autoEnabledClaGroupID),
		},
		":b": {
			BOOL: aws.Bool(branchProtectionEnabled),
		},
		":m": {
			S: aws.String(currentTime),
		},
		":e": {
			BOOL: aws.Bool(enabled),
		},
	}
	updateExpression := "SET #A = :a, #C = :c, #B = :b, #M = :m, #E = :e"

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			GitLabOrganizationsOrganizationIDColumn: {
				S: aws.String(gitlabOrg.OrganizationID),
			},
		},
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		UpdateExpression:          &updateExpression,
		TableName:                 aws.String(repo.gitlabOrgTableName),
	}

	log.WithFields(f).Debugf("updating gitlab organization record: %+v", input)
	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.WithFields(f).Warnf("unable to update Gitlab organization record, error: %+v", updateErr)
		return updateErr
	}

	return nil
}

// UpdateGitlabOrganizationByExternalID updates the GitLab group based on the specified values
func (repo *Repository) UpdateGitLabOrganizationByExternalID(ctx context.Context, projectSFID string, groupID int64, organizationName, fullPath string, autoEnabled bool, autoEnabledClaGroupID string, branchProtectionEnabled bool, enabled bool) error {
	f := logrus.Fields{
		"functionName":            "gitlab_organizations.repository.UpdateGitLabOrganizationByExternalID",
		utils.XREQUESTID:          ctx.Value(utils.XREQUESTID),
		"projectSFID":             projectSFID,
		"groupID":                 groupID,
		"organizationName":        organizationName,
		"autoEnabled":             autoEnabled,
		"autoEnabledClaGroupID":   autoEnabledClaGroupID,
		"branchProtectionEnabled": branchProtectionEnabled,
		"tableName":               repo.gitlabOrgTableName,
	}

	_, currentTime := utils.CurrentTime()
	gitlabOrg, lookupErr := repo.GetGitLabOrganizationByExternalID(ctx, groupID)
	if lookupErr != nil {
		log.WithFields(f).Warnf("error looking up GitLab group by ID: %d, error: %+v", groupID, lookupErr)
		return lookupErr
	}
	if gitlabOrg == nil {
		log.WithFields(f).Warn("error looking up GitLab group - no results")
		return errors.New("unable to lookup GitLab group by ID")
	}

	expressionAttributeNames := map[string]*string{
		"#A":  aws.String(GitLabOrganizationsAutoEnabledColumn),
		"#C":  aws.String(GitLabOrganizationsAutoEnabledCLAGroupIDColumn),
		"#B":  aws.String(GitLabOrganizationsBranchProtectionEnabledColumn),
		"#N":  aws.String(GitLabOrganizationsOrganizationNameColumn),
		"#NL": aws.String(GitLabOrganizationsOrganizationNameLowerColumn),
		"#M":  aws.String(GitLabOrganizationsDateModifiedColumn),
		"#E":  aws.String(GitLabOrganizationsEnabledColumn),
	}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":a": {
			BOOL: aws.Bool(autoEnabled),
		},
		":c": {
			S: aws.String(autoEnabledClaGroupID),
		},
		":b": {
			BOOL: aws.Bool(branchProtectionEnabled),
		},
		":n": {
			S: aws.String(organizationName),
		},
		":nl": {
			S: aws.String(strings.ToLower(organizationName)),
		},
		":m": {
			S: aws.String(currentTime),
		},
		":e": {
			BOOL: aws.Bool(enabled),
		},
	}
	updateExpression := "SET #A = :a, #C = :c, #B = :b, #N = :n, #NL = :nl, #M = :m, #E = :e "

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			GitLabOrganizationsOrganizationIDColumn: {
				S: aws.String(gitlabOrg.OrganizationID),
			},
		},
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		UpdateExpression:          &updateExpression,
		TableName:                 aws.String(repo.gitlabOrgTableName),
	}

	log.WithFields(f).Debugf("updating GitLab organization record: %+v", input)
	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.WithFields(f).Warnf("unable to update GitLab organization record, error: %+v", updateErr)
		return updateErr
	}

	return nil
}

// DeleteGitLabOrganization deletes the specified GitLab organization
func (repo *Repository) DeleteGitLabOrganization(ctx context.Context, projectSFID, gitlabOrgName string) error {
	f := logrus.Fields{
		"functionName":   "v1.gitlab_organizations.repository.DeleteGitLabOrganization",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
		"gitlabOrgName":  gitlabOrgName,
	}

	var gitlabOrganizationID string
	orgs, orgErr := repo.GetGitLabOrganizations(ctx, projectSFID)
	if orgErr != nil {
		errMsg := fmt.Sprintf("gitlab organization is not found using projectSFID: %s, error: %+v", projectSFID, orgErr)
		log.WithFields(f).Warn(errMsg)
		return errors.New(errMsg)
	}

	for _, gitLabOrg := range orgs.List {
		if strings.EqualFold(gitLabOrg.OrganizationName, gitlabOrgName) {
			gitlabOrganizationID = gitLabOrg.OrganizationID
			break
		}
	}

	log.WithFields(f).Debug("Deleting GitLab organization...")
	// Update enabled flag as false
	_, currentTime := utils.CurrentTime()
	note := fmt.Sprintf("Enabled set to false due to org deletion at %s ", currentTime)
	_, err := repo.dynamoDBClient.UpdateItem(
		&dynamodb.UpdateItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				GitLabOrganizationsOrganizationIDColumn: {
					S: aws.String(gitlabOrganizationID),
				},
			},
			ExpressionAttributeNames: map[string]*string{
				"#E": aws.String(GitLabOrganizationsEnabledColumn),
				"#N": aws.String(GitLabOrganizationsNoteColumn),
				"#D": aws.String(GitLabOrganizationsDateModifiedColumn),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":e": {
					BOOL: aws.Bool(false),
				},
				":n": {
					S: aws.String(note),
				},
				":d": {
					S: aws.String(currentTime),
				},
			},
			UpdateExpression: aws.String("SET #E = :e, #N = :n, #D = :d"),
			TableName:        aws.String(repo.gitlabOrgTableName),
		},
	)
	if err != nil {
		errMsg := fmt.Sprintf("error deleting gitlab organization: %s - %+v", gitlabOrgName, err)
		log.WithFields(f).Warnf(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func buildGitlabOrganizationListModels(ctx context.Context, gitlabOrganizations []*common.GitLabOrganization) []*models2.GitlabOrganization {
	f := logrus.Fields{
		"functionName":   "buildGitlabOrganizationListModels",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
	}

	log.WithFields(f).Debugf("fetching gitlab info for the list")
	// Convert the database model to a response model
	return common.ToModels(gitlabOrganizations)

	// TODO: Fetch the gitlab information
}
