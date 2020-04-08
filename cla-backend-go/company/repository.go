// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package company

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/communitybridge/easycla/cla-backend-go/user"

	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/go-openapi/strfmt"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/communitybridge/easycla/cla-backend-go/gen/models"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gofrs/uuid"
)

// errors
var (
	ErrCompanyDoesNotExist = errors.New("company does not exist")
)

// CompanyRepository interface methods
type CompanyRepository interface { //nolint
	GetMetrics() (*models.CompaniesMetrics, error)
	GetCompanies() (*models.Companies, error)
	GetCompany(companyID string) (*models.Company, error)
	SearchCompanyByName(companyName string, nextKey string) (*models.Companies, error)
	GetCompaniesByUserManager(userID string, userModel user.User) (*models.Companies, error)
	GetCompaniesByUserManagerWithInvites(userID string, userModel user.User) (*models.CompaniesWithInvites, error)

	AddPendingCompanyInviteRequest(companyID string, userID string) error
	GetCompanyInviteRequests(companyID string) ([]Invite, error)
	GetCompanyUserInviteRequests(companyID string, userID string) (*Invite, error)
	GetUserInviteRequests(userID string) ([]Invite, error)
	RejectCompanyInviteRequest(companyID string, userID string) error
	DeletePendingCompanyInviteRequest(InviteID string) error

	UpdateCompanyAccessList(companyID string, companyACL []string) error
}

type repository struct {
	stage          string
	dynamoDBClient *dynamodb.DynamoDB
}

// Company data model
type Company struct {
	CompanyID   string   `dynamodbav:"company_id"`
	CompanyName string   `dynamodbav:"company_name"`
	CompanyACL  []string `dynamodbav:"company_acl"`
	Created     string   `dynamodbav:"date_created"`
	Updated     string   `dynamodbav:"date_modified"`
}

// Invite data model
type Invite struct {
	CompanyInviteID    string `dynamodbav:"company_invite_id"`
	RequestedCompanyID string `dynamodbav:"requested_company_id"`
	UserID             string `dynamodbav:"user_id"`
	Status             string `dynamodbav:"status"`
}

// NewRepository creates a new company repository instance
func NewRepository(awsSession *session.Session, stage string) CompanyRepository {
	return repository{
		stage:          stage,
		dynamoDBClient: dynamodb.New(awsSession),
	}
}

// GetCompanies retrieves all the companies
func (repo repository) GetCompanies() (*models.Companies, error) {
	tableName := fmt.Sprintf("cla-%s-companies", repo.stage)

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithProjection(buildCompanyProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for get all companies scan error: %v", err)
		return nil, err
	}

	// Assemble the query input parameters
	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	var lastEvaluatedKey string
	var companies []models.Company

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		results, dbErr := repo.dynamoDBClient.Scan(scanInput)
		if dbErr != nil {
			log.Warnf("error retrieving get all companies, error: %v", dbErr)
			return nil, dbErr
		}

		// Convert the list of DB models to a list of response models
		companyList, modelErr := buildCompanyModels(results)
		if modelErr != nil {
			log.Warnf("error retrieving get all companies, error: %v", modelErr)
			return nil, modelErr
		}

		// Add to our response model list
		companies = append(companies, companyList...)

		if results.LastEvaluatedKey["company_id"] != nil {
			//log.Debugf("LastEvaluatedKey: %+v", result.LastEvaluatedKey["signature_id"])
			lastEvaluatedKey = *results.LastEvaluatedKey["company_id"].S
			scanInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"company_id": {
					S: aws.String(lastEvaluatedKey),
				},
			}
		} else {
			lastEvaluatedKey = ""
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}

	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total company record count, error: %v", err)
		return nil, err
	}

	totalCount := *describeTableResult.Table.ItemCount

	return &models.Companies{
		ResultCount:    int64(len(companies)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Companies:      companies,
	}, nil
}

// GetCompany returns a company based on the company ID
func (repo repository) GetCompany(companyID string) (*models.Company, error) {

	tableName := fmt.Sprintf("cla-%s-companies", repo.stage)
	queryStartTime := time.Now()

	companyTableData, err := repo.dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"company_id": {
				S: aws.String(companyID),
			},
		},
	})

	if err != nil {
		log.Warnf(err.Error())
		log.Warnf("error fetching company table data using company id: %s, error: %v", companyID, err)
		return nil, err
	}

	if len(companyTableData.Item) == 0 {
		return nil, ErrCompanyDoesNotExist
	}
	log.Debugf("Get company query took: %v", utils.FmtDuration(time.Since(queryStartTime)))

	dbCompanyModel := Company{}
	err = dynamodbattribute.UnmarshalMap(companyTableData.Item, &dbCompanyModel)
	if err != nil {
		log.Warnf("error unmarshalling company table data, error: %v", err)
		return nil, err
	}
	const timeFormat = "2006-01-02T15:04:05.999999+0000"
	// Convert the "string" date time
	createdDateTime, err := time.Parse(timeFormat, dbCompanyModel.Created)
	if err != nil {
		log.Warnf("Error converting created date time for company: %s, error: %v", companyID, err)
		return nil, err
	}
	updateDateTime, err := time.Parse(timeFormat, dbCompanyModel.Updated)
	if err != nil {
		log.Warnf("Error converting updated date time for company: %s, error: %v", companyID, err)
		return nil, err
	}

	// Convert the local DB model to a public swagger model
	return &models.Company{
		CompanyACL:  dbCompanyModel.CompanyACL,
		CompanyID:   dbCompanyModel.CompanyID,
		CompanyName: dbCompanyModel.CompanyName,
		Created:     strfmt.DateTime(createdDateTime),
		Updated:     strfmt.DateTime(updateDateTime),
	}, nil

}

// SearchCompanyByName locates companies by the matching name and return any potential matches
func (repo repository) SearchCompanyByName(companyName string, nextKey string) (*models.Companies, error) {
	// Sorry, no results if empty company name
	if strings.TrimSpace(companyName) == "" {
		return &models.Companies{
			Companies:      []models.Company{},
			LastKeyScanned: "",
			ResultCount:    0,
			SearchTerms:    companyName,
			TotalCount:     0,
		}, nil
	}

	queryStartTime := time.Now()

	tableName := fmt.Sprintf("cla-%s-companies", repo.stage)

	// This is the company name we want to match
	filter := expression.Name("company_name").Contains(companyName)

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithFilter(filter).WithProjection(buildCompanyProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for company scan, companyName: %s, error: %v",
			companyName, err)
		return nil, err
	}

	// Assemble the query input parameters
	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	// If we have the next key, set the exclusive start key value
	if nextKey != "" {
		log.Debugf("Received a nextKey, value: %s", nextKey)
		// The primary key of the first item that this operation will evaluate.
		// and the query key (if not the same)
		scanInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"company_id": {
				S: aws.String(nextKey),
			},
		}
	}

	//log.Debugf("Running company search scan using queryInput: %+v", scanInput)

	var lastEvaluatedKey string
	var companies []models.Company

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		results, dbErr := repo.dynamoDBClient.Scan(scanInput)
		if dbErr != nil {
			log.Warnf("error retrieving companies for search term: %s, error: %v", companyName, dbErr)
			return nil, dbErr
		}

		// Convert the list of DB models to a list of response models
		companyList, modelErr := buildCompanyModels(results)
		if modelErr != nil {
			log.Warnf("error retrieving companies for companyName %s in ACL, error: %v", companyName, modelErr)
			return nil, modelErr
		}

		// Add to our response model list
		companies = append(companies, companyList...)

		log.Debugf("Company search scan took: %v resulting in %d results",
			utils.FmtDuration(time.Since(queryStartTime)), len(results.Items))

		if results.LastEvaluatedKey["company_id"] != nil {
			//log.Debugf("LastEvaluatedKey: %+v", result.LastEvaluatedKey["signature_id"])
			lastEvaluatedKey = *results.LastEvaluatedKey["company_id"].S
			scanInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"company_id": {
					S: aws.String(lastEvaluatedKey),
				},
			}
		} else {
			lastEvaluatedKey = ""
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}

	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total company record count for companyName: %s, error: %v", companyName, err)
		return nil, err
	}

	totalCount := *describeTableResult.Table.ItemCount

	log.Debugf("Total company search took: %v resulting in %d results",
		utils.FmtDuration(time.Since(queryStartTime)), len(companies))

	return &models.Companies{
		ResultCount:    int64(len(companies)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Companies:      companies,
	}, nil
}

// GetCompanyUserManager the get a list of companies when provided the company id and user manager
func (repo repository) GetCompaniesByUserManager(userID string, userModel user.User) (*models.Companies, error) {
	// Sorry, no results if empty user ID
	if strings.TrimSpace(userID) == "" {
		return &models.Companies{
			Companies:      []models.Company{},
			LastKeyScanned: "",
			ResultCount:    0,
			TotalCount:     0,
		}, nil
	}

	queryStartTime := time.Now()

	tableName := fmt.Sprintf("cla-%s-companies", repo.stage)

	// This is the user name we want to match
	var filter expression.ConditionBuilder
	if userModel.LFUsername != "" {
		filter = expression.Name("company_acl").Contains(userModel.LFUsername)
	} else if userModel.UserName != "" {
		filter = expression.Name("company_acl").Contains(userModel.UserName)
	} else {
		log.Warnf("unable to query user with no LF username or username in their data model - user iD: %s.", userID)
		return &models.Companies{
			Companies:      []models.Company{},
			LastKeyScanned: "",
			ResultCount:    0,
			TotalCount:     0,
		}, nil
	}

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithFilter(filter).WithProjection(buildCompanyProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for company scan, userID %s in ACL, error: %v", userID, err)
		return nil, err
	}

	// Assemble the query input parameters
	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	//log.Debugf("Running company search scan using queryInput: %+v", scanInput)
	var lastEvaluatedKey string
	var companies []models.Company

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		results, dbErr := repo.dynamoDBClient.Scan(scanInput)
		if dbErr != nil {
			log.Warnf("error retrieving companies for userID %s in ACL, error: %v", userID, dbErr)
			return nil, dbErr
		}

		// Convert the list of DB models to a list of response models
		companyList, modelErr := buildCompanyModels(results)
		if modelErr != nil {
			log.Warnf("error retrieving companies for userID %s in ACL, error: %v", userID, modelErr)
			return nil, modelErr
		}

		// Add to our response model list
		companies = append(companies, companyList...)

		log.Debugf("Company search with user in ACL scan took: %v resulting in %d results",
			utils.FmtDuration(time.Since(queryStartTime)), len(results.Items))

		if results.LastEvaluatedKey["company_invite_id"] != nil {
			//log.Debugf("LastEvaluatedKey: %+v", result.LastEvaluatedKey["signature_id"])
			lastEvaluatedKey = *results.LastEvaluatedKey["company_invite_id"].S
			scanInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"company_invite_id": {
					S: aws.String(lastEvaluatedKey),
				},
			}
		} else {
			lastEvaluatedKey = ""
		}
	}

	// How many total records do we have - may not be up-to-date as this value is updated only periodically
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}

	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total company record count, error: %v", err)
		return nil, err
	}

	totalCount := *describeTableResult.Table.ItemCount

	log.Debugf("Total company search took: %v resulting in %d results",
		utils.FmtDuration(time.Since(queryStartTime)), len(companies))

	return &models.Companies{
		ResultCount:    int64(len(companies)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Companies:      companies,
	}, nil
}

// GetCompanyUserManagerWithInvites the get a list of companies including status when provided the company id and user manager
func (repo repository) GetCompaniesByUserManagerWithInvites(userID string, userModel user.User) (*models.CompaniesWithInvites, error) {
	companies, err := repo.GetCompaniesByUserManager(userID, userModel)
	if err != nil {
		log.Warnf("error retrieving companies for userID %s in ACL, error: %v", userID, err)
		return nil, err
	}

	// Query the invites table for list of invitations for this user
	invites, err := repo.GetUserInviteRequests(userID)
	if err != nil {
		log.Warnf("error retrieving companies invites for userID %s, error: %v", userID, err)
		return nil, err
	}

	return repo.buildCompaniesByUserManagerWithInvites(companies, invites), nil
}

func (repo repository) buildCompaniesByUserManagerWithInvites(companies *models.Companies, invites []Invite) *models.CompaniesWithInvites {
	companiesWithInvites := models.CompaniesWithInvites{
		ResultCount: int64(len(companies.Companies) + len(invites)),
		TotalCount:  companies.TotalCount + int64(len(invites)),
	}

	var companyWithInvite []models.CompanyWithInvite
	for _, company := range companies.Companies {
		companyWithInvite = append(companyWithInvite, models.CompanyWithInvite{
			CompanyName: company.CompanyName,
			CompanyID:   company.CompanyID,
			CompanyACL:  company.CompanyACL,
			Created:     company.Created,
			Updated:     company.Updated,
			Status:      "Joined",
		})
	}

	for _, invite := range invites {
		company, err := repo.GetCompany(invite.RequestedCompanyID)
		if err != nil {
			log.Warnf("error retrieving company with company ID %s, error: %v - skipping invite", company, err)
			continue
		}

		// Default status is pending if there's a record but no status
		if invite.Status == "" {
			invite.Status = StatusPending
		}

		companyWithInvite = append(companyWithInvite, models.CompanyWithInvite{
			CompanyName: company.CompanyName,
			CompanyID:   company.CompanyID,
			CompanyACL:  company.CompanyACL,
			Created:     company.Created,
			Updated:     company.Updated,
			Status:      invite.Status,
		})
	}

	companiesWithInvites.CompaniesWithInvites = companyWithInvite

	return &companiesWithInvites
}

// buildCompanyModels converts the response model into a response data model
func buildCompanyModels(results *dynamodb.ScanOutput) ([]models.Company, error) {
	var companies []models.Company

	type ItemSignature struct {
		CompanyID   string   `json:"company_id"`
		CompanyName string   `json:"company_name"`
		CompanyACL  []string `json:"company_acl"`
		Created     string   `json:"date_created"`
		Modified    string   `json:"date_modified"`
	}

	// The DB company model
	var dbCompanies []ItemSignature

	err := dynamodbattribute.UnmarshalListOfMaps(results.Items, &dbCompanies)
	if err != nil {
		log.Warnf("error unmarshalling companies from database, error: %v", err)
		return nil, err
	}

	for _, dbCompany := range dbCompanies {
		createdDateTime, err := utils.ParseDateTime(dbCompany.Created)
		if err != nil {
			log.Warnf("Unable to parse company created date time: %s, error: %v - using current time",
				dbCompany.Created, err)
			createdDateTime = time.Now()
		}

		modifiedDateTime, err := utils.ParseDateTime(dbCompany.Modified)
		if err != nil {
			log.Warnf("Unable to parse company modified date time: %s, error: %v - using current time",
				dbCompany.Created, err)
			modifiedDateTime = time.Now()
		}

		companies = append(companies, models.Company{
			CompanyACL:  dbCompany.CompanyACL,
			CompanyID:   dbCompany.CompanyID,
			CompanyName: dbCompany.CompanyName,
			Created:     strfmt.DateTime(createdDateTime),
			Updated:     strfmt.DateTime(modifiedDateTime),
		})
	}

	return companies, nil
}

// GetCompanyInviteRequests returns a list of company invites when provided the company ID
func (repo repository) GetCompanyInviteRequests(companyID string) ([]Invite, error) {

	queryStartTime := time.Now()

	tableName := fmt.Sprintf("cla-%s-company-invites", repo.stage)

	input := &dynamodb.QueryInput{
		KeyConditions: map[string]*dynamodb.Condition{
			"requested_company_id": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(companyID),
					},
				},
			},
		},
		TableName: aws.String(tableName),
		IndexName: aws.String("requested-company-index"),
	}
	companyInviteAV, err := repo.dynamoDBClient.Query(input)
	if err != nil {
		log.Warnf("Unable to retrieve data from Company-Invites table, error: %v", err)
		return nil, err
	}

	log.Debugf("Company Invites query took: %v",
		utils.FmtDuration(time.Since(queryStartTime)))

	var companyInvites []Invite
	err = dynamodbattribute.UnmarshalListOfMaps(companyInviteAV.Items, &companyInvites)
	if err != nil {
		log.Warnf("error unmarshalling company invite data, error: %v", err)
		return nil, err
	}

	return companyInvites, nil
}

// GetCompanyUserInviteRequests returns a list of company invites when provided the company ID and user ID
func (repo repository) GetCompanyUserInviteRequests(companyID string, userID string) (*Invite, error) {
	queryStartTime := time.Now()

	tableName := fmt.Sprintf("cla-%s-company-invites", repo.stage)

	// These are the keys we want to match
	condition := expression.Key("requested_company_id").Equal(expression.Value(companyID))
	filter := expression.Name("user_id").Equal(expression.Value(userID))

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().
		WithKeyCondition(condition).
		WithFilter(filter).
		WithProjection(buildInvitesProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for company scan, companyID: %s with userID: %s, error: %v",
			companyID, userID, err)
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
		IndexName:                 aws.String("requested-company-index"), // Name of a secondary index
	}

	queryResults, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.Warnf("Unable to retrieve data from Company-Invites table using company id: %s and user id: %s, error: %v", companyID, userID, err)
		return nil, err
	}

	log.Debugf("Company Invites query took: %v with %d results",
		utils.FmtDuration(time.Since(queryStartTime)), len(queryResults.Items))

	var companyInvites []Invite
	err = dynamodbattribute.UnmarshalListOfMaps(queryResults.Items, &companyInvites)
	if err != nil {
		log.Warnf("error unmarshalling company invite data using company id: %s and user id: %s, error: %v",
			companyID, userID, err)
		return nil, err
	}

	if len(companyInvites) == 0 {
		log.Debugf("Unable to find company invite for company id: %s and user id: %s", companyID, userID)
		return nil, nil
	}

	if len(companyInvites) > 1 {
		log.Warnf("Company invite should have one result, found: %d for company id: %s and user id: %s",
			len(companyInvites), companyID, userID)
	}

	return &companyInvites[0], nil
}

// GetUserInviteRequests returns a list of company invites when provided the user ID
func (repo repository) GetUserInviteRequests(userID string) ([]Invite, error) {

	queryStartTime := time.Now()

	tableName := fmt.Sprintf("cla-%s-company-invites", repo.stage)
	filter := expression.Name("user_id").Equal(expression.Value(userID))

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().
		WithFilter(filter).
		WithProjection(buildInvitesProjection()).Build()
	if err != nil {
		log.Warnf("error building expression for company scan with userID: %s, error: %v", userID, err)
		return nil, err
	}

	// Assemble the query input parameters
	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	var lastEvaluatedKey string
	var companyInvites []Invite

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {

		queryResults, err := repo.dynamoDBClient.Scan(scanInput)
		if err != nil {
			log.Warnf("Unable to retrieve data from Company-Invites table using user id: %s, error: %v", userID, err)
			return nil, err
		}

		log.Debugf("Company Invites query with user ID %s took: %v with %d results", userID,
			utils.FmtDuration(time.Since(queryStartTime)), len(queryResults.Items))

		var companyInvitesList []Invite
		err = dynamodbattribute.UnmarshalListOfMaps(queryResults.Items, &companyInvitesList)
		if err != nil {
			log.Warnf("error unmarshalling company invite data using user id: %s, error: %v", userID, err)
			return nil, err
		}

		// Add to our response model
		companyInvites = append(companyInvites, companyInvitesList...)

		// Determine if we have more records - if so, update the start key and loop again
		if queryResults.LastEvaluatedKey["company_invite_id"] != nil {
			//log.Debugf("LastEvaluatedKey: %+v", result.LastEvaluatedKey["signature_id"])
			lastEvaluatedKey = *queryResults.LastEvaluatedKey["company_invite_id"].S
			scanInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"company_invite_id": {
					S: aws.String(lastEvaluatedKey),
				},
			}
		} else {
			lastEvaluatedKey = ""
		}
	}

	return companyInvites, nil
}

// AddPendingCompanyInviteRequest adds a pending company invite when provided the company ID and user ID
func (repo repository) AddPendingCompanyInviteRequest(companyID string, userID string) error {

	// First, let's check if we already have a previous invite for this company and user ID pair
	previousInvite, err := repo.GetCompanyUserInviteRequests(companyID, userID)
	if err != nil {
		log.Warnf("Previous invite already exists for company id: %s and user: %s, error: %v",
			companyID, userID, err)
		return err
	}

	// We we already have an invite...don't create another one
	if previousInvite != nil {
		log.Warnf("Invite already exists for company id: %s and user: %s - skipping creation", companyID, userID)
		return nil
	}

	companyInviteID, err := uuid.NewV4()
	if err != nil {
		log.Warnf("Unable to generate a UUID for a pending invite, error: %v", err)
		return err
	}

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"company_invite_id": {
				S: aws.String(companyInviteID.String()),
			},
			"requested_company_id": {
				S: aws.String(companyID),
			},
			"user_id": {
				S: aws.String(userID),
			},
		},
		TableName: aws.String(fmt.Sprintf("cla-%s-company-invites", repo.stage)),
	}

	_, err = repo.dynamoDBClient.PutItem(input)
	if err != nil {
		log.Warnf("Unable to create a new pending invite, error: %v", err)
		return err
	}

	return nil
}

// RejectCompanyInviteRequest rejects a pending company invite when provided the company ID and user ID
func (repo repository) RejectCompanyInviteRequest(companyID string, userID string) error {
	log.Warnf("RejectCompanyInviteRequest not implemented")
	return nil
}

// DeletePendingCompanyInviteRequest deletes the spending invite
func (repo repository) DeletePendingCompanyInviteRequest(inviteID string) error {
	tableName := fmt.Sprintf("cla-%s-company-invites", repo.stage)
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"company_invite_id": {
				S: aws.String(inviteID),
			},
		},
		TableName: aws.String(tableName),
	}

	_, err := repo.dynamoDBClient.DeleteItem(input)
	if err != nil {
		log.Warnf("Unable to delete Company Invite Request, error: %v", err)
		return err
	}

	return nil
}

// UpdateCompanyAccessList updates the company ACL when provided the company ID and ACL list
func (repo repository) UpdateCompanyAccessList(companyID string, companyACL []string) error {
	tableName := fmt.Sprintf("cla-%s-companies", repo.stage)
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#S": aws.String("company_acl"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":s": {
				SS: aws.StringSlice(companyACL),
			},
		},
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"company_id": {
				S: aws.String(companyID),
			},
		},
		UpdateExpression: aws.String("SET #S = :s"),
	}

	_, err := repo.dynamoDBClient.UpdateItem(input)
	if err != nil {
		log.Warnf("Error updating Company Access List, error: %v", err)
		return err
	}

	return nil
}
func (repo repository) GetMetrics() (*models.CompaniesMetrics, error) {
	var out models.CompaniesMetrics
	tableName := fmt.Sprintf("cla-%s-companies", repo.stage)
	// Do these counts in parallel
	var wg sync.WaitGroup
	wg.Add(2)

	var totalCount int64
	var companies []models.CompanySimpleModel

	go func(tableName string) {
		defer wg.Done()
		// How many total records do we have - may not be up-to-date as this value is updated only periodically
		describeTableInput := &dynamodb.DescribeTableInput{
			TableName: &tableName,
		}
		describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
		if err != nil {
			log.Warnf("error retrieving total record count, error: %v", err)
		}
		// Meta-data for the response
		totalCount = *describeTableResult.Table.ItemCount
	}(tableName)

	go func() {
		defer wg.Done()

		// Use the last evaluated key to determine if we have more to process
		lastEvaluatedKey := ""

		for ok := true; ok; ok = lastEvaluatedKey != "" {
			companyModels, err := repo.GetCompanies()
			if err != nil {
				log.Warnf("error retrieving companies for metrics, error: %v", err)
			}
			// Convert the full response model to a simple model for metrics
			companies = append(companies, buildSimpleModel(companyModels)...)

			// Save the last evaluated key - use it to determine if we have more to process
			lastEvaluatedKey = companyModels.LastKeyScanned
		}
	}()

	// Wait for the counts to finish
	wg.Wait()

	out.TotalCount = totalCount
	out.Companies = companies
	return &out, nil
}

// buildSimpleModel converts the DB model to a simple response model
func buildSimpleModel(dbCompaniesModel *models.Companies) []models.CompanySimpleModel {
	if dbCompaniesModel == nil || dbCompaniesModel.Companies == nil {
		return []models.CompanySimpleModel{}
	}

	var simpleModels []models.CompanySimpleModel
	for _, dbModel := range dbCompaniesModel.Companies {
		simpleModels = append(simpleModels, models.CompanySimpleModel{
			CompanyName:         dbModel.CompanyName,
			CompanyManagerCount: int64(len(dbModel.CompanyACL)),
		})
	}

	return simpleModels
}
