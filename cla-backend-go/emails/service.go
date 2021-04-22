// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package emails

import (
	"context"
	"fmt"

	"github.com/communitybridge/easycla/cla-backend-go/project"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
)

// Service is a service with some helper functions for rendering templates and also sending emails
type Service interface {
	EmailTemplateService
	NotifyClaManagersForClaGroupID(ctx context.Context, claGrpoupID, subject, body string) error
}

type service struct {
	EmailTemplateService
	claService project.Service
}

// NewService is constructor for emails.Service
func NewService(emailTemplateService EmailTemplateService, claService project.Service) Service {
	return &service{
		EmailTemplateService: emailTemplateService,
		claService:           claService,
	}
}

func (s *service) NotifyClaManagersForClaGroupID(ctx context.Context, claGrpoupID, subject, body string) error {
	claManagers, err := s.claService.GetCLAManagers(ctx, claGrpoupID)
	if err != nil {
		return fmt.Errorf("fetching cla manager for cla group : %s failed : %v", claGrpoupID, err)
	}

	if len(claManagers) == 0 {
		return fmt.Errorf("no cla managers registered for the claGroup : %s, none to notify", claGrpoupID)
	}

	var recipientEmails []string
	for _, claManager := range claManagers {
		recipientEmails = append(recipientEmails, claManager.UserEmail)
	}

	return utils.SendEmail(subject, body, recipientEmails)
}
