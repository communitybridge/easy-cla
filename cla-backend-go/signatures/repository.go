// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package signatures

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/communitybridge/easycla/cla-backend-go/users"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/signatures"

	"github.com/communitybridge/easycla/cla-backend-go/company"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// SignatureRepository interface defines the functions for the github whitelist service
type SignatureRepository interface {
	GetGithubOrganizationsFromWhitelist(signatureID string) ([]models.GithubOrg, error)
	AddGithubOrganizationToWhitelist(signatureID, githubOrganizationID string) ([]models.GithubOrg, error)
	DeleteGithubOrganizationFromWhitelist(signatureID, githubOrganizationID string) ([]models.GithubOrg, error)

	GetSignatures(params signatures.GetSignaturesParams, pageSize int64) (*models.Signatures, error)
	GetProjectSignatures(params signatures.GetProjectSignaturesParams, pageSize int64) (*models.Signatures, error)
	GetProjectCompanySignatures(params signatures.GetProjectCompanySignaturesParams, pageSize int64) (*models.Signatures, error)
	GetProjectCompanyEmployeeSignatures(params signatures.GetProjectCompanyEmployeeSignaturesParams, pageSize int64) (*models.Signatures, error)
	GetCompanySignatures(params signatures.GetCompanySignaturesParams, pageSize int64) (*models.Signatures, error)
	GetUserSignatures(params signatures.GetUserSignaturesParams, pageSize int64) (*models.Signatures, error)
}

// repository data model
type repository struct {
	stage          string
	dynamoDBClient *dynamodb.DynamoDB
	companyRepo    company.RepositoryService
	usersRepo      users.Service
}

// NewRepository creates a new instance of the whitelist service
func NewRepository(awsSession *session.Session, stage string, companyRepo company.RepositoryService, usersRepo users.Service) SignatureRepository {
	return repository{
		stage:          stage,
		dynamoDBClient: dynamodb.New(awsSession),
		companyRepo:    companyRepo,
		usersRepo:      usersRepo,
	}
}

// GetGithubOrganizationsFromWhitelist returns a list of GH organizations stored in the whitelist
func (repo repository) GetGithubOrganizationsFromWhitelist(signatureID string) ([]models.GithubOrg, error) {
	// get item from dynamoDB table
	tableName := fmt.Sprintf("cla-%s-signatures", repo.stage)
	result, err := repo.dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
	})

	if err != nil {
		log.Warnf("Error retrieving GH organization whitelist for signatureID: %s, error: %v", signatureID, err)
		return nil, err
	}

	itemFromMap, ok := result.Item["github_org_whitelist"]
	if !ok {
		return nil, nil
	}

	var orgs []models.GithubOrg
	for _, org := range itemFromMap.L {
		selected := true
		orgs = append(orgs, models.GithubOrg{
			ID:       org.S,
			Selected: &selected,
		})
	}

	// Sort the array based on the ID
	sort.Slice(orgs, func(i, j int) bool {
		return *orgs[i].ID < *orgs[j].ID
	})

	return orgs, nil
}

// AddGithubOrganizationToWhitelist adds the specified GH organization to the whitelist
func (repo repository) AddGithubOrganizationToWhitelist(signatureID, GithubOrganizationID string) ([]models.GithubOrg, error) {
	// get item from dynamoDB table
	tableName := fmt.Sprintf("cla-%s-signatures", repo.stage)
	log.Debugf("querying database for github organization whitelist using signatureID: %s", signatureID)

	result, err := repo.dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
	})

	if err != nil {
		log.Warnf("Error retrieving GH organization whitelist for signatureID: %s and GH Org: %s, error: %v",
			signatureID, GithubOrganizationID, err)
		return nil, err
	}

	itemFromMap, ok := result.Item["github_org_whitelist"]
	if !ok {
		log.Debugf("signatureID: %s is missing the 'github_org_whitelist' column - will add", signatureID)
		itemFromMap = &dynamodb.AttributeValue{}
	}

	// generate new List L without element to be deleted
	// if we find a org with the same id just return without updating the record
	var newList []*dynamodb.AttributeValue
	for _, element := range itemFromMap.L {
		newList = append(newList, element)
		if *element.S == GithubOrganizationID {
			log.Debugf("github organization for signature: %s already in the list - nothing to do, org id: %s",
				signatureID, GithubOrganizationID)
			return buildResponse(itemFromMap.L), nil
		}
	}

	// Add the organization to list
	log.Debugf("adding github organization for signature: %s to the list, org id: %s",
		signatureID, GithubOrganizationID)
	newList = append(newList, &dynamodb.AttributeValue{
		S: aws.String(GithubOrganizationID),
	})

	// return values flag - Returns all of the attributes of the item, as they appear after the UpdateItem operation.
	addReturnValues := "ALL_NEW" // nolint

	// Update dynamoDB table
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#L": aws.String("github_org_whitelist"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":l": {
				L: newList,
			},
		},
		UpdateExpression: aws.String("SET #L = :l"),
		ReturnValues:     &addReturnValues,
	}

	log.Warnf("updating database record using signatureID: %s with values: %v", signatureID, newList)
	updatedValues, err := repo.dynamoDBClient.UpdateItem(input)
	if err != nil {
		log.Warnf("Error updating white list, error: %v", err)
		return nil, err
	}

	updatedItemFromMap, ok := updatedValues.Attributes["github_org_whitelist"]
	if !ok {
		msg := fmt.Sprintf("unable to fetch updated whitelist organization values for "+
			"organization id: %s for signature: %s - list is empty - returning empty list",
			GithubOrganizationID, signatureID)
		log.Debugf(msg)
		return []models.GithubOrg{}, nil
	}

	return buildResponse(updatedItemFromMap.L), nil
}

// DeleteGithubOrganizationFromWhitelist removes the specified GH organization from the whitelist
func (repo repository) DeleteGithubOrganizationFromWhitelist(signatureID, GithubOrganizationID string) ([]models.GithubOrg, error) {
	// get item from dynamoDB table
	tableName := fmt.Sprintf("cla-%s-signatures", repo.stage)
	result, err := repo.dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
	})

	if err != nil {
		log.Warnf("error retrieving GH organization whitelist for signatureID: %s and GH Org: %s, error: %v",
			signatureID, GithubOrganizationID, err)
		return nil, err
	}

	itemFromMap, ok := result.Item["github_org_whitelist"]
	if !ok {
		log.Warnf("unable to remove whitelist organization: %s for signature: %s - list is empty",
			GithubOrganizationID, signatureID)
		return nil, errors.New("no github_org_whitelist column")
	}

	// generate new List L without element to be deleted
	var newList []*dynamodb.AttributeValue
	for _, element := range itemFromMap.L {
		if *element.S != GithubOrganizationID {
			newList = append(newList, element)
		}
	}

	if len(newList) == 0 {
		// Since we don't have any items in our list, we can't simply update dynamoDB with an empty list,
		// nooooo, that would be too easy. Instead:
		// We need to set the value to NULL to clear it out (otherwise we'll get a validation error like:)
		// ValidationException: ExpressionAttributeValues contains invalid value: Supplied AttributeValue
		// is empty, must contain exactly one of the supported datatypes for key)

		log.Debugf("clearing out github org whitelist for organization: %s for signature: %s - list is empty",
			GithubOrganizationID, signatureID)
		nullFlag := true

		// update dynamoDB table
		input := &dynamodb.UpdateItemInput{
			ExpressionAttributeNames: map[string]*string{
				"#L": aws.String("github_org_whitelist"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":l": {
					NULL: &nullFlag,
				},
			},
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"signature_id": {
					S: aws.String(signatureID),
				},
			},
			UpdateExpression: aws.String("SET #L = :l"),
		}

		_, err = repo.dynamoDBClient.UpdateItem(input)
		if err != nil {
			log.Warnf("error updating github org whitelist to NULL value, error: %v", err)
			return nil, err
		}

		// Return an empty list
		return []models.GithubOrg{}, nil
	}

	// return values flag - Returns all of the attributes of the item, as they appear after the UpdateItem operation.
	updatedReturnValues := "ALL_NEW" // nolint

	// update dynamoDB table
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#L": aws.String("github_org_whitelist"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":l": {
				L: newList,
			},
		},
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
		UpdateExpression: aws.String("SET #L = :l"),
		ReturnValues:     &updatedReturnValues,
	}

	updatedValues, err := repo.dynamoDBClient.UpdateItem(input)
	if err != nil {
		log.Warnf("Error updating github org whitelist, error: %v", err)
		return nil, err
	}

	updatedItemFromMap, ok := updatedValues.Attributes["github_org_whitelist"]
	if !ok {
		msg := fmt.Sprintf("unable to fetch updated whitelist organization values for "+
			"organization id: %s for signature: %s - list is empty - returning empty list",
			GithubOrganizationID, signatureID)
		log.Debugf(msg)
		return []models.GithubOrg{}, nil
	}

	return buildResponse(updatedItemFromMap.L), nil

}

// GetSignatures returns a list of signatures for the specified sigature ID
func (repo repository) GetSignatures(params signatures.GetSignaturesParams, pageSize int64) (*models.Signatures, error) {

	queryStartTime := time.Now()

	// The table we're interested in
	tableName := fmt.Sprintf("cla-%s-signatures", repo.stage)

	// This is the key we want to match
	condition := expression.Key("signature_id").Equal(expression.Value(params.SignatureID))

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(buildProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for signature ID query, signatureID: %s, error: %v",
			params.SignatureID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		Limit:                     aws.Int64(pageSize), // The maximum number of items to evaluate (not necessarily the number of matching items)
		TableName:                 aws.String(tableName),
	}

	// If we have the next key, set the exclusive start key value
	if params.NextKey != nil {
		log.Debugf("Received a nextKey, value: %s", *params.NextKey)
		// The primary key of the first item that this operation will evaluate.
		// and the query key (if not the same)
		queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: params.NextKey,
			},
		}
	}

	var signatures []models.Signature
	var lastEvaluatedKey string

	// Loop until we have all the records - should only find the one signature
	// but we'll implement the same last key evaluation logic anyway
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		//log.Debugf("Running signature project company query using queryInput: %+v", queryInput)
		results, queryErr := repo.dynamoDBClient.Query(queryInput)
		if queryErr != nil {
			log.Warnf("error retrieving signature ID: %s, error: %v",
				params.SignatureID, queryErr)
			return nil, queryErr
		}

		log.Debugf("Signature query took: %v resulting in %d results",
			utils.FmtDuration(time.Since(queryStartTime)), len(results.Items))

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, "")
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for signature: %s, error: %v",
				params.SignatureID, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		signatures = append(signatures, signatureList...)

		log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"signature_id": {
					S: aws.String(lastEvaluatedKey),
				},
			}
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(signatures)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count, error: %v", err)
		return nil, err
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount

	return &models.Signatures{
		ProjectID:      "",
		ResultCount:    int64(len(signatures)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     signatures,
	}, nil
}

// GetProjectSignatures returns a list of signatures for the specified project
func (repo repository) GetProjectSignatures(params signatures.GetProjectSignaturesParams, pageSize int64) (*models.Signatures, error) {

	queryStartTime := time.Now()

	// The table we're interested in
	tableName := fmt.Sprintf("cla-%s-signatures", repo.stage)

	indexName := "project-signature-index"

	// This is the key we want to match
	condition := expression.Key("signature_project_id").Equal(expression.Value(params.ProjectID))

	builder := expression.NewBuilder().WithProjection(buildProjection())

	if params.SearchTerm != nil && params.SearchField != nil {
		if *params.FullMatch {
			indexName = "reference-signature-search-index"

			condition = condition.And(expression.Key("signature_reference_name_lower").Equal(expression.Value(strings.ToLower(*params.SearchTerm))))

			filter := expression.Name("signature_reference_type").Equal(expression.Value(params.SearchField))
			builder = builder.WithFilter(filter)
		} else {
			filter := expression.Name("signature_reference_name_lower").Contains(strings.ToLower(*params.SearchTerm)).
				And(expression.Name("signature_reference_type").Equal(expression.Value(params.SearchField)))
			builder = builder.WithFilter(filter)
		}
	}

	builder = builder.WithKeyCondition(condition)

	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		log.Warnf("error building expression for project signature query, projectID: %s, error: %v",
			params.ProjectID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(tableName),
		Limit:                     aws.Int64(pageSize),   // The maximum number of items to evaluate (not necessarily the number of matching items)
		IndexName:                 aws.String(indexName), // Name of a secondary index to scan
	}

	// If we have the next key, set the exclusive start key value
	if params.NextKey != nil {
		log.Debugf("Received a nextKey, value: %s", *params.NextKey)
		// The primary key of the first item that this operation will evaluate.
		// and the query key (if not the same)
		queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: params.NextKey,
			},
			"signature_project_id": {
				S: &params.ProjectID,
			},
		}
	}

	var signatures []models.Signature
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		log.Debugf("Running signature project query using queryInput: %+v", queryInput)
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		if errQuery != nil {
			log.Warnf("error retrieving project signature ID for project: %s, error: %v",
				params.ProjectID, errQuery)
			return nil, errQuery
		}

		log.Debugf("Signature project query took: %v resulting in %d results",
			utils.FmtDuration(time.Since(queryStartTime)), len(results.Items))

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, params.ProjectID)
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for signatures with project %s, error: %v",
				params.ProjectID, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		signatures = append(signatures, signatureList...)

		log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"signature_id": {
					S: aws.String(lastEvaluatedKey),
				},
				"signature_project_id": {
					S: &params.ProjectID,
				},
			}
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(signatures)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count for project: %s, error: %v", params.ProjectID, err)
		return nil, err
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount

	return &models.Signatures{
		ProjectID:      params.ProjectID,
		ResultCount:    int64(len(signatures)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     signatures,
	}, nil
}

// GetProjectCompanySignatures returns a list of signatures for the specified project and specified company
func (repo repository) GetProjectCompanySignatures(params signatures.GetProjectCompanySignaturesParams, pageSize int64) (*models.Signatures, error) {

	queryStartTime := time.Now()

	// The table we're interested in
	tableName := fmt.Sprintf("cla-%s-signatures", repo.stage)

	// These are the keys we want to match
	condition := expression.Key("signature_project_id").Equal(expression.Value(params.ProjectID))
	filter := expression.Name("signature_reference_id").Equal(expression.Value(params.CompanyID))

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithFilter(filter).WithProjection(buildProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for project signature ID query, project: %s, error: %v",
			params.ProjectID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("project-signature-index"), // Name of a secondary index to scan
		//Limit:                     aws.Int64(pageSize),                   // The maximum number of items to evaluate (not necessarily the number of matching items)
	}

	// If we have the next key, set the exclusive start key value
	if params.NextKey != nil {
		log.Debugf("Received a nextKey, value: %s", *params.NextKey)
		// The primary key of the first item that this operation will evaluate.
		// and the query key (if not the same)
		queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: params.NextKey,
			},
			"signature_project_id": {
				S: &params.ProjectID,
			},
		}
	}

	var signatures []models.Signature
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		//log.Debugf("Running signature project company query using queryInput: %+v", queryInput)
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		if errQuery != nil {
			log.Warnf("error retrieving project signature ID for project: %s with company: %s, error: %v",
				params.ProjectID, params.CompanyID, errQuery)
			return nil, errQuery
		}

		log.Debugf("Signature project company query took: %v resulting in %d results",
			utils.FmtDuration(time.Since(queryStartTime)), len(results.Items))

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, params.ProjectID)
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for signatures with project %s with company: %s, error: %v",
				params.ProjectID, params.CompanyID, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		signatures = append(signatures, signatureList...)

		log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"signature_id": {
					S: aws.String(lastEvaluatedKey),
				},
				"signature_project_id": {
					S: &params.ProjectID,
				},
			}
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(signatures)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count for project: %s, error: %v", params.ProjectID, err)
		return nil, err
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount

	return &models.Signatures{
		ProjectID:      params.ProjectID,
		ResultCount:    int64(len(signatures)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     signatures,
	}, nil
}

// GetProjectCompanyEmployeeSignatures returns a list of employee signatures for the specified project and specified company
func (repo repository) GetProjectCompanyEmployeeSignatures(params signatures.GetProjectCompanyEmployeeSignaturesParams, pageSize int64) (*models.Signatures, error) {

	queryStartTime := time.Now()

	// The table we're interested in
	tableName := fmt.Sprintf("cla-%s-signatures", repo.stage)

	// This is the keys we want to match
	condition := expression.Key("signature_project_id").Equal(expression.Value(params.ProjectID))
	filter := expression.Name("signature_user_ccla_company_id").Equal(expression.Value(params.CompanyID))

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithFilter(filter).WithProjection(buildProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for project signature ID query, project: %s, error: %v",
			params.ProjectID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("project-signature-index"), // Name of a secondary index to scan
		Limit:                     aws.Int64(pageSize),                   // The maximum number of items to evaluate (not necessarily the number of matching items)
	}

	// If we have the next key, set the exclusive start key value
	if params.NextKey != nil {
		log.Debugf("Received a nextKey, value: %s", *params.NextKey)
		// The primary key of the first item that this operation will evaluate.
		// and the query key (if not the same)
		queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: params.NextKey,
			},
			"signature_project_id": {
				S: &params.ProjectID,
			},
		}
	}

	var signatures []models.Signature
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		//log.Debugf("Running signature project company query using queryInput: %+v", queryInput)
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		if errQuery != nil {
			log.Warnf("error retrieving project company employee signature ID for project: %s with company: %s, error: %v",
				params.ProjectID, params.CompanyID, errQuery)
			return nil, errQuery
		}

		log.Debugf("Signature project company employee query took: %v resulting in %d results",
			utils.FmtDuration(time.Since(queryStartTime)), len(results.Items))

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, params.ProjectID)
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for employee signatures with project %s with company: %s, error: %v",
				params.ProjectID, params.CompanyID, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		signatures = append(signatures, signatureList...)

		log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"signature_id": {
					S: aws.String(lastEvaluatedKey),
				},
				"signature_project_id": {
					S: &params.ProjectID,
				},
			}
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(signatures)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count for project: %s, error: %v", params.ProjectID, err)
		return nil, err
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount

	return &models.Signatures{
		ProjectID:      params.ProjectID,
		ResultCount:    int64(len(signatures)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     signatures,
	}, nil
}

// GetCompanySignatures returns a list of company signatures for the specified company
func (repo repository) GetCompanySignatures(params signatures.GetCompanySignaturesParams, pageSize int64) (*models.Signatures, error) {

	queryStartTime := time.Now()

	// The table we're interested in
	tableName := fmt.Sprintf("cla-%s-signatures", repo.stage)

	// This is the keys we want to match
	condition := expression.Key("signature_reference_id").Equal(expression.Value(params.CompanyID))

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(buildProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for company signature query, companyID: %s, error: %v",
			params.CompanyID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("reference-signature-index"), // Name of a secondary index to scan
		//Limit:                     aws.Int64(pageSize),                   // The maximum number of items to evaluate (not necessarily the number of matching items)
	}

	// If we have the next key, set the exclusive start key value
	if params.NextKey != nil {
		log.Debugf("Received a nextKey, value: %s", *params.NextKey)
		// The primary key of the first item that this operation will evaluate.
		// and the query key (if not the same)
		queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: params.NextKey,
			},
			"signature_reference_id": {
				S: &params.CompanyID,
			},
		}
	}

	var signatures []models.Signature
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		//log.Debugf("Running signature project company query using queryInput: %+v", queryInput)
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		if errQuery != nil {
			log.Warnf("error retrieving company signature ID for company: %s with company: %s, error: %v",
				params.CompanyID, params.CompanyID, errQuery)
			return nil, errQuery
		}

		log.Debugf("Signature company query took: %v resulting in %d results",
			utils.FmtDuration(time.Since(queryStartTime)), len(results.Items))

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, "")
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for signatures with company: %s, error: %v",
				params.CompanyID, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		signatures = append(signatures, signatureList...)

		log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"signature_id": {
					S: aws.String(lastEvaluatedKey),
				},
				"signature_reference_id": {
					S: &params.CompanyID,
				},
			}
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(signatures)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count for company: %s/%s, error: %v",
			params.CompanyID, *params.CompanyName, err)
		return nil, err
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount

	return &models.Signatures{
		ProjectID:      "",
		ResultCount:    int64(len(signatures)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     signatures,
	}, nil
}

// GetUserSignatures returns a list of user signatures for the specified user
func (repo repository) GetUserSignatures(params signatures.GetUserSignaturesParams, pageSize int64) (*models.Signatures, error) {

	queryStartTime := time.Now()

	// The table we're interested in
	tableName := fmt.Sprintf("cla-%s-signatures", repo.stage)

	// This is the keys we want to match
	condition := expression.Key("signature_reference_id").Equal(expression.Value(params.UserID))

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(buildProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for user signature query, userID: %s, error: %v",
			params.UserID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("reference-signature-index"), // Name of a secondary index to scan
		//Limit:                     aws.Int64(pageSize),                   // The maximum number of items to evaluate (not necessarily the number of matching items)
	}

	// If we have the next key, set the exclusive start key value
	if params.NextKey != nil {
		log.Debugf("Received a nextKey, value: %s", *params.NextKey)
		// The primary key of the first item that this operation will evaluate.
		// and the query key (if not the same)
		queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: params.NextKey,
			},
			"signature_reference_id": {
				S: &params.UserID,
			},
		}
	}

	var signatures []models.Signature
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		//log.Debugf("Running signature project company query using queryInput: %+v", queryInput)
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		if errQuery != nil {
			log.Warnf("error retrieving user signatures for user: %s/%s, error: %v",
				params.UserID, *params.UserName, errQuery)
			return nil, errQuery
		}

		log.Debugf("Signature user query took: %v resulting in %d results",
			utils.FmtDuration(time.Since(queryStartTime)), len(results.Items))

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, "")
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for signatures for user %s/%s, error: %v",
				params.UserID, *params.UserName, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		signatures = append(signatures, signatureList...)

		log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"signature_id": {
					S: aws.String(lastEvaluatedKey),
				},
				"signature_reference_id": {
					S: &params.UserID,
				},
			}
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(signatures)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count for user: %s/%s, error: %v",
			params.UserID, *params.UserName, err)
		return nil, err
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount

	return &models.Signatures{
		ProjectID:      "",
		ResultCount:    int64(len(signatures)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     signatures,
	}, nil
}

// buildProjectSignatureModels converts the response model into a response data model
func (repo repository) buildProjectSignatureModels(results *dynamodb.QueryOutput, projectID string) ([]models.Signature, error) {
	var signatures []models.Signature

	type ItemSignature struct {
		SignatureID                   string   `json:"signature_id"`
		DateCreated                   string   `json:"date_created"`
		DateModified                  string   `json:"date_modified"`
		SignatureApproved             bool     `json:"signature_approved"`
		SignatureSigned               bool     `json:"signature_signed"`
		SignatureDocumentMajorVersion string   `json:"signature_document_major_version"`
		SignatureDocumentMinorVersion string   `json:"signature_document_minor_version"`
		SignatureReferenceID          string   `json:"signature_reference_id"`
		SignatureReferenceName        string   `json:"signature_reference_name"`
		SignatureReferenceNameLower   string   `json:"signature_reference_name_lower"`
		SignatureProjectID            string   `json:"signature_project_id"`
		SignatureReferenceType        string   `json:"signature_reference_type"`
		SignatureType                 string   `json:"signature_type"`
		SignatureUserCompanyID        string   `json:"signature_user_ccla_company_id"`
		EmailWhitelist                []string `json:"email_whitelist"`
		DomainWhitelist               []string `json:"domain_whitelist"`
		GitHubWhitelist               []string `json:"github_whitelist"`
		GitHubOrgWhitelist            []string `json:"github_org_whitelist"`
		SignatureACL                  []string `json:"signature_acl"`
	}

	// The DB signature model
	var dbSignature []ItemSignature

	err := dynamodbattribute.UnmarshalListOfMaps(results.Items, &dbSignature)
	if err != nil {
		log.Warnf("error unmarshalling signatures from database for project: %s, error: %v",
			projectID, err)
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(len(dbSignature))

	for _, dbSignature := range dbSignature {
		go func(dbSignature ItemSignature) {
			defer wg.Done()
			var companyName = ""
			var userName = ""
			var userLFID = ""
			var userGHID = ""

			if dbSignature.SignatureReferenceType == "user" {
				userModel, userErr := repo.usersRepo.GetUser(dbSignature.SignatureReferenceID)
				if userErr != nil || userModel == nil {
					log.Warnf("unable to lookup user using id: %s, error: %v", dbSignature.SignatureReferenceID, userErr)
				} else {
					userName = userModel.Username
					userLFID = userModel.LfUsername
					userGHID = userModel.GithubID
				}
				if dbSignature.SignatureUserCompanyID != "" {
					dbCompanyModel, companyErr := repo.companyRepo.GetCompany(dbSignature.SignatureUserCompanyID)
					if companyErr != nil {
						log.Warnf("unable to lookup company using id: %s, error: %v", dbSignature.SignatureUserCompanyID, companyErr)
					} else {
						companyName = dbCompanyModel.CompanyName
					}
				}
			} else if dbSignature.SignatureReferenceType == "company" {
				dbCompanyModel, companyErr := repo.companyRepo.GetCompany(dbSignature.SignatureReferenceID)
				if companyErr != nil {
					log.Warnf("unable to lookup company using id: %s, error: %v", dbSignature.SignatureReferenceID, companyErr)
				} else {
					companyName = dbCompanyModel.CompanyName
				}
			}

			var signatureACL []models.User
			for _, userName := range dbSignature.SignatureACL {
				userModel, userErr := repo.usersRepo.GetUserByUserName(userName, true)
				if userErr != nil {
					log.Warnf("unable to lookup user using username: %s, error: %v", userName, userErr)
				} else {
					if userModel == nil {
						log.Warnf("User looking for username is null: %s for signature: %s", userName, dbSignature.SignatureID)
					} else {
						signatureACL = append(signatureACL, *userModel)
					}
				}
			}

			signatures = append(signatures, models.Signature{
				SignatureID:                 dbSignature.SignatureID,
				CompanyName:                 companyName,
				SignatureCreated:            dbSignature.DateCreated,
				SignatureModified:           dbSignature.DateModified,
				SignatureType:               dbSignature.SignatureType,
				SignatureReferenceID:        dbSignature.SignatureReferenceID,
				SignatureReferenceName:      dbSignature.SignatureReferenceName,
				SignatureReferenceNameLower: dbSignature.SignatureReferenceNameLower,
				SignatureSigned:             dbSignature.SignatureSigned,
				SignatureApproved:           dbSignature.SignatureApproved,
				Version:                     dbSignature.SignatureDocumentMajorVersion + "." + dbSignature.SignatureDocumentMinorVersion,
				SignatureReferenceType:      dbSignature.SignatureReferenceType,
				ProjectID:                   dbSignature.SignatureProjectID,
				Created:                     dbSignature.DateCreated,
				Modified:                    dbSignature.DateModified,
				UserName:                    userName,
				UserLFID:                    userLFID,
				UserGHID:                    userGHID,
				EmailWhitelist:              dbSignature.EmailWhitelist,
				DomainWhitelist:             dbSignature.DomainWhitelist,
				GithubWhitelist:             dbSignature.GitHubWhitelist,
				GithubOrgWhitelist:          dbSignature.GitHubOrgWhitelist,
				SignatureACL:                signatureACL,
			})
		}(dbSignature)
	}

	wg.Wait()

	return signatures, nil
}

// buildResponse is a helper function which converts a database model to a GitHub organization response model
func buildResponse(items []*dynamodb.AttributeValue) []models.GithubOrg {
	// Convert to a response model
	var orgs []models.GithubOrg
	for _, org := range items {
		selected := true
		orgs = append(orgs, models.GithubOrg{
			ID:       org.S,
			Selected: &selected,
		})
	}

	return orgs
}

// buildProject is a helper function to build a common set of projection/columns for the query
func buildProjection() expression.ProjectionBuilder {
	// These are the columns we want returned
	return expression.NamesList(
		expression.Name("signature_id"),
		expression.Name("date_created"),
		expression.Name("date_modified"),
		expression.Name("signature_acl"),
		expression.Name("signature_approved"),
		expression.Name("signature_document_major_version"),
		expression.Name("signature_document_minor_version"),
		expression.Name("signature_reference_id"),
		expression.Name("signature_reference_name"),       // Added to support simplified UX queries
		expression.Name("signature_reference_name_lower"), // Added to support case insensitive UX queries
		expression.Name("signature_project_id"),
		expression.Name("signature_reference_type"),       // user or company
		expression.Name("signature_signed"),               // T/F
		expression.Name("signature_type"),                 // ccla or cla
		expression.Name("signature_user_ccla_company_id"), // reference to the company
		expression.Name("email_whitelist"),
		expression.Name("domain_whitelist"),
		expression.Name("github_whitelist"),
		expression.Name("github_org_whitelist"),
	)
}

// func buildUsers(items []*dynamodb.AttributeValue) []models.User {
// 	var users []models.User
// 	for _, user := range items {
// 		users = append(users, models.User{
// 			UserID: user.S,
// 		})
// 	}

// 	return users
// }
