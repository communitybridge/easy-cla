package dynamo_events

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"

	"github.com/communitybridge/easycla/cla-backend-go/company"

	"github.com/communitybridge/easycla/cla-backend-go/signatures"

	"github.com/sirupsen/logrus"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// constants
const (
	Insert = "INSERT"
	Modify = "MODIFY"
)

// EventHandlerFunc is type for dynamoDB event handler function
type EventHandlerFunc func(event events.DynamoDBEventRecord) error

type service struct {
	// key : tablename:action
	functions            map[string][]EventHandlerFunc
	signatureRepo        signatures.SignatureRepository
	companyRepo          company.IRepository
	projectsClaGroupRepo projects_cla_groups.Repository
}

// Service implements DynamoDB stream event handler service
type Service interface {
	ProcessEvents(event events.DynamoDBEvent)
}

// NewService creates DynamoDB stream event handler service
func NewService(stage string, signatureRepo signatures.SignatureRepository, companyRepo company.IRepository, pcgRepo projects_cla_groups.Repository) Service {
	SignaturesTable := fmt.Sprintf("cla-%s-signatures", stage)
	s := &service{
		functions:            make(map[string][]EventHandlerFunc),
		signatureRepo:        signatureRepo,
		companyRepo:          companyRepo,
		projectsClaGroupRepo: pcgRepo,
	}
	s.registerCallback(SignaturesTable, Modify, s.SignatureSignedEvent)
	s.registerCallback(SignaturesTable, Modify, s.SignatureAddSigTypeSignedApprovedID)
	s.registerCallback(SignaturesTable, Insert, s.SignatureAddSigTypeSignedApprovedID)
	s.registerCallback(SignaturesTable, Insert, s.SignatureAddUsersDetails)
	return s
}

func (s *service) registerCallback(tableName, eventName string, callbackFunction EventHandlerFunc) {
	key := fmt.Sprintf("%s:%s", tableName, eventName)
	funcArr := s.functions[key]
	funcArr = append(funcArr, callbackFunction)
	s.functions[key] = funcArr
}

func (s *service) ProcessEvents(events events.DynamoDBEvent) {
	for _, event := range events.Records {
		tableName := strings.Split(event.EventSourceArn, "/")[1]
		fields := logrus.Fields{
			"table_name": tableName,
			"event":      event.EventName,
		}
		b, _ := json.Marshal(events) // nolint
		fields["events_data"] = string(b)
		log.WithFields(fields).Debug("Processing event")
		key := fmt.Sprintf("%s:%s", tableName, event.EventName)
		for _, f := range s.functions[key] {
			err := f(event)
			if err != nil {
				log.WithFields(fields).WithField("event", event).Error("unable to process event", err)
			}
		}
	}
}

// UnmarshalStreamImage converts events.DynamoDBAttributeValue to struct
func unmarshalStreamImage(attribute map[string]events.DynamoDBAttributeValue, out interface{}) error {
	dbAttrMap := make(map[string]*dynamodb.AttributeValue)
	for k, v := range attribute {
		var dbAttr dynamodb.AttributeValue
		bytes, marshalErr := v.MarshalJSON()
		if marshalErr != nil {
			return marshalErr
		}
		err := json.Unmarshal(bytes, &dbAttr)
		if err != nil {
			return err
		}
		dbAttrMap[k] = &dbAttr
	}
	return dynamodbattribute.UnmarshalMap(dbAttrMap, out)
}
