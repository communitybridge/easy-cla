package template

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/LF-Engineering/cla-monorepo/cla-backend-go/docraptor"
	"github.com/LF-Engineering/cla-monorepo/cla-backend-go/gen/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aymerick/raymond"
)

type Service interface {
	GetTemplates(ctx context.Context) ([]models.Template, error)
	CreateCLAGroupTemplate(ctx context.Context, claGroupID string, claGroupFields *models.CreateClaGroupTemplate)
}

type service struct {
	templateRepo    Repository
	docraptorClient docraptor.DocraptorClient
	s3Client        *s3manager.Uploader
}

func NewService(templateRepo Repository, docraptorClient docraptor.DocraptorClient, awsSession *session.Session) service {
	return service{
		templateRepo:    templateRepo,
		docraptorClient: docraptorClient,
		s3Client:        s3manager.NewUploader(awsSession),
	}
}

func (s service) GetTemplates(ctx context.Context) ([]models.Template, error) {
	templates, err := s.templateRepo.GetTemplates()
	if err != nil {
		return nil, err
	}

	// Remove HTML from template
	for i, template := range templates {
		template.HTMLBody = ""
		templates[i] = template
	}

	return templates, nil
}

func (s service) CreateCLAGroupTemplate(ctx context.Context, claGroupID string, claGroupFields *models.CreateClaGroupTemplate) error {
	// Verify claGroupID matches an existing CLA Group
	_, err := s.templateRepo.GetCLAGroup(claGroupID)
	if err != nil {
		return err
	}

	// Verify the caller is authorized for the project that owns this CLA Group

	// Get Template
	template, err := s.templateRepo.GetTemplate(claGroupFields.TemplateID)
	if err != nil {
		return err
	}

	// Apply template fields
	templateHTML, err := s.InjectProjectInformationIntoTemplate(template, claGroupFields.MetaFields)
	if err != nil {
		return err
	}

	// Create PDF
	pdf, err := s.docraptorClient.CreatePDF(templateHTML)
	if err != nil {
		return err
	}
	defer pdf.Close()

	// Save PDF to S3
	bucket := "cla-signature-files-dev"
	fileNameTemplate := "contract-group/%s/template/%s"
	iclaFileName := fmt.Sprintf(fileNameTemplate, claGroupID, "icla.pdf")
	// cclaFileName := fmt.Sprintf(fileNameTemplate, claGroupID, "ccla.pdf")
	err = s.SaveTemplateToS3(bucket, iclaFileName, pdf)
	if err != nil {
		return err
	}

	// Save Template to Dynamodb
	tableName := "cla-dev-projects"
	err = s.templateRepo.UpdateDynamoContractGroupTemplates(ctx, claGroupID, tableName, template)
	if err != nil {
		return err
	}

	return nil
}

func (s service) InjectProjectInformationIntoTemplate(template models.Template, fields []*models.MetaField) (string, error) {
	// TODO: Verify all template fields in template.MetaFields are present

	lookupMap := map[string]models.MetaField{}
	for _, field := range template.MetaFields {
		lookupMap[field.Name] = *field
	}

	fieldsMap := map[string]string{}
	for _, field := range fields {

		val, ok := lookupMap[field.Name]
		if !ok {
			continue
		}

		if val.Name == field.Name && val.TemplateVariable == field.TemplateVariable {
			fieldsMap[field.TemplateVariable] = field.Value
		}
	}
	if len(template.MetaFields) != len(fieldsMap) {
		return "", errors.New("Required fields for template were not found")
	}

	templateHTML, err := raymond.Render(template.HTMLBody, fieldsMap)
	if err != nil {
		return "", err
	}

	return templateHTML, nil
}

func (s service) SaveTemplateToS3(bucket, filepath string, template io.ReadCloser) error {
	defer template.Close()

	// Upload the file to S3.
	_, err := s.s3Client.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filepath),
		Body:   template,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3 Bucket, %v", err)
	}

	return nil
}
