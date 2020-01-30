// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package users

import (
	"fmt"
	"strconv"
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
	GetMetrics() (*models.UserMetrics, error)
	CreateUser(user *models.User) (*models.User, error)
	Save(user *models.UserUpdate) (*models.User, error)
	Delete(userID string) error
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

// CreateUser creates a new user
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
	attributes["date_created"] = &dynamodb.AttributeValue{
		S: aws.String(time.Now().UTC().Format(time.RFC3339)),
	}
	attributes["date_modified"] = &dynamodb.AttributeValue{
		S: aws.String(time.Now().UTC().Format(time.RFC3339)),
	}
	attributes["version"] = &dynamodb.AttributeValue{
		S: aws.String("v1"),
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

// Save saves the user model to the data store
func (repo repository) Save(user *models.UserUpdate) (*models.User, error) {
	// The table we're interested in
	tableName := fmt.Sprintf("cla-%s-users", repo.stage)

	log.Debugf("Save User - looking up user by username: %s", user.LfUsername)
	oldUserModel, err := repo.GetUserByUserName(user.LfUsername, true)
	if err != nil || oldUserModel == nil {
		log.Warnf("Error fetching existing user record: %+v, error: %v", user, err)
		return nil, err
	}

	log.Debugf("Found user by username: %+v", oldUserModel)

	// return values flag - Returns all of the attributes of the item, as they appear after the UpdateItem operation.
	addReturnValues := "ALL_NEW" // nolint

	updatedDateTime := time.Now().UTC()
	expressionAttributeNames := map[string]*string{}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{}
	updateExpression := "SET "

	if user.LfEmail != "" {
		log.Debugf("Save User - adding lf_email: %s", user.LfEmail)
		expressionAttributeNames["#E"] = aws.String("lf_email")
		expressionAttributeValues[":e"] = &dynamodb.AttributeValue{S: aws.String(user.LfEmail)}
		updateExpression = updateExpression + " #E = :e, "
	}

	if user.CompanyID != "" {
		log.Debugf("Save User - adding user_company_id: %s", user.CompanyID)
		expressionAttributeNames["#C"] = aws.String("user_company_id")
		expressionAttributeValues[":c"] = &dynamodb.AttributeValue{S: aws.String(user.CompanyID)}
		updateExpression = updateExpression + " #C = :c, "
	}

	log.Debugf("Save User - adding date_modified: %s", updatedDateTime.Format(time.RFC3339))
	expressionAttributeNames["#D"] = aws.String("date_modified")
	expressionAttributeValues[":d"] = &dynamodb.AttributeValue{S: aws.String(updatedDateTime.Format(time.RFC3339))}
	updateExpression = updateExpression + " #D = :d "

	// Update dynamoDB table
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
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
		log.Warnf("Error updating user record: %+v, error: %v", user, err)
		return nil, err
	}

	log.Debugf("Save User - looking up saved user by username: %s", user.LfUsername)
	newUserModel, err := repo.GetUserByUserName(user.LfUsername, true)
	if err != nil || newUserModel == nil {
		log.Warnf("Error fetching updated user record: %+v, error: %v", user, err)
		return nil, err
	}

	log.Debugf("Returning updated user: %+v", newUserModel)
	return newUserModel, err
}

// Delete deletes the specified user
func (repo repository) Delete(userID string) error {
	// The table we're interested in
	tableName := fmt.Sprintf("cla-%s-users", repo.stage)
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(userID),
			},
		},
		TableName: aws.String(tableName),
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

func (repo repository) GetUserByUserName(userName string, fullMatch bool) (*models.User, error) {

	tableName := fmt.Sprintf("cla-%s-users", repo.stage)

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
		TableName:                 aws.String(tableName),
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

func (repo repository) GetMetrics() (*models.UserMetrics, error) {
	var out models.UserMetrics
	tableName := fmt.Sprintf("cla-%s-users", repo.stage)
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	describeTableResult, err := repo.dynamoDBClient.DescribeTable(describeTableInput)
	if err != nil {
		log.Warnf("error retrieving total record count of users, error: %v", err)
		return nil, err
	}

	out.TotalCount = *describeTableResult.Table.ItemCount
	return &out, nil
}
