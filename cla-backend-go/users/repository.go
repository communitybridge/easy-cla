// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package users

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/google/uuid"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Repository interface defines the functions for the users service
type Repository interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUser(userID string) (*models.User, error)
	GetUserByUserName(userName string, fullMatch bool) (*models.User, error)
	SearchUsers(searchField string, searchTerm string, fullMatch bool) (*models.Users, error)
}

// repository data model
type repository struct {
	stage          string
	dynamoDBClient *dynamodb.DynamoDB
}

// NewRepository creates a new instance of the whitelist service
func NewRepository(awsSession *session.Session, stage string) Repository {
	return repository{
		stage:          stage,
		dynamoDBClient: dynamodb.New(awsSession),
	}
}

// DBUser data model
type DBUser struct {
	UserID             string   `json:"user_id"`
	LFEmail            string   `json:"lf_email"`
	LFUsername         string   `json:"lf_username"`
	DateCreated        string   `json:"date_created"`
	DateModified       string   `json:"date_modified"`
	UserName           string   `json:"user_name"`
	Version            string   `json:"version"`
	UserEmails         []string `json:"user_emails"`
	UserGithubID       string   `json:"user_github_id"`
	UserCompanyID      string   `json:"user_company_id"`
	UserGithubUsername string   `json:"user_github_username"`
}

func (repo repository) CreateUser(user *models.User) (*models.User, error) {
	putStartTime := time.Now()

	tableName := fmt.Sprintf("cla-%s-users", repo.stage)

	theUUID, err := uuid.NewUUID()
	if err != nil {
		return &models.User{}, err
	}
	newUUID := theUUID.String()

	// Set and add the attributes from the request
	attributes := map[string]*dynamodb.AttributeValue{
		"user_id": {
			S: aws.String(newUUID),
		},
	}

	if user.LfEmail != "" {
		attributes["lf_email"] = &dynamodb.AttributeValue{
			S: aws.String(user.LfEmail),
		}
	}
	if user.LfUsername != "" {
		attributes["lf_username"] = &dynamodb.AttributeValue{
			S: aws.String(user.LfUsername),
		}
	}
	if user.Username != "" {
		attributes["user_name"] = &dynamodb.AttributeValue{
			S: aws.String(user.Username),
		}
	}

	// Build the put request
	input := &dynamodb.PutItemInput{
		Item: attributes,
		//ConditionExpression: aws.String("attribute_not_exists"),
		TableName: aws.String(tableName),
	}

	_, err = repo.dynamoDBClient.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.Warnf("dynamodb.ErrCodeConditionalCheckFailedException: %v", aerr.Error())
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				log.Warnf("dynamodb.ErrCodeProvisionedThroughputExceededExceptio: %vn", aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				log.Warnf("dynamodb.ErrCodeResourceNotFoundException: %v", aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				log.Warnf("dynamodb.ErrCodeItemCollectionSizeLimitExceededException: %v", aerr.Error())
			case dynamodb.ErrCodeTransactionConflictException:
				log.Warnf("dynamodb.ErrCodeTransactionConflictException: %v", aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				log.Warnf("dynamodb.ErrCodeRequestLimitExceeded: %v", aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				log.Warnf("dynamodb.ErrCodeInternalServerError: %v", aerr.Error())
			default:
				log.Warnf(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Warnf(err.Error())
		}
		return &models.User{}, err
	}

	log.Debugf("AddUser put took: %v", utils.FmtDuration(time.Since(putStartTime)))
	log.Debugf("Created new user: %+v", user)
	userModel, err := repo.GetUserByUserName(user.Username, true)
	if err != nil {
		log.Warnf("Error locating new user after creation, user: %+v, error: %+v", user, err)
	}
	log.Debugf("Returning new user: %+v", userModel)

	return userModel, err
}

func (repo repository) GetUser(userID string) (*models.User, error) {
	queryStartTime := time.Now()

	tableName := fmt.Sprintf("cla-%s-users", repo.stage)

	// This is the key we want to match
	condition := expression.Key("user_id").Equal(expression.Value(userID))

	// These are the columns we want returned
	projection := buildUserProjection()

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(projection).Build()
	if err != nil {
		log.Warnf("error building expression for user_id : %s, error: %v", userID, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
		//IndexName:                 aws.String(fmt.Sprintf("cla-%s-users", repo.stage)),
	}

	//log.Debugf("Running user query using queryInput: %+v", queryInput)

	// Make the DynamoDB Query API call
	result, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.Warnf("Error retrieving user by user_id: %s, error: %+v", userID, err)
		return nil, err
	}

	//log.Debugf("Result count : %d", *result.Count)
	//log.Debugf("Scanned count: %d", *result.ScannedCount)
	//log.Debugf("Result: %+v", *result)

	// The user model
	var dbUserModels []DBUser

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &dbUserModels)
	if err != nil {
		log.Warnf("error unmarshalling user record from database for user_id: %s, error: %+v", userID, err)
		return nil, err
	}

	log.Debugf("GetUser by ID query took: %v resulting in %d results",
		utils.FmtDuration(time.Since(queryStartTime)), len(dbUserModels))

	if len(dbUserModels) == 0 {
		return nil, nil
	} else if len(dbUserModels) > 1 {
		log.Warnf("retrieved %d results for the getUser(id) query when we should return 0 or 1", len(dbUserModels))
	}

	//log.Debugf("%+v", dbUserModels[0])
	return convertDBUserModel(dbUserModels[0]), nil
}

func (repo repository) GetUserByUserName(userName string, fullMatch bool) (*models.User, error) {
	queryStartTime := time.Now()

	tableName := fmt.Sprintf("cla-%s-users", repo.stage)

	// These are the columns we want returned
	projection := buildUserProjection()

	builder := expression.NewBuilder().WithProjection(projection)
	// This is the filter we want to match
	if fullMatch {
		filter := expression.Name("lf_username").Equal(expression.Value(userName))
		builder.WithFilter(filter)
	} else {
		filter := expression.Name("lf_username").Contains(userName)
		builder.WithFilter(filter)
	}

	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		log.Warnf("error building expression for user name: %s, error: %v", userName, err)
		return nil, err
	}

	// Assemble the scan input parameters
	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	//log.Debugf("Running user query using scanInput: %+v", scanInput)

	var lastEvaluatedKey string
	// The user model
	var dbUserModels []DBUser

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		result, err := repo.dynamoDBClient.Scan(scanInput)
		if err != nil {
			log.Warnf("Error retrieving user by user name: %s, error: %+v", userName, err)
			return nil, err
		}

		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &dbUserModels)
		if err != nil {
			log.Warnf("error unmarshalling user record from database for user name: %s, error: %+v", userName, err)
			return nil, err
		}

		log.Debugf("GetUser by User Name '%s' query took: %v resulting in %d results",
			userName, utils.FmtDuration(time.Since(queryStartTime)), len(dbUserModels))

		if len(dbUserModels) == 1 {
			return convertDBUserModel(dbUserModels[0]), nil
		} else if len(dbUserModels) > 1 {
			log.Warnf("retrieved %d results for the getUser(id) query when we should return 0 or 1", len(dbUserModels))
			return convertDBUserModel(dbUserModels[0]), nil
		}

		// If we have another page of results...
		if result.LastEvaluatedKey["user_id"] != nil {
			lastEvaluatedKey = *result.LastEvaluatedKey["user_id"].S
			scanInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"user_id": {
					S: aws.String(lastEvaluatedKey),
				},
			}
		} else {
			lastEvaluatedKey = ""
		}
	}

	return nil, nil
}

func (repo repository) SearchUsers(searchField string, searchTerm string, fullMatch bool) (*models.Users, error) {
	log.Debugf("Starting Search User")
	// Sorry, no results if empty search field or search term
	if strings.TrimSpace(searchTerm) == "" || strings.TrimSpace(searchField) == "" {
		return &models.Users{
			Users:          []models.User{},
			LastKeyScanned: "",
			ResultCount:    0,
			SearchTerm:     searchTerm,
			TotalCount:     0,
		}, nil
	}

	queryStartTime := time.Now()

	tableName := fmt.Sprintf("cla-%s-users", repo.stage)

	// These are the columns we want returned
	projection := buildUserProjection()

	builder := expression.NewBuilder().WithProjection(projection)
	// This is the filter we want to match
	if fullMatch {
		filter := expression.Name(searchField).Equal(expression.Value(searchTerm))
		builder.WithFilter(filter)
	} else {
		filter := expression.Name(searchField).Contains(searchTerm)
		builder.WithFilter(filter)
	}

	// Use the nice builder to create the expression
	expr, err := builder.Build()
	if err != nil {
		log.Warnf("error building expression for user name: %s, error: %v", searchTerm, err)
		return nil, err
	}

	// Assemble the scan input parameters
	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	var lastEvaluatedKey string
	var users []models.User

	// Loop until we have all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		results, dbErr := repo.dynamoDBClient.Scan(scanInput)
		if dbErr != nil {
			log.Warnf("error retrieving users for search term: %s, error: %v", searchTerm, dbErr)
			return nil, dbErr
		}

		// Convert the list of DB models to a list of response models
		userList, modelErr := buildDBUserModels(results)
		if modelErr != nil {
			log.Warnf("error retrieving users for searchTerm %s in ACL, error: %v", searchTerm, modelErr)
			return nil, modelErr
		}

		// Add to our response model list
		users = append(users, userList...)

		log.Debugf("User search scan took: %v resulting in %d results",
			utils.FmtDuration(time.Since(queryStartTime)), len(results.Items))

		if results.LastEvaluatedKey["user_id"] != nil {
			//log.Debugf("LastEvaluatedKey: %+v", result.LastEvaluatedKey["signature_id"])
			lastEvaluatedKey = *results.LastEvaluatedKey["user_id"].S
			scanInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
				"user_id": {
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
		log.Warnf("error retrieving total user record count for searchTerm: %s, error: %v", searchTerm, err)
		return nil, err
	}

	totalCount := *describeTableResult.Table.ItemCount

	log.Debugf("Total user search took: %v resulting in %d results",
		utils.FmtDuration(time.Since(queryStartTime)), len(users))

	return &models.Users{
		ResultCount:    int64(len(users)),
		TotalCount:     totalCount,
		LastKeyScanned: lastEvaluatedKey,
		Users:          users,
	}, nil

}

// convertDBUserModel translates a dyanamoDB data model into a service response model
func convertDBUserModel(user DBUser) *models.User {
	return &models.User{
		UserID:         user.UserID,
		LfEmail:        user.LFEmail,
		LfUsername:     user.LFUsername,
		DateCreated:    user.DateCreated,
		DateModified:   user.DateModified,
		Username:       user.UserName,
		Version:        user.Version,
		Emails:         user.UserEmails,
		GithubID:       user.UserGithubID,
		CompanyID:      user.UserCompanyID,
		GithubUsername: user.UserGithubUsername,
	}
}

func buildUserProjection() expression.ProjectionBuilder {
	// These are the columns we want returned
	return expression.NamesList(
		expression.Name("user_id"),
		expression.Name("date_created"),
		expression.Name("date_modified"),
		expression.Name("lf_email"),
		expression.Name("lf_username"),
		expression.Name("user_emails"),
		expression.Name("user_name"),
		expression.Name("user_company_id"),
		expression.Name("user_github_username"),
		expression.Name("user_github_id"),
	)
}

// buildDBUserModels converts the response model into a response data model
func buildDBUserModels(results *dynamodb.ScanOutput) ([]models.User, error) {
	var users []models.User

	type ItemSignature struct {
		UserID       string `json:"user_id"`
		DateCreated  string `json:"date_created"`
		DateModified string `json:"date_modified"`
	}

	// The DB company model
	var dbUsers []ItemSignature

	err := dynamodbattribute.UnmarshalListOfMaps(results.Items, &dbUsers)
	if err != nil {
		log.Warnf("error unmarshalling users from database, error: %v", err)
		return nil, err
	}

	for _, dbUser := range dbUsers {
		users = append(users, models.User{
			UserID:       dbUser.UserID,
			DateCreated:  dbUser.DateCreated,
			DateModified: dbUser.DateModified,
		})
	}

	return users, nil
}
