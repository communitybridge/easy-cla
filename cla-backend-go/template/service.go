// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package template

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	"github.com/communitybridge/easycla/cla-backend-go/docraptor"
	"github.com/communitybridge/easycla/cla-backend-go/gen/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aymerick/raymond"
)

const (
	claTypeICLA = "icla"
	claTypeCCLA = "ccla"
)

// Service interface
type Service interface {
	GetTemplates(ctx context.Context) ([]models.Template, error)
	CreateCLAGroupTemplate(ctx context.Context, claGroupID string, claGroupFields *models.CreateClaGroupTemplate) (models.TemplatePdfs, error)
	CreateTemplatePreview(claGroupFields *models.CreateClaGroupTemplate, templateFor string) ([]byte, error)
	GetCLATemplatePreview(ctx context.Context, claGroupID, claType string, watermark bool) ([]byte, error)
}

type service struct {
	stage           string // The AWS stage (dev, staging, prod)
	templateRepo    Repository
	docraptorClient docraptor.Client
	s3Client        *s3manager.Uploader
}

// NewService API call
func NewService(stage string, templateRepo Repository, docraptorClient docraptor.Client, awsSession *session.Session) service {
	return service{
		stage:           stage,
		templateRepo:    templateRepo,
		docraptorClient: docraptorClient,
		s3Client:        s3manager.NewUploader(awsSession),
	}
}

// GetTemplates API call
func (s service) GetTemplates(ctx context.Context) ([]models.Template, error) {
	templates, err := s.templateRepo.GetTemplates()
	if err != nil {
		return nil, err
	}

	// Remove HTML from template
	for i, template := range templates {
		template.IclaHTMLBody = ""
		template.CclaHTMLBody = ""
		templates[i] = template
	}

	return templates, nil
}

func (s service) CreateTemplatePreview(claGroupFields *models.CreateClaGroupTemplate, templateFor string) ([]byte, error) {
	var template models.Template
	var err error
	if claGroupFields.TemplateID != "" {
		// Get Template
		template, err = s.templateRepo.GetTemplate(claGroupFields.TemplateID)
		if err != nil {
			log.Warnf("Unable to fetch template fields: %s, error: %v",
				claGroupFields.TemplateID, err)
			return nil, err
		}
	} else {
		// use default Apache template if template_id is not provided
		template, err = s.templateRepo.GetTemplate(ApacheStyleTemplateID)
		if err != nil {
			log.Warnf("Unable to fetch default template fields: %s, error: %v",
				claGroupFields.TemplateID, err)
			return nil, err
		}
	}

	// Apply template fields
	iclaTemplateHTML, cclaTemplateHTML, err := s.InjectProjectInformationIntoTemplate(template, claGroupFields.MetaFields)
	if err != nil {
		log.Warnf("Unable to inject metadata details into template, error: %v", err)
		return nil, err
	}
	var templateHTML string
	switch templateFor {
	case claTypeICLA:
		templateHTML = iclaTemplateHTML
	case claTypeCCLA:
		templateHTML = cclaTemplateHTML
	default:
		return nil, errors.New("invalid value of template_for")
	}
	pdf, err := s.docraptorClient.CreatePDF(templateHTML)
	if err != nil {
		return nil, err
	}
	defer pdf.Close()
	return ioutil.ReadAll(pdf)
}

// CreateCLAGroupTemplate
func (s service) CreateCLAGroupTemplate(ctx context.Context, claGroupID string, claGroupFields *models.CreateClaGroupTemplate) (models.TemplatePdfs, error) {
	f := logrus.Fields{
		"functionName":   "CreateCLAGroupTemplate",
		"claGroupID":     claGroupID,
		"claGroupFields": claGroupFields,
	}

	// Verify claGroupID matches an existing CLA Group
	claGroup, err := s.templateRepo.GetCLAGroup(claGroupID)
	if err != nil {
		log.WithFields(f).Warnf("Unable to fetch CLA group by id: %s, error: %v - returning empty template PDFs", claGroupID, err)
		return models.TemplatePdfs{}, err
	}

	// Verify the caller is authorized for the project that owns this CLA Group

	// Get Template
	template, err := s.templateRepo.GetTemplate(claGroupFields.TemplateID)
	if err != nil {
		log.WithFields(f).Warnf("Unable to fetch template fields: %s, error: %v - returning empty template PDFs",
			claGroupFields.TemplateID, err)
		return models.TemplatePdfs{}, err
	}

	// Apply template fields
	iclaTemplateHTML, cclaTemplateHTML, err := s.InjectProjectInformationIntoTemplate(template, claGroupFields.MetaFields)
	if err != nil {
		log.WithFields(f).Warnf("Unable to inject metadata details into template, error: %v - returning empty template PDFs", err)
		return models.TemplatePdfs{}, err
	}

	bucket := fmt.Sprintf("cla-signature-files-%s", s.stage)

	// Create PDF
	var pdfUrls models.TemplatePdfs
	var iclaFileURL string
	var cclaFileURL string

	// Use an error group to keep track of errors thrown in the below go routines
	// Using go routines sped up the logic from ~8 seconds to ~5 seconds as we wait for the generation to complete
	var eg errgroup.Group

	if claGroup.ProjectICLAEnabled {
		// Invoke the go routine - any errors will be handled below
		eg.Go(func() error {
			log.WithFields(f).Debugf("Creating PDF for %s", claTypeICLA)
			iclaPdf, iclaErr := s.docraptorClient.CreatePDF(iclaTemplateHTML)
			if iclaErr != nil {
				log.WithFields(f).Warnf("Problem generating ICLA template via docraptor client, error: %v - returning empty template PDFs", err)
				return err
			}
			defer func() {
				closeErr := iclaPdf.Close()
				if closeErr != nil {
					log.WithFields(f).Warnf("error closing ICLA PDF, error: %v", closeErr)
				}
			}()
			iclaFileName := s.generateTemplateS3FilePath(claGroupID, claTypeICLA)
			iclaFileURL, err = s.SaveTemplateToS3(bucket, iclaFileName, iclaPdf)
			if err != nil {
				log.WithFields(f).Warnf("Problem uploading ICLA PDF: %s to s3, error: %v - returning empty template PDFs", iclaFileName, err)
				return err
			}

			template.IclaHTMLBody = iclaTemplateHTML
			return nil
		})
	}

	if claGroup.ProjectCCLAEnabled {
		// Invoke the go routine - any errors will be handled below
		eg.Go(func() error {
			log.WithFields(f).Debugf("Creating PDF for %s", claTypeCCLA)
			cclaPdf, cclaErr := s.docraptorClient.CreatePDF(cclaTemplateHTML)
			if cclaErr != nil {
				log.WithFields(f).Warnf("Problem generating CCLA template via docraptor client, error: %v - returning empty template PDFs", err)
				return err
			}
			defer func() {
				closeErr := cclaPdf.Close()
				if closeErr != nil {
					log.WithFields(f).Warnf("error closing CCLA PDF, error: %v", closeErr)
				}
			}()
			cclaFileName := s.generateTemplateS3FilePath(claGroupID, claTypeCCLA)
			cclaFileURL, err = s.SaveTemplateToS3(bucket, cclaFileName, cclaPdf)
			if err != nil {
				log.WithFields(f).Warnf("Problem uploading CCLA PDF: %s to s3, error: %v - returning empty template PDFs", cclaFileName, err)
				return err
			}

			template.CclaHTMLBody = cclaTemplateHTML
			return nil
		})
	}

	// Wait for the go routines to finish
	log.WithFields(f).Debug("Waiting for PDF generation to complete...")
	if pdfErr := eg.Wait(); pdfErr != nil {
		return models.TemplatePdfs{}, pdfErr
	}

	if claGroup.ProjectICLAEnabled && claGroup.ProjectCCLAEnabled {
		pdfUrls = models.TemplatePdfs{
			IndividualPDFURL: iclaFileURL,
			CorporatePDFURL:  cclaFileURL,
		}
	} else if claGroup.ProjectCCLAEnabled {
		pdfUrls = models.TemplatePdfs{
			CorporatePDFURL: cclaFileURL,
		}
	} else if claGroup.ProjectICLAEnabled {
		pdfUrls = models.TemplatePdfs{
			IndividualPDFURL: iclaFileURL,
		}
	}

	// Save Template to DynamoDB
	err = s.templateRepo.UpdateDynamoContractGroupTemplates(ctx, claGroupID, template, pdfUrls, claGroup.ProjectCCLAEnabled, claGroup.ProjectICLAEnabled)
	if err != nil {
		log.WithFields(f).Warnf("Problem updating the database with ICLA/CCLA new PDF details, error: %v - returning empty template PDFs", err)
		return models.TemplatePdfs{}, err
	}

	return pdfUrls, nil
}

func (s service) GetCLATemplatePreview(ctx context.Context, claGroupID, claType string, watermark bool) ([]byte, error) {
	// Verify claGroupID matches an existing CLA Group
	claGroup, err := s.templateRepo.GetCLAGroup(claGroupID)
	if err != nil {
		log.Warnf("unable to fetch CLA group by id: %s, error: %v - returning empty PDF", claGroupID, err)
		return nil, err
	}

	var projectDocuments []models.ProjectDocument

	switch claType {
	case claTypeICLA:
		if !claGroup.ProjectICLAEnabled {
			err = fmt.Errorf("icla required for the group id : %s, but not enabled", claGroupID)
			log.WithError(err)
			return nil, err
		}

	case claTypeCCLA:
		if !claGroup.ProjectCCLAEnabled {
			err = fmt.Errorf("ccla required for the group id : %s, but not enabled", claGroupID)
			log.WithError(err)
			return nil, err
		}

	default:
		err = fmt.Errorf("not supported cla type provided : %s", claType)
		return nil, err
	}

	projectDocuments, err = s.templateRepo.GetCLADocuments(claGroupID, claType)
	if err != nil {
		log.Warnf("fetching icla document failed for claGroupID : %s : %v", claGroupID, err)
		return nil, err
	}

	// process the documents and try to fetch the document from s3
	if len(projectDocuments) == 0 {
		err = fmt.Errorf("no documents found in groupID : %s", claGroupID)
		return nil, err
	}

	doc := projectDocuments[0]
	pdfS3URL := doc.DocumentS3URL
	if pdfS3URL == "" {
		err = fmt.Errorf("s3 url is empty for groupID : %s and document %s", claGroupID, doc.DocumentFileID)
		return nil, err
	}

	//fetch the document from s3 at this stage
	fileName := s.generateTemplateS3FilePath(claGroupID, claType)
	b, err := utils.DownloadFromS3(fileName)
	if err != nil {
		return nil, err
	}

	// do the watermarking here if enabled
	if watermark {
		b, err = utils.WatermarkPdf(b, "Not for Execution")
		if err != nil {
			log.WithError(err)
			return nil, err
		}
	}

	return b, nil
}

// InjectProjectInformationIntoTemplate
func (s service) InjectProjectInformationIntoTemplate(template models.Template, metaFields []*models.MetaField) (string, string, error) {
	lookupMap := map[string]models.MetaField{}
	for _, field := range template.MetaFields {
		lookupMap[field.Name] = *field
	}

	metaFieldsMap := map[string]string{}
	for _, metaField := range metaFields {

		val, ok := lookupMap[metaField.Name]
		if !ok {
			continue
		}

		if val.Name == metaField.Name && val.TemplateVariable == metaField.TemplateVariable {
			if metaField.Value == "" {
				return "", "", fmt.Errorf("bad request: template field value of variable %s cannot be empty", metaField.TemplateVariable)
			}
			metaFieldsMap[metaField.TemplateVariable] = metaField.Value
		}
	}
	if len(template.MetaFields) != len(metaFieldsMap) {
		return "", "", errors.New("bad request: required fields for template were not found")
	}

	iclaTemplateHTML, err := raymond.Render(template.IclaHTMLBody, metaFieldsMap)
	if err != nil {
		return "", "", err
	}

	cclaTemplateHTML, err := raymond.Render(template.CclaHTMLBody, metaFieldsMap)
	if err != nil {
		return "", "", err
	}

	return iclaTemplateHTML, cclaTemplateHTML, nil
}

func (s service) generateTemplateS3FilePath(claGroupID, claType string) string {
	fileNameTemplate := "contract-group/%s/template/%s"
	var ext string
	switch claType {
	case claTypeICLA:
		ext = "icla.pdf"
	case claTypeCCLA:
		ext = "ccla.pdf"
	default:
		return ""
	}
	fileName := fmt.Sprintf(fileNameTemplate, claGroupID, ext)
	return fileName
}

// SaveTemplateToS3
func (s service) SaveTemplateToS3(bucket, filepath string, template io.ReadCloser) (string, error) {
	defer template.Close()

	// Upload the file to S3.
	result, err := s.s3Client.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(filepath),
		Body:        template,
		ACL:         aws.String("public-read"),
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3 Bucket: %s / %s, %v", bucket, filepath, err)
	}

	return result.Location, nil
}
