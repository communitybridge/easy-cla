// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package users

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/go-openapi/errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// UserRepository interface defines the functions for the users repository
type UserRepository interface {
	CreateUser(user *models.User) (*models.User, error)
	Save(user *models.UserUpdate) (*models.User, error)
	Delete(userID string) error
	GetUser(userID string) (*models.User, error)
	GetUserByLFUserName(lfUserName string) (*models.User, error)
	GetUserByUserName(userName string, fullMatch bool) (*models.User, error)
	GetUserByEmail(userEmail string) (*models.User, error)
	GetUserByGitHubUsername(gitHubUsername string) (*models.User, error)
	SearchUsers(searchField string, searchTerm string, fullMatch bool) (*models.Users, error)
}

// repository data model
type repository struct {
	stage          string
	dynamoDBClient *dynamodb.DynamoDB
	tableName      string
}

// NewRepository creates a new instance of the whitelist service
func NewRepository(awsSession *session.Session, stage string) UserRepository {
	return repository{
		stage:          stage,
		dynamoDBClient: dynamodb.New(awsSession),
		tableName:      fmt.Sprintf("cla-%s-users", stage),
	}
}

// CreateUser creates a new user
func (repo repository) CreateUser(user *models.User) (*models.User, error) {

	theUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	newUUID := theUUID.String()

	// Set and add the attributes from the request
	user.UserID = newUUID
	attributes := map[string]*dynamodb.AttributeValue{
		"user_id": {
			S: aws.String(newUUID),
		},
	}

	attributes["admin"] = &dynamodb.AttributeValue{
		BOOL: aws.Bool(user.Admin),
	}

	if user.UserExternalID != "" {
		attributes["user_external_id"] = &dynamodb.AttributeValue{
			S: aws.String(user.UserExternalID),
		}
	}

	if user.GithubID != "" {
		attributes["user_github_id"] = &dynamodb.AttributeValue{
			S: aws.String(user.GithubID),
		}
	}

	if user.GithubUsername != "" {
		attributes["user_github_username"] = &dynamodb.AttributeValue{
			S: aws.String(user.GithubUsername),
		}
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
	now := time.Now().UTC().Format(time.RFC3339)

	user.DateCreated = now
	attributes["date_created"] = &dynamodb.AttributeValue{
		S: aws.String(now),
	}

	user.DateModified = now
	attributes["date_modified"] = &dynamodb.AttributeValue{
		S: aws.String(now),
	}

	user.Version = "v1"
	attributes["version"] = &dynamodb.AttributeValue{
		S: aws.String("v1"),
	}

	// Build the put request
	input := &dynamodb.PutItemInput{
		Item: attributes,
		//ConditionExpression: aws.String("attribute_not_exists"),
		TableName: aws.String(repo.tableName),
	}

	_, err = repo.dynamoDBClient.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.Warnf("dynamodb.ErrCodeConditionalCheckFailedException: %v", aerr.Error())
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				log.Warnf("dynamodb.ErrCodeProvisionedThroughputExceededException: %vn", aerr.Error())
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
		return nil, err
	}

	log.Debugf("Created new user: %+v", user)
	return user, err
}

func (repo repository) getUserByUpdateModel(user *models.UserUpdate) (*models.User, error) {
	// Log fields
	f := logrus.Fields{
		"functionName":     "GetUserByUpdateModel",
		"lf_username":      user.LfUsername,
		"lf_email":         user.LfEmail,
		"github_username":  user.GithubUsername,
		"github_id":        user.GithubID,
		"company_id":       user.CompanyID,
		"user_id":          user.UserID,
		"user_external_id": user.UserExternalID,
	}

	var err error
	var existingUserModel *models.User
	if user.LfUsername != "" {
		log.WithFields(f).Debugf("looking up user by username: %s", user.LfUsername)
		existingUserModel, err = repo.GetUserByUserName(user.LfUsername, true)
		if err != nil {
			log.WithFields(f).Warnf("error fetching existing user record: %+v, error: %v", user, err)
			return nil, err
		}
	}

	// Didn't find it lookup up via LF Username/ID, so, let's try another way
	// Try to lookup via GH username, if provided...
	if existingUserModel == nil && user.GithubUsername != "" {
		log.WithFields(f).Debugf("looking up user by github username: %s", user.GithubUsername)
		existingUserModel, err = repo.GetUserByGitHubUsername(user.GithubUsername)
		if err != nil {
			log.WithFields(f).Warnf("error fetching existing user record by GitHub username: %s, error: %v",
				user.GithubUsername, err)
			return nil, err
		}
		log.WithFields(f).Debugf("Found user by GitHub Username: %+v", existingUserModel)
	}

	// Still couldn't find it - time to give up
	if existingUserModel == nil {
		log.WithFields(f).Warnf("error fetching existing user record: %+v, error: %v", user, err)
		return nil, nil
	}

	return existingUserModel, nil
}

// Save saves the user model to the data store
func (repo repository) Save(user *models.UserUpdate) (*models.User, error) {
	// Log fields
	f := logrus.Fields{
		"functionName":     "Save",
		"lf_username":      user.LfUsername,
		"lf_email":         user.LfEmail,
		"github_username":  user.GithubUsername,
		"github_id":        user.GithubID,
		"company_id":       user.CompanyID,
		"user_id":          user.UserID,
		"user_external_id": user.UserExternalID,
		"tableName":        repo.tableName,
	}

	var oldUserModel *models.User
	var err error
	oldUserModel, err = repo.getUserByUpdateModel(user)
	if err != nil {
		log.WithFields(f).Warnf("error fetching existing user record, error: %v", err)
		return nil, err
	}

	// Still couldn't find it - time to give up
	if oldUserModel == nil {
		log.WithFields(f).Warnf("error fetching existing user record: %+v, error: %v", user, err)
		return nil, nil
	}

	// return values flag - Returns all of the attributes of the item, as they appear after the UpdateItem operation.
	addReturnValues := "ALL_NEW" // nolint

	updatedDateTime := time.Now().UTC()
	expressionAttributeNames := map[string]*string{}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{}
	updateExpression := "SET "

	if user.LfEmail != "" && oldUserModel.LfEmail != user.LfEmail {
		log.WithFields(f).Debugf("building query - adding lf_email: %s", user.LfEmail)
		expressionAttributeNames["#E"] = aws.String("lf_email")
		expressionAttributeValues[":e"] = &dynamodb.AttributeValue{S: aws.String(user.LfEmail)}
		updateExpression = updateExpression + " #E = :e, "
	}

	if user.LfUsername != "" && oldUserModel.LfUsername != user.LfUsername {
		log.WithFields(f).Debugf("building query - adding lf_username: %s", user.LfUsername)
		expressionAttributeNames["#U"] = aws.String("lf_username")
		expressionAttributeValues[":u"] = &dynamodb.AttributeValue{S: aws.String(user.LfUsername)}
		updateExpression = updateExpression + " #U = :u, "
	}

	if user.CompanyID != "" && oldUserModel.CompanyID != user.CompanyID {
		log.WithFields(f).Debugf("building query - adding user_company_id: %s", user.CompanyID)
		expressionAttributeNames["#C"] = aws.String("user_company_id")
		expressionAttributeValues[":c"] = &dynamodb.AttributeValue{S: aws.String(user.CompanyID)}
		updateExpression = updateExpression + " #C = :c, "
	}

	if user.GithubUsername != "" && oldUserModel.GithubUsername != user.GithubUsername {
		log.WithFields(f).Debugf("building query - adding user_github_username: %s", user.GithubUsername)
		expressionAttributeNames["#GU"] = aws.String("user_github_username")
		expressionAttributeValues[":gu"] = &dynamodb.AttributeValue{S: aws.String(user.GithubUsername)}
		updateExpression = updateExpression + " #GU = :gu, "
	}

	if user.GithubID != "" && oldUserModel.GithubID != user.GithubID {
		log.WithFields(f).Debugf("building query - adding user_github_id: %s", user.GithubID)
		expressionAttributeNames["#GI"] = aws.String("user_github_id")
		expressionAttributeValues[":gi"] = &dynamodb.AttributeValue{N: aws.String(user.GithubID)}
		updateExpression = updateExpression + " #GI = :gi, "
	}

	log.Debugf("building query - updating date_modified: %s", updatedDateTime.Format(time.RFC3339))
	expressionAttributeNames["#D"] = aws.String("date_modified")
	expressionAttributeValues[":d"] = &dynamodb.AttributeValue{S: aws.String(updatedDateTime.Format(time.RFC3339))}
	updateExpression = updateExpression + " #D = :d "

	// Update dynamoDB table
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(oldUserModel.UserID),
			},
		},
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		UpdateExpression:          &updateExpression,
		ReturnValues:              &addReturnValues,
	}

	_, err = repo.dynamoDBClient.UpdateItem(input)
	if err != nil {
		log.WithFields(f).Warnf("Error updating user record: %+v, error: %v", user, err)
		return nil, err
	}

	log.WithFields(f).Debugf("Save User - looking up saved user by username: %s", user.LfUsername)
	newUserModel, err := repo.getUserByUpdateModel(user)
	if err != nil || newUserModel == nil {
		log.WithFields(f).Warnf("Error fetching updated user record: %+v, error: %v", user, err)
		return nil, err
	}

	log.WithFields(f).Debugf("Returning updated user: %+v", newUserModel)
	return newUserModel, err
}

// Delete deletes the specified user
func (repo repository) Delete(userID string) error {
	// The table we're interested in
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(userID),
			},
		},
		TableName: aws.String(repo.tableName),
	}

	_, err := repo.dynamoDBClient.DeleteItem(input)
	if err != nil {
		log.Warnf("Unable to delete user by user id: %s, error: %v", userID, err)
		return err
	}

	return nil
}

// GetUser retrieves the specified user using the user id
func (repo repository) GetUser(userID string) (*models.User, error) {

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
		TableName:                 aws.String(repo.tableName),
	}

	// Make the DynamoDB Query API call
	result, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.Warnf("Error retrieving user by user_id: %s, error: %+v", userID, err)
		return nil, err
	}

	// The user model
	var dbUserModels []DBUser

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &dbUserModels)
	if err != nil {
		log.Warnf("error unmarshalling user record from database for user_id: %s, error: %+v", userID, err)
		return nil, err
	}

	if len(dbUserModels) == 0 {
		return nil, nil
	} else if len(dbUserModels) > 1 {
		log.Warnf("retrieved %d results for the getUser(id) query when we should return 0 or 1", len(dbUserModels))
	}

	return convertDBUserModel(dbUserModels[0]), nil
}

// GetuserByLFUserName returns the user record associated with the LF Username value
func (repo repository) GetUserByLFUserName(lfUserName string) (*models.User, error) {

	// This is the key we want to match
	condition := expression.Key("lf_username").Equal(expression.Value(lfUserName))

	// These are the columns we want returned
	projection := buildUserProjection()

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(projection).Build()
	if err != nil {
		log.Warnf("error building expression for lf_username : %s, error: %v", lfUserName, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(repo.tableName),
		IndexName:                 aws.String("lf-username-index"),
	}

	// Make the DynamoDB Query API call
	result, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.Warnf("Error retrieving user by lf_username: %s, error: %+v", lfUserName, err)
		return nil, err
	}

	// The user model
	var dbUserModels []DBUser

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &dbUserModels)
	if err != nil {
		log.Warnf("error unmarshalling user record from database for lf_username: %s, error: %+v", lfUserName, err)
		return nil, err
	}

	if len(dbUserModels) == 0 {
		return nil, nil
	} else if len(dbUserModels) > 1 {
		log.Warnf("retrieved %d results for the getUser(id) query when we should return 0 or 1", len(dbUserModels))
	}

	return convertDBUserModel(dbUserModels[0]), nil

}

func (repo repository) GetUserByUserName(userName string, fullMatch bool) (*models.User, error) {

	var indexName string

	// This is the filter we want to match
	var condition expression.KeyConditionBuilder

	if strings.Contains(userName, "github:") {
		indexName = "github-user-index"
		// Username for Github comes in as github:123456, so we want to remove the initial string
		githubID, err := strconv.Atoi(strings.Replace(userName, "github:", "", 1))
		if err != nil {
			log.Warnf("Unable to convert Github ID to number: %s", err)
			return nil, err
		}
		condition = expression.Key("user_github_id").Equal(expression.Value(githubID))
	} else {
		indexName = "lf-username-index"
		condition = expression.Key("lf_username").Equal(expression.Value(userName))
	}

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
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(projection).Build()
	if err != nil {
		log.Warnf("error building expression for user name: %s, error: %v", userName, err)
		return nil, err
	}

	// Assemble the scan input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(repo.tableName),
		IndexName:                 aws.String(indexName),
	}

	var lastEvaluatedKey string
	// The database user model
	var dbUserModels []DBUser

	// Loop until we find a match or exhausted all the records
	for ok := true; ok; ok = lastEvaluatedKey != "" {
		// Make the DynamoDB Query API call
		result, err := repo.dynamoDBClient.Query(queryInput)
		if err != nil {
			log.Warnf("Error retrieving user by user name: %s, error: %+v", userName, err)
			return nil, err
		}

		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &dbUserModels)
		if err != nil {
			log.Warnf("error unmarshalling user record from database for user name: %s, error: %+v", userName, err)
			return nil, err
		}

		if len(dbUserModels) == 1 {
			return convertDBUserModel(dbUserModels[0]), nil
		} else if len(dbUserModels) > 1 {
			log.Warnf("retrieved %d results for the getUser(id) query when we should return 0 or 1", len(dbUserModels))
			return convertDBUserModel(dbUserModels[0]), nil
		}

		// Didn't find a match so far...need to keep looking via the next page of data

		// If we have another page of results...
		if result.LastEvaluatedKey["user_id"] != nil {
			lastEvaluatedKey = *result.LastEvaluatedKey["user_id"].S
			queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
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

// GetUserByEmail fetches the user record by email
func (repo repository) GetUserByEmail(userEmail string) (*models.User, error) {
	// This is the key we want to match
	condition := expression.Key("lf_email").Equal(expression.Value(userEmail))

	// These are the columns we want returned
	projection := buildUserProjection()

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(projection).Build()
	if err != nil {
		log.Warnf("error building expression for lf_email : %s, error: %v", userEmail, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(repo.tableName),
		IndexName:                 aws.String("lf-email-index"),
	}

	// Make the DynamoDB Query API call
	result, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.Warnf("error retrieving user by lf_email: %s, error: %+v", userEmail, err)
		return nil, err
	}

	// The user model
	var dbUserModels []DBUser

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &dbUserModels)
	if err != nil {
		log.Warnf("error unmarshalling user record from database for lf_email: %s, error: %+v", userEmail, err)
		return nil, err
	}

	if len(dbUserModels) == 0 {
		return nil, errors.NotFound("user not found when searching by lf_email: %s", userEmail)
	} else if len(dbUserModels) > 1 {
		log.Warnf("retrieved %d results for the lf_email query when we should return 0 or 1", len(dbUserModels))
	}

	return convertDBUserModel(dbUserModels[0]), nil
}

// GetUserByGitHubUsername fetches the user record by github username
func (repo repository) GetUserByGitHubUsername(gitHubUsername string) (*models.User, error) {
	// This is the key we want to match
	condition := expression.Key("user_github_username").Equal(expression.Value(gitHubUsername))

	// These are the columns we want returned
	projection := buildUserProjection()

	// Use the nice builder to create the expression
	expr, err := expression.NewBuilder().WithKeyCondition(condition).WithProjection(projection).Build()
	if err != nil {
		log.Warnf("error building expression for user_github_username : %s, error: %v", gitHubUsername, err)
		return nil, err
	}

	// Assemble the query input parameters
	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(repo.tableName),
		IndexName:                 aws.String("github-username-index"),
	}

	// Make the DynamoDB Query API call
	result, err := repo.dynamoDBClient.Query(queryInput)
	if err != nil {
		log.Warnf("error retrieving user by user_github_username: %s, error: %+v", gitHubUsername, err)
		return nil, err
	}

	// The user model
	var dbUserModels []DBUser

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &dbUserModels)
	if err != nil {
		log.Warnf("error unmarshalling user record from database for user_github_username: %s, error: %+v", gitHubUsername, err)
		return nil, err
	}

	if len(dbUserModels) == 0 {
		return nil, errors.NotFound("user not found when searching by user_github_username: %s", gitHubUsername)
	} else if len(dbUserModels) > 1 {
		log.Warnf("retrieved %d results for the user_github_username query when we should return 0 or 1", len(dbUserModels))
	}

	return convertDBUserModel(dbUserModels[0]), nil
}

func (repo repository) SearchUsers(searchField string, searchTerm string, fullMatch bool) (*models.Users, error) {
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
		TableName:                 aws.String(repo.tableName),
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
		TableName: &repo.tableName,
	}

	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total user record count for searchTerm: %s, error: %v", searchTerm, err)
		return nil, err
	}

	totalCount := *describeTableResult.Table.ItemCount

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
		UserExternalID: user.UserExternalID,
		Admin:          user.Admin,
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
		Note:           user.Note,
	}
}

func buildUserProjection() expression.ProjectionBuilder {
	// These are the columns we want returned
	return expression.NamesList(
		expression.Name("user_id"),
		expression.Name("user_external_id"),
		expression.Name("user_company_id"),
		expression.Name("admin"),
		expression.Name("lf_email"),
		expression.Name("lf_username"),
		expression.Name("user_name"),
		expression.Name("user_emails"),
		expression.Name("user_github_username"),
		expression.Name("user_github_id"),
		expression.Name("date_created"),
		expression.Name("date_modified"),
		expression.Name("version"),
		expression.Name("note"),
	)
}

// buildDBUserModels converts the database model into a service response data model
func buildDBUserModels(results *dynamodb.ScanOutput) ([]models.User, error) {
	var users []models.User

	// The DB company model
	var dbUsers []DBUser

	// Decode the database scan output into a database model
	err := dynamodbattribute.UnmarshalListOfMaps(results.Items, &dbUsers)
	if err != nil {
		log.Warnf("error unmarshalling users from database, error: %v", err)
		return nil, err
	}

	// Covert the database models to a list of API response models
	for _, dbUser := range dbUsers {
		users = append(users, *convertDBUserModel(dbUser))
	}

	return users, nil
}
