// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package signatures

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/go-openapi/strfmt"

	"github.com/sirupsen/logrus"

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

// constants
const (
	LoadACLDetails                                 = true
	DontLoadACLDetails                             = false
	SignatureProjectIDIndex                        = "project-signature-index"
	SignatureProjectReferenceIndex                 = "signature-project-reference-index"
	SignatureProjectIDSigTypeSignedApprovedIDIndex = "signature-project-id-sigtype-signed-approved-id-index"
	SignatureProjectIDTypeIndex                    = "signature-project-id-type-index"
	SignatureReferenceIndex                        = "reference-signature-index"

	// ReferenceTypeUser is the signature reference type for user signatures - individual and employee
	ReferenceTypeUser = "user"
	// ReferenceTypeCompany is the signature reference type for corporate signatures - signed by CLA Signatories, managed by CLA Managers
	ReferenceTypeCompany = "company"

	// SignatureTypeCLA is the cla signature type in the DB
	SignatureTypeCLA = "cla"
	// SignatureTypeCCLA is the ccla signature type in the DB
	SignatureTypeCCLA = "ccla"

	// ICLA represents individual contributor CLA records
	ICLA = "icla"
	// ECLA represents employee contributor CLA records (acknowledgements)
	ECLA = "ecla"
	// CCLA represents corporate CLA records (includes approval lists)
	CCLA = "ccla"

	HugePageSize = 10000
)

// SignatureRepository interface defines the functions for the github whitelist service
type SignatureRepository interface {
	GetGithubOrganizationsFromWhitelist(signatureID string) ([]models.GithubOrg, error)
	AddGithubOrganizationToWhitelist(signatureID, githubOrganizationID string) ([]models.GithubOrg, error)
	DeleteGithubOrganizationFromWhitelist(signatureID, githubOrganizationID string) ([]models.GithubOrg, error)
	InvalidateProjectRecord(signatureID string, projectName string) error

	GetSignature(signatureID string) (*models.Signature, error)
	GetIndividualSignature(claGroupID, userID string) (*models.Signature, error)
	GetCorporateSignature(claGroupID, companyID string) (*models.Signature, error)
	GetSignatureACL(signatureID string) ([]string, error)
	GetProjectSignatures(params signatures.GetProjectSignaturesParams, pageSize int64) (*models.Signatures, error)
	GetProjectCompanySignature(companyID, projectID string, signed, approved *bool, nextKey *string, pageSize *int64) (*models.Signature, error)
	GetProjectCompanySignatures(companyID, projectID string, signed, approved *bool, nextKey *string, pageSize *int64) (*models.Signatures, error)
	GetProjectCompanyEmployeeSignatures(params signatures.GetProjectCompanyEmployeeSignaturesParams, pageSize int64) (*models.Signatures, error)
	GetCompanySignatures(params signatures.GetCompanySignaturesParams, pageSize int64, loadACL bool) (*models.Signatures, error)
	GetCompanyIDsWithSignedCorporateSignatures(claGroupID string) ([]SignatureCompanyID, error)
	GetUserSignatures(params signatures.GetUserSignaturesParams, pageSize int64) (*models.Signatures, error)
	ProjectSignatures(projectID string) (*models.Signatures, error)
	UpdateApprovalList(projectID, companyID string, params *models.ApprovalList) (*models.Signature, error)

	AddCLAManager(signatureID, claManagerID string) (*models.Signature, error)
	RemoveCLAManager(signatureID, claManagerID string) (*models.Signature, error)

	removeColumn(signatureID, columnName string) (*models.Signature, error)

	AddSigTypeSignedApprovedID(signatureID string, val string) error
	AddUsersDetails(signatureID string, userID string) error
	AddSignedOn(signatureID string) error

	GetClaGroupICLASignatures(claGroupID string, searchTerm *string) (*models.IclaSignatures, error)
	GetClaGroupCorporateContributors(claGroupID string, companyID *string, searchTerm *string) (*models.CorporateContributorList, error)
}

// repository data model
type repository struct {
	stage              string
	dynamoDBClient     *dynamodb.DynamoDB
	companyRepo        company.IRepository
	usersRepo          users.UserRepository
	signatureTableName string
}

// NewRepository creates a new instance of the whitelist service
func NewRepository(awsSession *session.Session, stage string, companyRepo company.IRepository, usersRepo users.UserRepository) SignatureRepository {
	return repository{
		stage:              stage,
		dynamoDBClient:     dynamodb.New(awsSession),
		companyRepo:        companyRepo,
		usersRepo:          usersRepo,
		signatureTableName: fmt.Sprintf("cla-%s-signatures", stage),
	}
}

// GetGithubOrganizationsFromWhitelist returns a list of GH organizations stored in the whitelist
func (repo repository) GetGithubOrganizationsFromWhitelist(signatureID string) ([]models.GithubOrg, error) {
	// get item from dynamoDB table
	result, err := repo.dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(repo.signatureTableName),
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
	log.Debugf("querying database for github organization whitelist using signatureID: %s", signatureID)

	result, err := repo.dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(repo.signatureTableName),
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
		TableName: aws.String(repo.signatureTableName),
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
	result, err := repo.dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(repo.signatureTableName),
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
			TableName: aws.String(repo.signatureTableName),
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
		TableName: aws.String(repo.signatureTableName),
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

// GetSignature returns the signature for the specified signature id
func (repo repository) GetSignature(signatureID string) (*models.Signature, error) {
	// This is the key we want to match
	condition := expression.Key("signature_id").Equal(expression.Value(signatureID))

	// Use the builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(buildProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for signature ID query, signatureID: %s, error: %v",
			signatureID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(repo.signatureTableName),
	}

	// Make the DynamoDB Query API call
	results, queryErr := repo.dynamoDBClient.Query(queryInput)
	if queryErr != nil {
		log.Warnf("error retrieving signature ID: %s, error: %v", signatureID, queryErr)
		return nil, queryErr
	}

	// No match, didn't find it
	if *results.Count == 0 {
		return nil, nil
	}

	// Convert the list of DB models to a list of response models - should have zero or 1 given that we query by ID
	signatureList, modelErr := repo.buildProjectSignatureModels(results, "", LoadACLDetails)
	if modelErr != nil {
		log.Warnf("error converting DB model to response model for signature: %s, error: %v",
			signatureID, modelErr)
		return nil, modelErr
	}

	if len(signatureList) == 0 {
		return nil, nil
	}

	return signatureList[0], nil
}

// GetIndividualSignature returns the signature record for the specified CLA Group and User
func (repo repository) GetIndividualSignature(claGroupID, userID string) (*models.Signature, error) {
	f := logrus.Fields{
		"functionName":           "GetIndividualSignature",
		"tableName":              repo.signatureTableName,
		"claGroupID":             claGroupID,
		"userID":                 userID,
		"signatureType":          SignatureTypeCLA,
		"signatureReferenceType": ReferenceTypeUser,
		"signatureApproved":      "true",
		"signatureSigned":        "true",
	}

	// These are the keys we want to match for an ICLA Signature with a given CLA Group and User ID
	condition := expression.Key("signature_project_id").Equal(expression.Value(claGroupID)).
		And(expression.Key("signature_reference_id").Equal(expression.Value(userID)))
	filter := expression.Name("signature_type").Equal(expression.Value(SignatureTypeCLA)).
		And(expression.Name("signature_reference_type").Equal(expression.Value("user"))).
		And(expression.Name("signature_approved").Equal(expression.Value(aws.Bool(true)))).
		And(expression.Name("signature_signed").Equal(expression.Value(aws.Bool(true)))).
		And(expression.Name("signature_user_ccla_company_id").AttributeNotExists())

	builder := expression.NewBuilder().
		WithKeyCondition(condition).
		WithFilter(filter).
		WithProjection(buildProjection())

	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		log.WithFields(f).Warnf("error building expression for project ICLA signature query, error: %v", err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(repo.signatureTableName),
		Limit:                     aws.Int64(100),                             // The maximum number of items to evaluate (not necessarily the number of matching items)
		IndexName:                 aws.String(SignatureProjectReferenceIndex), // Name of a secondary index to scan
	}

	sigs := make([]*models.Signature, 0)
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		//log.Debugf("Running signature project query using queryInput: %+v", queryInput)
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		//log.Debugf("Ran signature project query, results: %+v, error: %+v", results, errQuery)
		if errQuery != nil {
			log.WithFields(f).Warnf("error retrieving project ICLA signature ID, error: %v", errQuery)
			return nil, errQuery
		}

		// Convert the list of DB models to a list of response models
		//log.Debug("Building response models...")
		signatureList, modelErr := repo.buildProjectSignatureModels(results, claGroupID, LoadACLDetails)
		if modelErr != nil {
			log.WithFields(f).Warnf("error converting DB model to response model for signatures, error: %v",
				modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		sigs = append(sigs, signatureList...)

		//log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey)
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = results.LastEvaluatedKey
		} else {
			lastEvaluatedKey = ""
		}
	}

	// Didn't find a matching record
	if len(sigs) == 0 {
		return nil, nil
	}

	if len(sigs) > 1 {
		log.WithFields(f).Warnf("found multiple matching ICLA signatures - found %d total", len(sigs))
	}

	return sigs[0], nil
}

// GetCorporateSignature returns the signature record for the specified CLA Group and Company ID
func (repo repository) GetCorporateSignature(claGroupID, companyID string) (*models.Signature, error) {
	f := logrus.Fields{
		"functionName":           "GetCorporateSignature",
		"tableName":              repo.signatureTableName,
		"claGroupID":             claGroupID,
		"companyID":              companyID,
		"signatureType":          "ccla",
		"signatureReferenceType": "company",
		"signatureApproved":      "true",
		"signatureSigned":        "true",
	}

	// These are the keys we want to match for an ICLA Signature with a given CLA Group and User ID
	condition := expression.Key("signature_project_id").Equal(expression.Value(claGroupID)).
		And(expression.Key("signature_reference_id").Equal(expression.Value(companyID)))
	filter := expression.Name("signature_type").Equal(expression.Value("ccla")).
		And(expression.Name("signature_reference_type").Equal(expression.Value("company"))).
		And(expression.Name("signature_approved").Equal(expression.Value(aws.Bool(true)))).
		And(expression.Name("signature_signed").Equal(expression.Value(aws.Bool(true)))).
		And(expression.Name("signature_user_ccla_company_id").AttributeNotExists())

	builder := expression.NewBuilder().
		WithKeyCondition(condition).
		WithFilter(filter).
		WithProjection(buildProjection())

	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		log.WithFields(f).Warnf("error building expression for project CCLA signature query, error: %v", err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(repo.signatureTableName),
		Limit:                     aws.Int64(100),                             // The maximum number of items to evaluate (not necessarily the number of matching items)
		IndexName:                 aws.String(SignatureProjectReferenceIndex), // Name of a secondary index to scan
	}

	sigs := make([]*models.Signature, 0)
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		//log.Debugf("Running signature project query using queryInput: %+v", queryInput)
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		//log.Debugf("Ran signature project query, results: %+v, error: %+v", results, errQuery)
		if errQuery != nil {
			log.WithFields(f).Warnf("error retrieving project CCLA signature, error: %v", errQuery)
			return nil, errQuery
		}

		// Convert the list of DB models to a list of response models
		//log.Debug("Building response models...")
		signatureList, modelErr := repo.buildProjectSignatureModels(results, claGroupID, LoadACLDetails)
		if modelErr != nil {
			log.WithFields(f).Warnf("error converting DB model to response model for signatures, error: %v",
				modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		sigs = append(sigs, signatureList...)

		//log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey)
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = results.LastEvaluatedKey
		} else {
			lastEvaluatedKey = ""
		}
	}

	// Didn't find a matching record
	if len(sigs) == 0 {
		return nil, nil
	}

	if len(sigs) > 1 {
		log.WithFields(f).Warnf("found multiple matching ICLA signatures - found %d total", len(sigs))
	}

	return sigs[0], nil
}

// GetSignatureACL returns the signature ACL for the specified signature id
func (repo repository) GetSignatureACL(signatureID string) ([]string, error) {
	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().
		WithProjection(buildSignatureACLProjection()).
		Build()
	if err != nil {
		log.Warnf("error building expression for signature ID query, signatureID: %s, error: %v",
			signatureID, err)
		return nil, err
	}

	// Assemble the query input parameters
	itemInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {S: aws.String(signatureID)},
		},
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
		TableName:                aws.String(repo.signatureTableName),
	}

	// Make the DynamoDB Query API call
	result, queryErr := repo.dynamoDBClient.GetItem(itemInput)
	if queryErr != nil {
		log.Warnf("error retrieving signature ID: %s, error: %v", signatureID, queryErr)
		return nil, queryErr
	}

	// No match, didn't find it
	if result.Item == nil {
		return nil, nil
	}

	var dbModel DBManagersModel
	// Unmarshall the DB response
	unmarshallErr := dynamodbattribute.UnmarshalMap(result.Item, &dbModel)
	if unmarshallErr != nil {
		log.Warnf("error converting DB model signature query using siganture ID: %s, error: %v",
			signatureID, unmarshallErr)
		return nil, unmarshallErr
	}

	return dbModel.SignatureACL, nil
}

func addConditionToFilter(filter expression.ConditionBuilder, cond expression.ConditionBuilder, filterAdded *bool) expression.ConditionBuilder {
	if !(*filterAdded) {
		*filterAdded = true
		filter = cond
	} else {
		filter = filter.And(cond)
	}
	return filter
}

// GetProjectSignatures returns a list of signatures for the specified project
func (repo repository) GetProjectSignatures(params signatures.GetProjectSignaturesParams, pageSize int64) (*models.Signatures, error) {

	indexName := SignatureProjectIDIndex

	// This is the key we want to match
	condition := expression.Key("signature_project_id").Equal(expression.Value(params.ProjectID))

	builder := expression.NewBuilder().WithProjection(buildProjection())
	var filter expression.ConditionBuilder
	var filterAdded bool

	if params.ClaType != nil {
		filterAdded = true
		if strings.ToLower(*params.ClaType) == ICLA {
			filter = expression.Name("signature_type").Equal(expression.Value(SignatureTypeCLA)).
				And(expression.Name("signature_reference_type").Equal(expression.Value(ReferenceTypeUser))).
				And(expression.Name("signature_approved").Equal(expression.Value(aws.Bool(true)))).
				And(expression.Name("signature_signed").Equal(expression.Value(aws.Bool(true)))).
				And(expression.Name("signature_user_ccla_company_id").AttributeNotExists())

		} else if strings.ToLower(*params.ClaType) == ECLA {
			filter = expression.Name("signature_type").Equal(expression.Value(SignatureTypeCLA)).
				And(expression.Name("signature_reference_type").Equal(expression.Value(ReferenceTypeUser))).
				And(expression.Name("signature_approved").Equal(expression.Value(aws.Bool(true)))).
				And(expression.Name("signature_signed").Equal(expression.Value(aws.Bool(true)))).
				And(expression.Name("signature_user_ccla_company_id").AttributeExists())
		} else if strings.ToLower(*params.ClaType) == CCLA {
			filter = expression.Name("signature_type").Equal(expression.Value(SignatureTypeCCLA)).
				And(expression.Name("signature_reference_type").Equal(expression.Value(ReferenceTypeCompany))).
				And(expression.Name("signature_approved").Equal(expression.Value(aws.Bool(true)))).
				And(expression.Name("signature_signed").Equal(expression.Value(aws.Bool(true)))).
				And(expression.Name("signature_user_ccla_company_id").AttributeNotExists())
		}
	} else {
		if params.SearchField != nil {
			searchFieldExpression := expression.Name("signature_reference_type").Equal(expression.Value(params.SearchField))
			filter = addConditionToFilter(filter, searchFieldExpression, &filterAdded)
		}

		if params.SignatureType != nil {
			if params.SearchTerm != nil && (params.FullMatch != nil && !*params.FullMatch) {
				indexName = SignatureProjectIDTypeIndex
				condition = condition.And(expression.Key("signature_type").Equal(expression.Value(strings.ToLower(*params.SignatureType))))
			} else {
				signatureTypeExpression := expression.Name("signature_type").Equal(expression.Value(params.SignatureType))
				filter = addConditionToFilter(filter, signatureTypeExpression, &filterAdded)
			}
			if *params.SignatureType == "ccla" {
				signatureReferenceIDExpression := expression.Name("signature_reference_id").AttributeExists()
				signatureUserCclaCompanyIDExpression := expression.Name("signature_user_ccla_company_id").AttributeNotExists()
				filter = addConditionToFilter(filter, signatureReferenceIDExpression, &filterAdded)
				filter = addConditionToFilter(filter, signatureUserCclaCompanyIDExpression, &filterAdded)
			}
		}

		if params.SearchTerm != nil {
			if *params.FullMatch {
				indexName = "reference-signature-search-index"
				condition = condition.And(expression.Key("signature_reference_name_lower").Equal(expression.Value(strings.ToLower(*params.SearchTerm))))
			} else {
				searchTermExpression := expression.Name("signature_reference_name_lower").Contains(strings.ToLower(*params.SearchTerm))
				filter = addConditionToFilter(filter, searchTermExpression, &filterAdded)
			}
		}

		// Filter condition to cater for approved and signed signatures
		signatureApprovedExpression := expression.Name("signature_approved").Equal(expression.Value(true))
		filter = addConditionToFilter(filter, signatureApprovedExpression, &filterAdded)

		signatureSignedExpression := expression.Name("signature_signed").Equal(expression.Value(true))
		filter = addConditionToFilter(filter, signatureSignedExpression, &filterAdded)
	}

	if filterAdded {
		builder = builder.WithFilter(filter)
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
		TableName:                 aws.String(repo.signatureTableName),
		Limit:                     aws.Int64(pageSize),   // The maximum number of items to evaluate (not necessarily the number of matching items)
		IndexName:                 aws.String(indexName), // Name of a secondary index to scan
	}

	// If we have the next key, set the exclusive start key value
	if params.NextKey != nil {
		log.Debugf("received a nextKey, value: %s", *params.NextKey)
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
		if params.FullMatch != nil && *params.FullMatch && params.SearchTerm != nil {
			queryInput.ExclusiveStartKey["signature_reference_name_lower"] = &dynamodb.AttributeValue{
				S: params.SearchTerm,
			}
		}
	}

	sigs := make([]*models.Signature, 0)
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

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, params.ProjectID, LoadACLDetails)
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for signatures with project %s, error: %v",
				params.ProjectID, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		sigs = append(sigs, signatureList...)

		//log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey)
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = results.LastEvaluatedKey
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(sigs)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &repo.signatureTableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count for project: %s, error: %v", params.ProjectID, err)
		return nil, err
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount
	if int64(len(sigs)) > pageSize {
		sigs = sigs[0:pageSize]
		lastEvaluatedKey = sigs[pageSize-1].SignatureID.String()
	}

	return &models.Signatures{
		ProjectID:      params.ProjectID,
		ResultCount:    int64(len(sigs)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     sigs,
	}, nil
}

// GetProjectCompanySignature returns a the signature for the specified project and specified company with the other query flags
func (repo repository) GetProjectCompanySignature(companyID, projectID string, signed, approved *bool, nextKey *string, pageSize *int64) (*models.Signature, error) {
	sigs, getErr := repo.GetProjectCompanySignatures(companyID, projectID, signed, approved, nextKey, pageSize)
	if getErr != nil {
		return nil, getErr
	}

	if sigs == nil || sigs.Signatures == nil {
		return nil, nil
	}

	if len(sigs.Signatures) > 1 {
		log.Warnf("more than 1 project company signatures returned in result using company ID: %s, project ID: %s - will return fist record",
			companyID, projectID)
	}

	return sigs.Signatures[0], nil
}

// GetProjectCompanySignatures returns a list of signatures for the specified project and specified company
func (repo repository) GetProjectCompanySignatures(companyID, projectID string, signed, approved *bool, nextKey *string, pageSize *int64) (*models.Signatures, error) {

	// These are the keys we want to match
	condition := expression.Key("signature_project_id").Equal(expression.Value(projectID))
	filter := expression.Name("signature_reference_id").Equal(expression.Value(companyID)).
		And(expression.Name("signature_type").Equal(expression.Value("ccla"))).
		And(expression.Name("signature_reference_type").Equal(expression.Value("company")))

	// If the caller provided a signature signed value...add the appropriate filter
	if signed != nil {
		log.Debugf("Filtering signature_signed: %+v", *signed)
		filter = filter.And(expression.Name("signature_signed").Equal(expression.Value(aws.Bool(*signed))))
	}

	// If the caller provided a signature approved value...add the appropriate filter
	if approved != nil {
		log.Debugf("Filter by signature_approved: %+v", *approved)
		filter = filter.And(expression.Name("signature_approved").Equal(expression.Value(aws.Bool(*approved))))
	}

	limit := int64(10)
	if pageSize != nil {
		limit = *pageSize
	}

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithFilter(filter).WithProjection(buildProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for project signature ID query, project: %s, error: %v",
			projectID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(repo.signatureTableName),
		IndexName:                 aws.String("project-signature-index"), // Name of a secondary index to scan
		Limit:                     aws.Int64(limit),
	}

	// If we have the next key, set the exclusive start key value
	if nextKey != nil {
		log.Debugf("Received a nextKey, value: %s", *nextKey)
		// The primary key of the first item that this operation will evaluate.
		// and the query key (if not the same)
		queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: nextKey,
			},
			"signature_project_id": {
				S: &projectID,
			},
		}
	}

	var sigs []*models.Signature
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		if errQuery != nil {
			log.Warnf("error retrieving project signature ID for project: %s with company: %s, error: %v",
				projectID, companyID, errQuery)
			return nil, errQuery
		}

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, projectID, LoadACLDetails)
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for signatures with project %s with company: %s, error: %v",
				projectID, companyID, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		sigs = append(sigs, signatureList...)

		// log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"signature_id": {
					S: aws.String(lastEvaluatedKey),
				},
				"signature_project_id": {
					S: &projectID,
				},
			}
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(sigs)) >= limit {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &repo.signatureTableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count for project: %s, error: %v", projectID, err)
		return nil, err
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount

	return &models.Signatures{
		ProjectID:      projectID,
		ResultCount:    int64(len(sigs)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     sigs,
	}, nil
}

// Get project signatures with no pagination
func (repo repository) ProjectSignatures(projectID string) (*models.Signatures, error) {

	indexName := SignatureProjectIDIndex

	// This is the key we want to match
	condition := expression.Key("signature_project_id").Equal(expression.Value(projectID))

	builder := expression.NewBuilder().WithProjection(buildProjection())
	var filter expression.ConditionBuilder
	var filterAdded bool

	// Filter condition to cater for approved and signed signatures
	signatureApprovedExpression := expression.Name("signature_approved").Equal(expression.Value(true))
	filter = addConditionToFilter(filter, signatureApprovedExpression, &filterAdded)

	signatureSignedExpression := expression.Name("signature_signed").Equal(expression.Value(true))
	filter = addConditionToFilter(filter, signatureSignedExpression, &filterAdded)

	if filterAdded {
		builder = builder.WithFilter(filter)
	}
	builder = builder.WithKeyCondition(condition)

	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		log.Warnf("error building expression for project signature query, projectID: %s, error: %v",
			projectID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(repo.signatureTableName),
		IndexName:                 aws.String(indexName), // Name of a secondary index to scan
	}

	results, errQuery := repo.dynamoDBClient.Query(queryInput)

	if errQuery != nil {
		log.Warnf("error retrieving project signature ID for project: %s, error: %v",
			projectID, errQuery)
		return nil, errQuery
	}

	// Convert the list of DB models to a list of response models
	sigs, modelErr := repo.buildProjectSignatureModels(results, projectID, LoadACLDetails)
	if modelErr != nil {
		log.Warnf("error converting DB model to response model for signatures with project %s, error: %v",
			projectID, modelErr)
		return nil, modelErr
	}

	return &models.Signatures{
		ProjectID:  projectID,
		Signatures: sigs,
	}, nil
}

// InvalidateProjectRecord invalidates the specified project record by setting the signature_approved flag to false
func (repo repository) InvalidateProjectRecord(signatureID string, projectName string) error {
	// Update project signatures for signature_approved and notes attributes
	signatureTableName := fmt.Sprintf("cla-%s-signatures", repo.stage)

	expressionAttributeNames := map[string]*string{}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{}
	updateExpression := "SET " // nolint

	expressionAttributeNames["#A"] = aws.String("signature_approved")
	expressionAttributeValues[":a"] = &dynamodb.AttributeValue{BOOL: aws.Bool(false)}
	updateExpression = updateExpression + " #A = :a,"

	expressionAttributeNames["#S"] = aws.String("note")
	note := fmt.Sprintf("Signature invalidated (approved set to false) due to CLA Group/Project: %s deletion", projectName)
	expressionAttributeValues[":s"] = &dynamodb.AttributeValue{S: aws.String(note)}
	updateExpression = updateExpression + " #S = :s"

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		UpdateExpression:          &updateExpression,
		TableName:                 aws.String(signatureTableName),
	}

	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.Warnf("error updating signature_approved for signature_id : %s error : %v ", signatureID, updateErr)
		return updateErr
	}

	return nil
}

// GetProjectCompanyEmployeeSignatures returns a list of employee signatures for the specified project and specified company
func (repo repository) GetProjectCompanyEmployeeSignatures(params signatures.GetProjectCompanyEmployeeSignaturesParams, pageSize int64) (*models.Signatures, error) {

	// This is the keys we want to match
	condition := expression.Key("signature_user_ccla_company_id").Equal(expression.Value(params.CompanyID)).And(
		expression.Key("signature_project_id").Equal(expression.Value(params.ProjectID)))
	// Check for approved signatures
	filter := expression.Name("signature_approved").Equal(expression.Value(aws.Bool(true))).
		And(expression.Name("signature_signed").Equal(expression.Value(aws.Bool(true))))

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
		TableName:                 aws.String(repo.signatureTableName),
		IndexName:                 aws.String("signature-user-ccla-company-index"), // Name of a secondary index to scan
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
			"signature_user_ccla_company_id": {
				S: &params.CompanyID,
			},
			"signature_project_id": {
				S: &params.ProjectID,
			},
		}
	}

	sigs := make([]*models.Signature, 0)
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

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, params.ProjectID, LoadACLDetails)
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for employee signatures with project %s with company: %s, error: %v",
				params.ProjectID, params.CompanyID, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		sigs = append(sigs, signatureList...)

		// log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = results.LastEvaluatedKey
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(sigs)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &repo.signatureTableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count for project: %s, error: %v", params.ProjectID, err)
		return nil, err
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount
	if int64(len(sigs)) > pageSize {
		sigs = sigs[0:pageSize]
		lastEvaluatedKey = sigs[pageSize-1].SignatureID.String()
	}

	return &models.Signatures{
		ProjectID:      params.ProjectID,
		ResultCount:    int64(len(sigs)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     sigs,
	}, nil
}

// GetCompanySignatures returns a list of company signatures for the specified company
func (repo repository) GetCompanySignatures(params signatures.GetCompanySignaturesParams, pageSize int64, loadACL bool) (*models.Signatures, error) {

	// This is the keys we want to match
	condition := expression.Key("signature_reference_id").Equal(expression.Value(params.CompanyID))

	// Check for approved signatures
	filter := expression.Name("signature_approved").Equal(expression.Value(aws.Bool(true))).
		And(expression.Name("signature_signed").Equal(expression.Value(aws.Bool(true))))

	if params.SignatureType != nil {
		filter = filter.And(expression.Name("signature_type").Equal(expression.Value(*params.SignatureType)))
	}

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithFilter(filter).WithProjection(buildProjection()).Build()
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
		TableName:                 aws.String(repo.signatureTableName),
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

	sigs := make([]*models.Signature, 0)
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

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, "", loadACL)
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for signatures with company: %s, error: %v",
				params.CompanyID, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		sigs = append(sigs, signatureList...)

		// log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = results.LastEvaluatedKey
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(sigs)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &repo.signatureTableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count for company: %s/%s, error: %v",
			params.CompanyID, *params.CompanyName, err)
		return nil, err
	}
	if int64(len(sigs)) > pageSize {
		sigs = sigs[0:pageSize]
		lastEvaluatedKey = sigs[pageSize-1].SignatureID.String()
	}

	// Meta-data for the response
	totalCount := *describeTableResult.Table.ItemCount

	return &models.Signatures{
		ProjectID:      "",
		ResultCount:    int64(len(sigs)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     sigs,
	}, nil
}

// GetCompanyIDsWithSignedCorporateSignatures returns a list of company IDs that have signed a CLA agreement
func (repo repository) GetCompanyIDsWithSignedCorporateSignatures(claGroupID string) ([]SignatureCompanyID, error) {
	f := logrus.Fields{
		"functionName":             "GetCompanyIDsWithSignedCorporateSignatures",
		"claGroupID":               claGroupID,
		"signature_project_id":     claGroupID,
		"signature_type":           "ccla",
		"signature_reference_type": "company",
		"signature_signed":         "true",
		"signature_approved":       "true",
		"tableName":                repo.signatureTableName,
		"stage":                    repo.stage,
	}

	// These are the keys we want to match
	condition := expression.Key("signature_project_id").Equal(expression.Value(claGroupID))
	filter := expression.Name("signature_type").Equal(expression.Value("ccla")).
		And(expression.Name("signature_reference_type").Equal(expression.Value("company"))).
		And(expression.Name("signature_signed").Equal(expression.Value(aws.Bool(true)))).
		And(expression.Name("signature_approved").Equal(expression.Value(aws.Bool(true))))

	// Batch size
	limit := int64(100)

	// Use the nice builder to create the expression - this one uses a simple projection with only the signature id (required) and company id - which is the signature reference id field
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithFilter(filter).WithProjection(buildCompanyIDProjection()).Build()
	if err != nil {
		log.WithFields(f).Warnf("error building expression, error: %v", err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(repo.signatureTableName),
		IndexName:                 aws.String("project-signature-index"), // Name of a secondary index to scan
		Limit:                     aws.Int64(limit),
	}

	var companyIDs []SignatureCompanyID
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		if errQuery != nil {
			log.Warnf("error retrieving signature record, error: %v", errQuery)
			return nil, errQuery
		}

		companyIDList, buildErr := repo.buildCompanyIDList(results)
		if buildErr != nil {
			log.WithFields(f).Warnf("problem converting db model to list of company IDs, error: %+v", buildErr)
			return nil, buildErr
		}

		// Convert the list of DB models to a list of response models
		companyIDs = append(companyIDs, companyIDList...)

		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"signature_id": {
					S: aws.String(lastEvaluatedKey),
				},
				"signature_project_id": {
					S: &claGroupID,
				},
			}
		} else {
			lastEvaluatedKey = ""
		}
	}

	return companyIDs, nil
}

// GetUserSignatures returns a list of user signatures for the specified user
func (repo repository) GetUserSignatures(params signatures.GetUserSignaturesParams, pageSize int64) (*models.Signatures, error) {

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
		TableName:                 aws.String(repo.signatureTableName),
		IndexName:                 aws.String(SignatureReferenceIndex), // Name of a secondary index to scan
		Limit:                     aws.Int64(pageSize),                 // The maximum number of items to evaluate (not necessarily the number of matching items)
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

	sigs := make([]*models.Signature, 0)
	var lastEvaluatedKey string

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		results, errQuery := repo.dynamoDBClient.Query(queryInput)
		if errQuery != nil {
			log.Warnf("error retrieving user signatures for user: %s/%s, error: %v",
				params.UserID, *params.UserName, errQuery)
			return nil, errQuery
		}

		// Convert the list of DB models to a list of response models
		signatureList, modelErr := repo.buildProjectSignatureModels(results, "", LoadACLDetails)
		if modelErr != nil {
			log.Warnf("error converting DB model to response model for signatures for user %s/%s, error: %v",
				params.UserID, *params.UserName, modelErr)
			return nil, modelErr
		}

		// Add to the signatures response model to the list
		sigs = append(sigs, signatureList...)

		// log.Debugf("LastEvaluatedKey: %+v", results.LastEvaluatedKey["signature_id"])
		if results.LastEvaluatedKey["signature_id"] != nil {
			lastEvaluatedKey = *results.LastEvaluatedKey["signature_id"].S
			queryInput.ExclusiveStartKey = results.LastEvaluatedKey
		} else {
			lastEvaluatedKey = ""
		}

		if int64(len(sigs)) >= pageSize {
			break
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &repo.signatureTableName,
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
		ResultCount:    int64(len(sigs)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Signatures:     sigs,
	}, nil
}

func (repo repository) AddCLAManager(signatureID, claManagerID string) (*models.Signature, error) {
	aclEntries, err := repo.GetSignatureACL(signatureID)
	if err != nil {
		log.Warnf("unable to fetch signature by ID: %s, error: %+v", signatureID, err)
		return nil, err
	}

	if aclEntries == nil {
		log.Warnf("unable to fetch signature by ID: %s - record not found", signatureID)
		return nil, nil
	}

	for _, manager := range aclEntries {
		if claManagerID == manager {
			return nil, errors.New("manager already in signature ACL")
		}
	}

	aclEntries = append(aclEntries, claManagerID)
	log.Debugf("To be updated acllist : %+v", aclEntries)

	_, now := utils.CurrentTime()

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#A": aws.String("signature_acl"),
			"#M": aws.String("date_modified"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				SS: aws.StringSlice(aclEntries),
			},
			":m": {
				S: aws.String(now),
			},
		},
		UpdateExpression: aws.String("SET #A = :a, #M = :m"),
		TableName:        aws.String(fmt.Sprintf("cla-%s-signatures", repo.stage)),
	}

	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.Warnf("add CLA manager - unable to update request with new ACL entry of '%s' for signature ID: %s, error: %v",
			claManagerID, signatureID, updateErr)
		return nil, updateErr
	}

	// Load the updated document and return it
	sigModel, err := repo.GetSignature(signatureID)
	if err != nil {
		log.Warnf("unable to fetch signature by ID: %s - record not found", signatureID)
		return nil, err
	}

	return sigModel, nil
}

func (repo repository) RemoveCLAManager(signatureID, claManagerID string) (*models.Signature, error) {
	aclEntries, err := repo.GetSignatureACL(signatureID)
	if err != nil {
		log.Warnf("unable to fetch signature by ID: %s, error: %+v", signatureID, err)
		return nil, err
	}

	if aclEntries == nil {
		log.Warnf("unable to fetch signature by ID: %s - record not found", signatureID)
		return nil, nil
	}

	// A bit of logic to determine if the manager is listed and to build the new list without the specified manager
	found := false
	var updateEntries []string
	for _, manager := range aclEntries {
		if claManagerID == manager {
			found = true
		} else {
			updateEntries = append(updateEntries, manager)
		}
	}

	if !found {
		return nil, fmt.Errorf("manager ID: %s not found in signature ACL", claManagerID)
	}

	_, now := utils.CurrentTime()

	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#A": aws.String("signature_acl"),
			"#M": aws.String("date_modified"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":a": {
				SS: aws.StringSlice(updateEntries),
			},
			":m": {
				S: aws.String(now),
			},
		},
		UpdateExpression: aws.String("SET #A = :a, #M = :m"),
		TableName:        aws.String(fmt.Sprintf("cla-%s-signatures", repo.stage)),
	}

	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.Warnf("remove CLA manager - unable to remove ACL entry of '%s' for signature ID: %s, error: %v",
			claManagerID, signatureID, updateErr)
		return nil, updateErr
	}

	// Load the updated document and return it
	sigModel, err := repo.GetSignature(signatureID)
	if err != nil {
		log.Warnf("unable to fetch signature by ID: %s - record not found", signatureID)
		return nil, err
	}

	return sigModel, nil
}

// UpdateApprovalList updates the specified project/company signature with the updated approval list information
func (repo repository) UpdateApprovalList(projectID, companyID string, params *models.ApprovalList) (*models.Signature, error) { // nolint
	log.Debugf("querying database for approval list details using project ID: %s, company ID: %s", projectID, companyID)

	signed, approved := true, true
	pageSize := int64(10)
	log.Debugf("querying database for approval list details using company ID: %s project ID: %s, type: ccla, signed: true, approved: true",
		companyID, projectID)
	sigs, sigErr := repo.GetProjectCompanySignatures(companyID, projectID, &signed, &approved, nil, &pageSize)
	if sigErr != nil {
		return nil, sigErr
	}

	if sigs == nil || sigs.Signatures == nil {
		msg := fmt.Sprintf("unable to locate signature for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t",
			companyID, projectID, signed, approved)
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	if len(sigs.Signatures) > 1 {
		log.Warnf("more than 1 CCLA signature returned for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t - expecting zero or 1 - using first record",
			companyID, projectID, signed, approved)
	}

	// Just grab and use the first one - need to figure out conflict resolution if more than one
	sig := sigs.Signatures[0]
	expressionAttributeNames := map[string]*string{}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{}
	haveAdditions := false
	updateExpression := ""

	// If we have an add or remove email list...we need to run an update for this column
	if params.AddEmailApprovalList != nil || params.RemoveEmailApprovalList != nil {
		columnName := "email_whitelist"
		attrList := buildApprovalAttributeList(sig.EmailApprovalList, params.AddEmailApprovalList, params.RemoveEmailApprovalList)
		// If no entries after consolidating all the updates, we need to remove the column
		if attrList == nil || attrList.L == nil {
			var rmColErr error
			sig, rmColErr = repo.removeColumn(sig.SignatureID.String(), columnName)
			if rmColErr != nil {
				msg := fmt.Sprintf("unable to remove column %s for signature for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t",
					columnName, companyID, projectID, signed, approved)
				log.Warn(msg)
				return nil, errors.New(msg)
			}
		} else {
			haveAdditions = true
			expressionAttributeNames["#E"] = aws.String("email_whitelist")
			expressionAttributeValues[":e"] = attrList
			updateExpression = updateExpression + " #E = :e, "
		}
	}

	if params.AddDomainApprovalList != nil || params.RemoveDomainApprovalList != nil {
		columnName := "domain_whitelist"
		attrList := buildApprovalAttributeList(sig.DomainApprovalList, params.AddDomainApprovalList, params.RemoveDomainApprovalList)
		// If no entries after consolidating all the updates, we need to remove the column
		if attrList == nil || attrList.L == nil {
			var rmColErr error
			sig, rmColErr = repo.removeColumn(sig.SignatureID.String(), columnName)
			if rmColErr != nil {
				msg := fmt.Sprintf("unable to remove column %s for signature for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t",
					columnName, companyID, projectID, signed, approved)
				log.Warn(msg)
				return nil, errors.New(msg)
			}
		} else {
			haveAdditions = true
			expressionAttributeNames["#D"] = aws.String(columnName)
			expressionAttributeValues[":d"] = attrList
			updateExpression = updateExpression + " #D = :d, "
		}
	}

	if params.AddGithubUsernameApprovalList != nil || params.RemoveGithubUsernameApprovalList != nil {
		columnName := "github_whitelist"
		attrList := buildApprovalAttributeList(sig.GithubUsernameApprovalList, params.AddGithubUsernameApprovalList, params.RemoveGithubUsernameApprovalList)
		// If no entries after consolidating all the updates, we need to remove the column
		if attrList == nil || attrList.L == nil {
			var rmColErr error
			sig, rmColErr = repo.removeColumn(sig.SignatureID.String(), columnName)
			if rmColErr != nil {
				msg := fmt.Sprintf("unable to remove column %s for signature for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t",
					columnName, companyID, projectID, signed, approved)
				log.Warn(msg)
				return nil, errors.New(msg)
			}
		} else {
			haveAdditions = true
			expressionAttributeNames["#G"] = aws.String(columnName)
			expressionAttributeValues[":g"] = attrList
			updateExpression = updateExpression + " #G = :g, "
		}
	}

	if params.AddGithubOrgApprovalList != nil || params.RemoveGithubOrgApprovalList != nil {
		columnName := "github_org_whitelist"
		attrList := buildApprovalAttributeList(sig.GithubOrgApprovalList, params.AddGithubOrgApprovalList, params.RemoveGithubOrgApprovalList)
		// If no entries after consolidating all the updates, we need to remove the column
		if attrList == nil || attrList.L == nil {
			var rmColErr error
			sig, rmColErr = repo.removeColumn(sig.SignatureID.String(), columnName)
			if rmColErr != nil {
				msg := fmt.Sprintf("unable to remove column %s for signature for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t",
					columnName, companyID, projectID, signed, approved)
				log.Warn(msg)
				return nil, errors.New(msg)
			}
		} else {
			haveAdditions = true
			expressionAttributeNames["#GO"] = aws.String("github_org_whitelist")
			expressionAttributeValues[":go"] = attrList
			updateExpression = updateExpression + " #GO = :go, "
		}
	}

	// Ensure at least one value is set for us to update
	if !haveAdditions {
		log.Debugf("no updates required to any of the approved list values company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t - expecting at least something to update",
			companyID, projectID, signed, approved)
		return sig, nil
	}

	// Remove trailing comma from the expression, if present
	updateExpression = utils.TrimRemoveTrailingComma("SET " + updateExpression)

	// Update dynamoDB table
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.signatureTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(sig.SignatureID.String()),
			},
		},
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		UpdateExpression:          aws.String(updateExpression), //aws.String("SET #L = :l"),
	}

	log.Debugf("updating approval list for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t",
		companyID, projectID, signed, approved)

	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.Warnf("error updating approval lists for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t, error: %v",
			companyID, projectID, signed, approved, updateErr)
		return nil, updateErr
	}

	log.Debugf("querying database for approval list details after update using company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t",
		companyID, projectID, signed, approved)

	updatedSig, sigErr := repo.GetProjectCompanySignatures(companyID, projectID, &signed, &approved, nil, &pageSize)
	if sigErr != nil {
		return nil, sigErr
	}

	if updatedSig == nil || updatedSig.Signatures == nil {
		msg := fmt.Sprintf("unable to locate signature after update for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t",
			companyID, projectID, signed, approved)
		log.Warn(msg)
		return nil, errors.New(msg)
	}

	if len(updatedSig.Signatures) > 1 {
		log.Warnf("more than 1 CCLA signature returned after update for company ID: %s project ID: %s, type: ccla, signed: %t, approved: %t - expecting zero or 1 - using first record",
			companyID, projectID, signed, approved)
	}

	// Just grab and use the first one - need to figure out conflict resolution if more than one
	return updatedSig.Signatures[0], nil
}

// removeColumn is a helper function to remove a given column when we need to zero out the column value - typically the approval list
func (repo repository) removeColumn(signatureID, columnName string) (*models.Signature, error) {
	log.Debugf("removing column %s from signature ID: %s", columnName, signatureID)

	// Update dynamoDB table
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.signatureTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#" + columnName: aws.String(columnName),
		},
		//ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
		//	":a": {
		//		S: aws.String("bar"),
		//	},
		//},
		UpdateExpression: aws.String("REMOVE #" + columnName), //aws.String("REMOVE github_org_whitelist"),
		ReturnValues:     aws.String(dynamodb.ReturnValueNone),
	}

	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.Warnf("error removing approval lists column %s for signature ID: %s, error: %v", columnName, signatureID, updateErr)
		return nil, updateErr
	}

	updatedSig, sigErr := repo.GetSignature(signatureID)
	if sigErr != nil {
		return nil, sigErr
	}

	return updatedSig, nil
}

func (repo repository) AddSigTypeSignedApprovedID(signatureID string, val string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.signatureTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#signature_project_id_skey": aws.String("sigtype_signed_approved_id"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val": {
				S: aws.String(val),
			},
		},
		UpdateExpression: aws.String("SET #signature_project_id_skey = :val"),
	}
	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.Warnf("unable to update sigtype_signed_approved_id for signature_id : %s", signatureID)
		return updateErr
	}
	return nil
}
func (repo repository) AddUsersDetails(signatureID string, userID string) error {
	userModel, err := repo.usersRepo.GetUser(userID)
	if err != nil {
		return err
	}
	if userModel == nil {
		log.WithFields(logrus.Fields{"user_id": userID, "signature_id": signatureID}).Error("invalid user_id")
		return fmt.Errorf("invalid user id : %s for signature : %s", userID, signatureID)
	}
	var email string
	if userModel.LfEmail != "" {
		email = userModel.LfEmail
	} else {
		if len(userModel.Emails) > 0 {
			email = userModel.Emails[0]
		}
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.signatureTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
	}
	ue := utils.NewDynamoUpdateExpression()
	ue.AddAttributeName("#gh_username", "user_github_username", userModel.GithubUsername != "")
	ue.AddAttributeName("#lf_username", "user_lf_username", userModel.LfUsername != "")
	ue.AddAttributeName("#name", "user_name", userModel.Username != "")
	ue.AddAttributeName("#email", "user_email", email != "")

	ue.AddAttributeValue(":gh_username", &dynamodb.AttributeValue{S: aws.String(userModel.GithubUsername)}, userModel.GithubUsername != "")
	ue.AddAttributeValue(":lf_username", &dynamodb.AttributeValue{S: aws.String(userModel.LfUsername)}, userModel.LfUsername != "")
	ue.AddAttributeValue(":name", &dynamodb.AttributeValue{S: aws.String(userModel.Username)}, userModel.Username != "")
	ue.AddAttributeValue(":email", &dynamodb.AttributeValue{S: aws.String(email)}, email != "")

	ue.AddUpdateExpression("#gh_username = :gh_username", userModel.GithubUsername != "")
	ue.AddUpdateExpression("#lf_username = :lf_username", userModel.LfUsername != "")
	ue.AddUpdateExpression("#name = :name", userModel.Username != "")
	ue.AddUpdateExpression("#email = :email", email != "")
	if ue.Expression == "" {
		// nothing to update
		return nil
	}
	input.UpdateExpression = aws.String(ue.Expression)
	input.ExpressionAttributeNames = ue.ExpressionAttributeNames
	input.ExpressionAttributeValues = ue.ExpressionAttributeValues
	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.Debugf("update input: %v", input)
		log.Warnf("unable to add users details to : %s . error = %s", signatureID, updateErr.Error())
		return updateErr
	}
	return nil
}

func (repo repository) AddSignedOn(signatureID string) error {
	_, currentTime := utils.CurrentTime()
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.signatureTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"signature_id": {
				S: aws.String(signatureID),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#signed_on": aws.String("signed_on"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":current_time": {
				S: aws.String(currentTime),
			},
		},
		UpdateExpression: aws.String("SET #signed_on = :current_time"),
	}
	_, updateErr := repo.dynamoDBClient.UpdateItem(input)
	if updateErr != nil {
		log.Debugf("update input: %v", input)
		log.Warnf("unable to signed_on to : %s . error = %s", signatureID, updateErr.Error())
		return updateErr
	}
	return nil
}

// buildProjectSignatureModels converts the response model into a response data model
func (repo repository) buildProjectSignatureModels(results *dynamodb.QueryOutput, projectID string, loadACLDetails bool) ([]*models.Signature, error) {
	var sigs []*models.Signature

	// The DB signature model
	var dbSignatures []ItemSignature

	err := dynamodbattribute.UnmarshalListOfMaps(results.Items, &dbSignatures)
	if err != nil {
		log.Warnf("error unmarshalling signatures from database for project: %s, error: %v",
			projectID, err)
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(len(dbSignatures))
	for _, dbSignature := range dbSignatures {

		// Set the signature type in the response
		var claType = ""
		// Corporate Signature
		if dbSignature.SignatureReferenceType == ReferenceTypeCompany && dbSignature.SignatureType == SignatureTypeCCLA {
			claType = CCLA
		}
		// Employee Signature
		if dbSignature.SignatureReferenceType == ReferenceTypeUser && dbSignature.SignatureType == SignatureTypeCLA && dbSignature.SignatureUserCompanyID != "" {
			claType = ECLA
		}

		// Individual Signature
		if dbSignature.SignatureReferenceType == ReferenceTypeUser && dbSignature.SignatureType == SignatureTypeCLA && dbSignature.SignatureUserCompanyID == "" {
			claType = ICLA
		}

		sig := &models.Signature{
			SignatureID:                 strfmt.UUID4(dbSignature.SignatureID),
			ClaType:                     claType,
			SignatureCreated:            dbSignature.DateCreated,
			SignatureModified:           dbSignature.DateModified,
			SignatureType:               dbSignature.SignatureType,
			SignatureReferenceID:        strfmt.UUID4(dbSignature.SignatureReferenceID),
			SignatureReferenceName:      dbSignature.SignatureReferenceName,
			SignatureReferenceNameLower: dbSignature.SignatureReferenceNameLower,
			SignatureSigned:             dbSignature.SignatureSigned,
			SignatureApproved:           dbSignature.SignatureApproved,
			SignatureMajorVersion:       dbSignature.SignatureDocumentMajorVersion,
			SignatureMinorVersion:       dbSignature.SignatureDocumentMinorVersion,
			Version:                     dbSignature.SignatureDocumentMajorVersion + "." + dbSignature.SignatureDocumentMinorVersion,
			SignatureReferenceType:      dbSignature.SignatureReferenceType,
			ProjectID:                   dbSignature.SignatureProjectID,
			Created:                     dbSignature.DateCreated,
			Modified:                    dbSignature.DateModified,
			EmailApprovalList:           dbSignature.EmailWhitelist,
			DomainApprovalList:          dbSignature.DomainWhitelist,
			GithubUsernameApprovalList:  dbSignature.GitHubWhitelist,
			GithubOrgApprovalList:       dbSignature.GitHubOrgWhitelist,
			UserName:                    dbSignature.UserName,
			UserLFID:                    dbSignature.UserLFUsername,
			UserGHID:                    dbSignature.UserGithubUsername,
			SignedOn:                    dbSignature.SignedOn,
			SignatoryName:               dbSignature.SignatoryName,
		}
		sigs = append(sigs, sig)
		go func(sigModel *models.Signature, signatureUserCompanyID string, sigACL []string) {
			defer wg.Done()
			var companyName = ""
			var userName = ""
			var userLFID = ""
			var userGHID = ""
			var userGHUsername = ""
			var swg sync.WaitGroup
			swg.Add(2)

			go func() {
				defer swg.Done()
				if sigModel.SignatureReferenceType == "user" {
					userModel, userErr := repo.usersRepo.GetUser(sigModel.SignatureReferenceID.String())
					if userErr != nil || userModel == nil {
						log.Warnf("unable to lookup user using id: %s, error: %v", sigModel.SignatureReferenceID, userErr)
					} else {
						userName = userModel.Username
						userLFID = userModel.LfUsername
						userGHID = userModel.GithubID
						userGHUsername = userModel.GithubUsername
					}

					if signatureUserCompanyID != "" {
						dbCompanyModel, companyErr := repo.companyRepo.GetCompany(signatureUserCompanyID)
						if companyErr != nil {
							log.Warnf("unable to lookup company using id: %s, error: %v", signatureUserCompanyID, companyErr)
						} else {
							companyName = dbCompanyModel.CompanyName
						}
					}
				} else if sigModel.SignatureReferenceType == "company" {
					dbCompanyModel, companyErr := repo.companyRepo.GetCompany(sigModel.SignatureReferenceID.String())
					if companyErr != nil {
						log.Warnf("unable to lookup company using id: %s, error: %v", sigModel.SignatureReferenceID, companyErr)
					} else {
						companyName = dbCompanyModel.CompanyName
					}
				}
			}()

			var signatureACL []models.User
			go func() {
				defer swg.Done()
				for _, userName := range sigACL {
					if loadACLDetails {
						userModel, userErr := repo.usersRepo.GetUserByUserName(userName, true)
						if userErr != nil {
							log.Warnf("unable to lookup user using username: %s, error: %v", userName, userErr)
						} else {
							if userModel == nil {
								log.Warnf("User looking for username is null: %s for signature: %s", userName, sigModel.SignatureID)
							} else {
								signatureACL = append(signatureACL, *userModel)
							}
						}
					} else {
						signatureACL = append(signatureACL, models.User{LfUsername: userName})
					}
				}
			}()
			swg.Wait()
			sigModel.CompanyName = companyName
			sigModel.UserName = userName
			sigModel.UserLFID = userLFID
			sigModel.UserGHID = userGHID
			sigModel.UserGHUsername = userGHUsername
			sigModel.SignatureACL = signatureACL
		}(sig, dbSignature.SignatureUserCompanyID, dbSignature.SignatureACL)
	}
	wg.Wait()
	return sigs, nil
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

// buildApprovalAttributeList builds the updated approval list based on the added and removed values
func buildApprovalAttributeList(existingList, addEntries, removeEntries []string) *dynamodb.AttributeValue {
	var updatedList []string
	log.Debugf("buildApprovalAttributeList - existing: %+v, add entries: %+v, remove entries: %+v",
		existingList, addEntries, removeEntries)

	// Add the existing entries to our response
	for _, value := range existingList {
		// No duplicates allowed
		if !utils.StringInSlice(value, updatedList) {
			log.Debugf("buildApprovalAttributeList - adding existing entry: %s", value)
			updatedList = append(updatedList, strings.TrimSpace(value))
		} else {
			log.Debugf("buildApprovalAttributeList - skipping existing entry: %s", value)
		}
	}

	// For all the new values...
	for _, value := range addEntries {
		// No duplicates allowed
		if !utils.StringInSlice(value, updatedList) {
			log.Debugf("buildApprovalAttributeList - adding new entry: %s", value)
			updatedList = append(updatedList, strings.TrimSpace(value))
		} else {
			log.Debugf("buildApprovalAttributeList - skipping new entry: %s", value)
		}
	}

	// Remove the items
	log.Debugf("buildApprovalAttributeList - before: %+v - removing entries: %+v", updatedList, removeEntries)
	updatedList = utils.RemoveItemsFromList(updatedList, removeEntries)
	log.Debugf("buildApprovalAttributeList - after: %+v - removing entries: %+v", updatedList, removeEntries)

	// Remove any duplicates - shouldn't have any if checked before adding
	log.Debugf("buildApprovalAttributeList - before: %+v - removing duplicates", updatedList)
	updatedList = utils.RemoveDuplicates(updatedList)
	log.Debugf("buildApprovalAttributeList - after: %+v - removing duplicates", updatedList)

	// Convert to the response type
	var responseList []*dynamodb.AttributeValue
	for _, value := range updatedList {
		responseList = append(responseList, &dynamodb.AttributeValue{S: aws.String(value)})
	}

	return &dynamodb.AttributeValue{L: responseList}
}

// buildCompanyIDList is a helper function to convert the DB response models into a simple list of company IDs
func (repo repository) buildCompanyIDList(results *dynamodb.QueryOutput) ([]SignatureCompanyID, error) {
	var response []SignatureCompanyID

	// The DB signature model
	var dbSignatures []ItemSignature
	err := dynamodbattribute.UnmarshalListOfMaps(results.Items, &dbSignatures)
	if err != nil {
		log.Warnf("error unmarshalling signatures from database, error: %v", err)
		return nil, err
	}

	// Loop and extract the company ID (signature_reference_id) value
	for _, item := range dbSignatures {
		// Lookup the company by ID - try to get more information like the external ID and name
		companyModel, companyLookupErr := repo.companyRepo.GetCompany(item.SignatureReferenceID)
		// Start building a model for this entry in the list
		signatureCompanyID := SignatureCompanyID{
			SignatureID: item.SignatureID,
			CompanyID:   item.SignatureReferenceID,
		}

		if companyLookupErr != nil || companyModel == nil {
			log.Warnf("problem looking up company using id: %s, error: %+v",
				item.SignatureReferenceID, companyLookupErr)
			response = append(response, signatureCompanyID)
		} else {
			if companyModel.CompanyExternalID != "" {
				signatureCompanyID.CompanySFID = companyModel.CompanyExternalID
			}
			if companyModel.CompanyName != "" {
				signatureCompanyID.CompanyName = companyModel.CompanyName
			}
			response = append(response, signatureCompanyID)
		}
	}

	return response, nil
}

func (repo repository) GetClaGroupICLASignatures(claGroupID string, searchTerm *string) (*models.IclaSignatures, error) {
	sortKeyPrefix := fmt.Sprintf("%s#%v#%v", ICLA, true, true)
	// This is the key we want to match
	condition := expression.Key("signature_project_id").Equal(expression.Value(claGroupID)).
		And(expression.Key("sigtype_signed_approved_id").BeginsWith(sortKeyPrefix))

	// Use the builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(buildProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for get cla group icla signatures, claGroupID: %s, error: %v",
			claGroupID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(repo.signatureTableName),
		IndexName:                 aws.String(SignatureProjectIDSigTypeSignedApprovedIDIndex),
		Limit:                     aws.Int64(HugePageSize),
	}
	out := &models.IclaSignatures{List: make([]*models.IclaSignature, 0)}
	if searchTerm != nil {
		searchTerm = aws.String(strings.ToLower(*searchTerm))
	}
	for {
		// Make the DynamoDB Query API call
		results, queryErr := repo.dynamoDBClient.Query(queryInput)
		if queryErr != nil {
			log.Warnf("error retrieving icla signatures for project: %s, error: %v", claGroupID, queryErr)
			return nil, queryErr
		}

		var dbSignatures []ItemSignature

		err := dynamodbattribute.UnmarshalListOfMaps(results.Items, &dbSignatures)
		if err != nil {
			log.Warnf("error unmarshalling icla signatures from database for cla group: %s, error: %v",
				claGroupID, err)
			return nil, err
		}

		for _, sig := range dbSignatures {
			if searchTerm != nil {
				if !strings.Contains(sig.SignatureReferenceNameLower, *searchTerm) {
					continue
				}
			}
			signedOn := sig.DateCreated
			if sig.SignedOn != "" {
				signedOn = sig.SignedOn
			}
			out.List = append(out.List, &models.IclaSignature{
				GithubUsername: sig.UserGithubUsername,
				LfUsername:     sig.UserLFUsername,
				SignatureID:    sig.SignatureID,
				UserEmail:      sig.UserEmail,
				UserName:       sig.UserName,
				SignedOn:       signedOn,
			})
		}

		if len(results.LastEvaluatedKey) == 0 {
			break
		}
		queryInput.ExclusiveStartKey = results.LastEvaluatedKey
		log.Debug("querying next page")
	}
	return out, nil
}

func (repo repository) GetClaGroupCorporateContributors(claGroupID string, companyID *string, searchTerm *string) (*models.CorporateContributorList, error) {
	condition := expression.Key("signature_project_id").Equal(expression.Value(claGroupID))
	if companyID != nil {
		sortKey := fmt.Sprintf("%s#%v#%v#%v", ECLA, true, true, *companyID)
		condition = condition.And(expression.Key("sigtype_signed_approved_id").Equal(expression.Value(sortKey)))
	} else {
		sortKeyPrefix := fmt.Sprintf("%s#%v#%v", ECLA, true, true)
		condition = condition.And(expression.Key("sigtype_signed_approved_id").BeginsWith(sortKeyPrefix))
	}

	// Use the builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(buildProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for get cla group icla signatures, claGroupID: %s, error: %v",
			claGroupID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(repo.signatureTableName),
		IndexName:                 aws.String(SignatureProjectIDSigTypeSignedApprovedIDIndex),
		Limit:                     aws.Int64(HugePageSize),
	}
	out := &models.CorporateContributorList{List: make([]*models.CorporateContributor, 0)}
	if searchTerm != nil {
		searchTerm = aws.String(strings.ToLower(*searchTerm))
	}
	for {
		// Make the DynamoDB Query API call
		results, queryErr := repo.dynamoDBClient.Query(queryInput)
		if queryErr != nil {
			log.Warnf("error retrieving icla signatures for project: %s, error: %v", claGroupID, queryErr)
			return nil, queryErr
		}

		var dbSignatures []ItemSignature

		err := dynamodbattribute.UnmarshalListOfMaps(results.Items, &dbSignatures)
		if err != nil {
			log.Warnf("error unmarshalling icla signatures from database for cla group: %s, error: %v",
				claGroupID, err)
			return nil, err
		}

		for _, sig := range dbSignatures {
			if searchTerm != nil {
				if !strings.Contains(sig.SignatureReferenceNameLower, *searchTerm) {
					continue
				}
			}
			var sigCreatedTime = sig.DateCreated
			t, err := utils.ParseDateTime(sig.DateCreated)
			if err != nil {
				log.Error("fillCorporateContributorModel: unable to parse time", err)
			} else {
				sigCreatedTime = utils.TimeToString(t)
			}
			signatureVersion := fmt.Sprintf("v%s.%s", sig.SignatureDocumentMajorVersion, sig.SignatureDocumentMinorVersion)
			out.List = append(out.List, &models.CorporateContributor{
				GithubID:          sig.UserGithubUsername,
				LinuxFoundationID: sig.UserLFUsername,
				Name:              sig.UserName,
				SignatureVersion:  signatureVersion,
				Email:             sig.UserEmail,
				Timestamp:         sigCreatedTime,
			})
		}

		if len(results.LastEvaluatedKey) == 0 {
			break
		}
		queryInput.ExclusiveStartKey = results.LastEvaluatedKey
		log.Debug("querying next page")
	}
	sort.Slice(out.List, func(i, j int) bool {
		return out.List[i].Name < out.List[j].Name
	})

	return out, nil
}
