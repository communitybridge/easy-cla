// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	// "fmt"
	"os"
	"sync"
	"flag"

	// "sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/communitybridge/easycla/cla-backend-go/company"

	// "github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/communitybridge/easycla/cla-backend-go/events"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v1/models"
	eventOps "github.com/communitybridge/easycla/cla-backend-go/gen/v1/restapi/operations/events"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/project/repository"
	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"
	"github.com/communitybridge/easycla/cla-backend-go/repositories"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/users"
	"github.com/sirupsen/logrus"
)

var stage string
var eventsRepo events.Repository
var eventsService events.Service
var signatureRepo signatures.SignatureRepository
var v1ProjectRepo repository.ProjectRepository
var usersRepo users.UserRepository
var companyRepo company.IRepository
var ghRepo repositories.Repository
var projectClaGroup projects_cla_groups.Repository
var awsSession = session.Must(session.NewSession(&aws.Config{}))

type combinedRepo struct {
	users.UserRepository
	company.IRepository
	repository.ProjectRepository
	projects_cla_groups.Repository
}

func init() {
	stage = os.Getenv("STAGE")
	if stage == "" {
		log.Fatal("STAGE environment variable not set")
	}

	log.Infof("STAGE: %s", stage)
	companyRepo = company.NewRepository(awsSession, stage)
	usersRepo = users.NewRepository(awsSession, stage)
	ghRepo = *repositories.NewRepository(awsSession, stage)
	projectClaGroup = projects_cla_groups.NewRepository(awsSession, stage)
	v1ProjectRepo = repository.NewRepository(awsSession, stage, &ghRepo, nil, projectClaGroup)
	eventsRepo = events.NewRepository(awsSession, stage)
	eventsService = events.NewService(eventsRepo, combinedRepo{
		usersRepo,
		companyRepo,
		v1ProjectRepo,
		projectClaGroup,
	})
	signatureRepo = signatures.NewRepository(awsSession, stage, companyRepo, usersRepo, eventsService, &ghRepo, nil, nil, nil)

	log.Infof("initializing migrate_ecla")
}

func main() {
	f := logrus.Fields{
		"functionName": "main",
		"stage":        stage,
	}

	dryRun := flag.Bool("dry-run",true, "boolean flag that can update the signature records")
	flag.Parse()
	log.WithFields(f).Info("migrate_ecla started")

	ctx := context.Background()
	// Fetch all the companies
	companies, err := companyRepo.GetCompanies(ctx)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem fetching companies")
		return
	}

	invalidSignatures := make([]*signatures.ItemSignature, 0)

	log.WithFields(f).Infof("processing %d companies", len(companies.Companies))
	// Fetch  the ecla records

	for _, company := range companies.Companies {
		// log.WithFields(f).Infof("processing company: %s", company.CompanyName)
		// Fetch the ecla records for the company
		invalidCompanyEcla := make([]*signatures.ItemSignature, 0)
		eclaRecords, err := signatureRepo.GetCompanyECLASignatures(ctx, company.CompanyID)
		if err != nil {
			log.WithFields(f).WithError(err).Warn("problem fetching ecla records")
			return
		}
		// log.WithFields(f).Infof("processing %d records for company: %s", len(eclaRecords), company.CompanyID)

		if len(eclaRecords) == 0 {
			log.WithFields(f).Infof("no records for company: %s", company.CompanyID)
			continue
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		for _, eclaRecord := range eclaRecords {
			wg.Add(1)
			go func(eclaRecord *signatures.ItemSignature) {
				defer wg.Done()
				// log.WithFields(f).Infof("processing signature record: %s", eclaRecord.SignatureID)
				if eclaRecord.DateCreated == "" {
					log.WithFields(f).Warnf("invalid signature record: %s", eclaRecord.SignatureID)
					mu.Lock()
					invalidCompanyEcla = append(invalidCompanyEcla, eclaRecord)
					mu.Unlock()
					return
				}
			}(eclaRecord)
		}
		wg.Wait()

		invalidSignatures = append(invalidSignatures, invalidCompanyEcla...)

		log.WithFields(f).Debugf("Found %d invalid eclas for company: %s", len(invalidCompanyEcla), company.CompanyID)
	}

	log.WithFields(f).Infof("Processesing %d invalid records ", len(invalidSignatures))
	// For each ecla record, check the events table for the corresponding event and update the signature table
	update := 0

	for _, eclaRecord := range invalidSignatures {
		events, err := getEvents(ctx, *eclaRecord)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("problem fetching events for ecla record: %s", eclaRecord.SignatureID)
			return
		}

		if len(events) == 0 {
			log.WithFields(f).Warnf("no events found for ecla record: %s and user_id: %s", eclaRecord.SignatureID, eclaRecord.SignatureReferenceID)
			continue
		}

		log.WithFields(f).Infof("found %d events for ecla record: %s", len(events), eclaRecord.SignatureID)

		latestEvent := getLatestEvent(events)
		if latestEvent == nil {
			log.WithFields(f).Warnf("no latest event found for ecla record: %s", eclaRecord.SignatureID)
			continue
		}

		log.Debugf("Found event:%s for user %s", latestEvent.EventID, eclaRecord.SignatureReferenceID)

		// Update the signature record
		update++

		update := map[string]interface{}{
			"date_created":  latestEvent.EventTime,
			"date_modified": latestEvent.EventTime,
		}

		log.WithFields(f).Infof("updating signature record: %s with : %+v", eclaRecord.SignatureID, update)

		if !*dryRun {
			err = signatureRepo.UpdateSignature(ctx, eclaRecord.SignatureID, update)
			if err != nil {
				log.WithFields(f).WithError(err).Warnf("problem updating signature record: %s", eclaRecord.SignatureID)
				return
			}
		} else {
			log.WithFields(f).Info("dry run....")
		}

		log.WithFields(f).Infof("updated signature record: %s", eclaRecord.SignatureID)
	}

	log.WithFields(f).Infof("updated :%d records", update)
}

func getEvents(ctx context.Context, ecla signatures.ItemSignature) ([]*models.Event, error) {

	f := logrus.Fields{
		"functionName": "getEvents",
		"companyID":    ecla.SignatureUserCompanyID,
		"projectID":    ecla.SignatureProjectID,
		"userID":       ecla.SignatureReferenceID,
	}

	log.WithFields(f).Debug("searching events")

	eventType := "EmployeeSignatureCreated"

	claGroupModel, err := v1ProjectRepo.GetCLAGroupByID(ctx, ecla.SignatureProjectID, false)
	if err != nil {
		log.WithFields(f).Debugf("Unable to fetch cla group model")
		return nil, err
	}

	eventParams := eventOps.SearchEventsParams{
		UserID:      &ecla.SignatureReferenceID,
		ProjectSFID: &claGroupModel.ProjectExternalID,
		EventType:   &eventType,
		CompanyID:   &ecla.SignatureUserCompanyID,
	}

	events, err := eventsService.SearchEvents(&eventParams)
	if err != nil {
		log.WithFields(f).WithError(err).Infof("Unable to get clagroup by id: %s", claGroupModel.ProjectID)
		return nil, err
	}

	return events.Events, nil

}

func getLatestEvent(events []*models.Event) *models.Event {
	// Fetch the latest event from the list of events
	var latestEvent *models.Event
	for _, event := range events {
		if latestEvent == nil {
			latestEvent = event
			continue
		}
		if event.EventTime > latestEvent.EventTime {
			latestEvent = event
		}
	}
	return latestEvent
}
