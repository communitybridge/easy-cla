package dynamo_events

import (
	"github.com/aws/aws-lambda-go/events"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	v2ProjectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
	"github.com/sirupsen/logrus"
)

// Event data model
type Event struct {
	EventID        string `json:"event_id"`
	EventProjectID string `json:"event_project_id"`
	EventCompanyID string `json:"event_company_id"`
}

// should be called when we insert Event
func (s *service) EventAddedEvent(event events.DynamoDBEventRecord) error {
	var newEvent Event
	err := unmarshalStreamImage(event.Change.NewImage, &newEvent)
	if err != nil {
		return err
	}
	f := logrus.Fields{"event": newEvent}
	var foundationSFID, projectSFID, projectSFName, companySFID string
	companyModel, err := s.companyRepo.GetCompany(newEvent.EventCompanyID)
	if err != nil {
		log.WithFields(f).Error("unable to get company detail", err)
	} else {
		companySFID = companyModel.CompanyExternalID
	}
	pmList, err := s.projectsClaGroupRepo.GetProjectsIdsForClaGroup(newEvent.EventProjectID)
	if err != nil || len(pmList) == 0 {
		log.WithFields(f).Error("unable to get project mapping detail", err)
	} else {
		if len(pmList) > 1 {
			foundationSFID = pmList[0].FoundationSFID
			projectSFID = pmList[0].FoundationSFID
			psc := v2ProjectService.GetClient()
			projectDetails, perr := psc.GetProject(foundationSFID)
			if perr != nil {
				log.WithFields(f).WithField("foundation_sfid", foundationSFID).Error("unable to fetch foundation details", perr)
			} else {
				projectSFName = projectDetails.Name
			}
		} else {
			foundationSFID = pmList[0].FoundationSFID
			projectSFID = pmList[0].ProjectSFID
			projectSFName = pmList[0].ProjectName
		}
	}
	err = s.eventsRepo.AddDataToEvent(newEvent.EventID, foundationSFID, projectSFID, projectSFName, companySFID, newEvent.EventProjectID)
	if err != nil {
		return err
	}
	return nil
}
