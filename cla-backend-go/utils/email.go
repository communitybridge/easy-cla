// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package utils

import (
	"errors"
	"fmt"

	"github.com/communitybridge/easycla/cla-backend-go/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
)

// EmailSender contains method to send email
type EmailSender interface {
	SendEmail(subject string, body string, recipients []string) error
}

var emailSender EmailSender

// SetEmailSender sets up default email sender
func SetEmailSender(es EmailSender) {
	emailSender = es
}

type snsEmail struct {
	snsClient          *sns.SNS
	snsEventTopicARN   string
	senderEmailAddress string
}

// SetSnsEmailSender set sns as mechanism to send email
func SetSnsEmailSender(awsSession *session.Session, snsEventTopicARN string, senderEmailAddress string) {
	emailSender = &snsEmail{
		snsClient:          sns.New(awsSession),
		snsEventTopicARN:   snsEventTopicARN,
		senderEmailAddress: senderEmailAddress,
	}
}

// SendEmail sends an email to the specified recipients
func (s *snsEmail) SendEmail(subject string, body string, recipients []string) error {
	event := CreateEventWrapper("cla-email-event")
	event.Data = ToEmailTemplateEvent(&s.senderEmailAddress, recipients, &subject, &body, "EasyCLA System Email Template")

	b, err := event.MarshalBinary()
	if err != nil {
		log.Warnf("Unable to marshal event model")
		return err
	}

	log.Debugf("Sending SNS message '%s' to topic: '%s'", b, s.snsEventTopicARN)
	input := &sns.PublishInput{
		Message:  aws.String(string(b)),          // Required
		TopicArn: aws.String(s.snsEventTopicARN), // Required
	}

	sendResp, err := s.snsClient.Publish(input)
	if err != nil {
		log.Warnf("Error publishing message to topic: %s, Error: %v", s.snsEventTopicARN, err)
		return err
	}

	log.Debugf("Successfully sent SNS message. Response: %v", sendResp)
	return nil
}

// SendEmail function send email. It uses emailSender interface.
func SendEmail(subject string, body string, recipients []string) error {
	if emailSender == nil {
		return errors.New("email sender not set")
	}
	return emailSender.SendEmail(subject, body, recipients)
}

// GetCorporateURL returns the corporate URL based on the specified flag
func GetCorporateURL(isV2Project bool) string {
	if isV2Project {
		return fmt.Sprintf("https://%s", config.GetConfig().CorporateConsoleV2URL)
	}

	return fmt.Sprintf("https://%s", config.GetConfig().CorporateConsoleURL)
}

// GetEmailHelpContent returns the standard email help paragraph details.
func GetEmailHelpContent(showV2HelpLink bool) string {
	if showV2HelpLink {
		return `<p>If you need help or have questions about EasyCLA, you can
<a href="https://docs.linuxfoundation.org/docs/v/v2/communitybridge/easycla" target="_blank">read the documentation</a> or
<a href="https://jira.linuxfoundation.org/servicedesk/customer/portal/4/create/143" target="_blank">reach out to us for
support</a>.</p>`
	}

	return `<p>If you need help or have questions about EasyCLA, you can
<a href="https://docs.linuxfoundation.org/docs/communitybridge/communitybridge-easycla" target="_blank">read the documentation</a> or
<a href="https://jira.linuxfoundation.org/servicedesk/customer/portal/4/create/143" target="_blank">reach out to us for
support</a>.</p>`
}

// GetEmailSignOffContent returns the standard email sign-off details
func GetEmailSignOffContent() string {
	return `<p>Thanks,</p>
<p>EasyCLA support team</p>`
}
