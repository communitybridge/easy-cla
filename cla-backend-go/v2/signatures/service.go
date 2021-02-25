// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package signatures

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"

	"github.com/jinzhu/copier"

	"github.com/communitybridge/easycla/cla-backend-go/company"
	v1Models "github.com/communitybridge/easycla/cla-backend-go/gen/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/project"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/sirupsen/logrus"
)

// constants
const (
	// used when we want to query all data from dependent service.
	HugePageSize      = int64(10000)
	CclaSignatureType = "ccla"
	ClaSignatureType  = "cla"
)

// errors
var (
	ErrZipNotPresent = errors.New("zip file not present")
)

type service struct {
	v1ProjectService      project.Service
	v1CompanyService      company.IService
	v1SignatureService    signatures.SignatureService
	projectsClaGroupsRepo projects_cla_groups.Repository
	s3                    *s3.S3
	signaturesBucket      string
}

// Service contains method of v2 signature service
type Service interface {
	GetProjectCompanySignatures(ctx context.Context, companyID, companySFID, projectSFID string) (*models.Signatures, error)
	GetProjectIclaSignaturesCsv(ctx context.Context, claGroupID string) ([]byte, error)
	GetProjectCclaSignaturesCsv(ctx context.Context, claGroupID string) ([]byte, error)
	GetProjectIclaSignatures(ctx context.Context, claGroupID string, searchTerm *string) (*models.IclaSignatures, error)
	GetClaGroupCorporateContributorsCsv(ctx context.Context, claGroupID string, companyID string) ([]byte, error)
	GetClaGroupCorporateContributors(ctx context.Context, claGroupID string, companySFID string, searchTerm *string) (*models.CorporateContributorList, error)
	GetSignedDocument(ctx context.Context, signatureID string) (*models.SignedDocument, error)
	GetSignedIclaZipPdf(claGroupID string) (*models.URLObject, error)
	GetSignedCclaZipPdf(claGroupID string) (*models.URLObject, error)
}

// NewService creates instance of v2 signature service
func NewService(awsSession *session.Session, signaturesBucketName string, v1ProjectService project.Service,
	v1CompanyService company.IService,
	v1SignatureService signatures.SignatureService,
	pcgRepo projects_cla_groups.Repository) *service {
	return &service{
		v1ProjectService:      v1ProjectService,
		v1CompanyService:      v1CompanyService,
		v1SignatureService:    v1SignatureService,
		projectsClaGroupsRepo: pcgRepo,
		s3:                    s3.New(awsSession),
		signaturesBucket:      signaturesBucketName,
	}
}

func (s *service) GetProjectCompanySignatures(ctx context.Context, companyID, companySFID, projectSFID string) (*models.Signatures, error) {
	pm, err := s.projectsClaGroupsRepo.GetClaGroupIDForProject(projectSFID)
	if err != nil {
		return nil, err
	}
	signed := true
	approved := true
	sig, err := s.v1SignatureService.GetProjectCompanySignature(ctx, companyID, pm.ClaGroupID, &signed, &approved, nil, aws.Int64(HugePageSize))
	if err != nil {
		return nil, err
	}
	resp := &v1Models.Signatures{
		Signatures: make([]*v1Models.Signature, 0),
	}
	if sig != nil {
		resp.ResultCount = 1
		resp.Signatures = append(resp.Signatures, sig)
	}
	return v2SignaturesReplaceCompanyID(resp, companyID, companySFID)
}

func eclaSigCsvLine(sig *v1Models.CorporateContributor) string {
	var dateTime string
	t, err := utils.ParseDateTime(sig.Timestamp)
	if err != nil {
		log.WithFields(logrus.Fields{"signature_id": sig.LinuxFoundationID, "signature_created": sig.Timestamp}).
			Error("invalid time format present for signatures")
	} else {
		dateTime = t.Format("Jan 2,2006")
	}
	return fmt.Sprintf("\n%s,%s,%s,%s,\"%s\"", sig.GithubID, sig.LinuxFoundationID, sig.Name, sig.Email, dateTime)
}

func (s service) GetClaGroupCorporateContributorsCsv(ctx context.Context, claGroupID string, companyID string) ([]byte, error) {
	var b bytes.Buffer
	result, err := s.v1SignatureService.GetClaGroupCorporateContributors(ctx, claGroupID, &companyID, nil)
	if err != nil {
		return nil, err
	}

	if len(result.List) == 0 {
		return nil, errors.New("not Found")
	}

	b.WriteString(`GitHub ID,LF_ID,Name,Email,Date Signed`)
	for _, sig := range result.List {
		b.WriteString(eclaSigCsvLine(sig))
	}
	return b.Bytes(), nil
}

func (s service) GetProjectIclaSignaturesCsv(ctx context.Context, claGroupID string) ([]byte, error) {
	var b bytes.Buffer
	result, err := s.v1SignatureService.GetClaGroupICLASignatures(ctx, claGroupID, nil)
	if err != nil {
		return nil, err
	}
	b.WriteString(`GitHub ID,LF_ID,Name,Email,Date Signed`)
	for _, sig := range result.List {
		b.WriteString(iclaSigCsvLine(sig))
	}
	return b.Bytes(), nil
}

func (s service) GetProjectCclaSignaturesCsv(ctx context.Context, claGroupID string) ([]byte, error) {
	f := logrus.Fields{
		"functionName":   "GetProjectCclaSignaturesCsv",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"claGroupID":     claGroupID,
	}
	log.WithFields(f).Debug("querying for CCLA signatures...")
	result, err := s.v1SignatureService.GetClaGroupCCLASignatures(ctx, claGroupID)
	if err != nil {
		log.WithFields(f).Warnf("error loading CCLA signatures for CLA group, error: %+v", err)
		return nil, err
	}
	log.WithFields(f).Debugf("loaded %d CCLA signatures", len(result.Signatures))

	var b bytes.Buffer
	log.WithFields(f).Debug("writing CCLA signatures header CSV...")
	b.WriteString(cclaSigCsvHeader())
	log.WithFields(f).Debugf("writing CCLA %d signatures records as CSV...", len(result.Signatures))
	for _, sig := range result.Signatures {
		b.WriteString(cclaSigCsvLine(sig))
	}
	return b.Bytes(), nil
}

func (s service) GetProjectIclaSignatures(ctx context.Context, claGroupID string, searchTerm *string) (*models.IclaSignatures, error) {
	f := logrus.Fields{
		"functionName":   "v2.signatures.service.GetProjectIclaSignatures",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"claGroupID":     claGroupID,
		"searchTerm":     utils.StringValue(searchTerm),
	}

	var out models.IclaSignatures
	result, err := s.v1SignatureService.GetClaGroupICLASignatures(ctx, claGroupID, searchTerm)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to load ICLA signatures using the specified search parameters")
		return nil, err
	}

	err = copier.Copy(&out, result)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to convert signature results from v1 to v2")
		return nil, err
	}

	return &out, nil
}

func (s service) GetSignedDocument(ctx context.Context, signatureID string) (*models.SignedDocument, error) {
	sig, err := s.v1SignatureService.GetSignature(ctx, signatureID)
	if err != nil {
		return nil, err
	}
	if sig.SignatureType == ClaSignatureType && sig.CompanyName != "" {
		return nil, errors.New("bad request. employee signature does not have signed document")
	}
	var url string
	switch sig.SignatureType {
	case ClaSignatureType:
		url = utils.SignedCLAFilename(sig.ProjectID, "icla", sig.SignatureReferenceID, sig.SignatureID)
	case CclaSignatureType:
		url = utils.SignedCLAFilename(sig.ProjectID, "ccla", sig.SignatureReferenceID, sig.SignatureID)
	}
	signedURL, err := utils.GetDownloadLink(url)
	if err != nil {
		return nil, err
	}
	return &models.SignedDocument{
		SignatureID:  signatureID,
		SignedClaURL: signedURL,
	}, nil
}

func (s service) GetSignedCclaZipPdf(claGroupID string) (*models.URLObject, error) {
	url := utils.SignedClaGroupZipFilename(claGroupID, CCLA)
	ok, err := s.IsZipPresentOnS3(url)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrZipNotPresent
	}
	signedURL, err := utils.GetDownloadLink(url)
	if err != nil {
		return nil, err
	}
	return &models.URLObject{
		URL: signedURL,
	}, nil
}

func (s service) GetSignedIclaZipPdf(claGroupID string) (*models.URLObject, error) {
	url := utils.SignedClaGroupZipFilename(claGroupID, ICLA)
	ok, err := s.IsZipPresentOnS3(url)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrZipNotPresent
	}
	signedURL, err := utils.GetDownloadLink(url)
	if err != nil {
		return nil, err
	}
	return &models.URLObject{
		URL: signedURL,
	}, nil
}

func (s service) IsZipPresentOnS3(zipFilePath string) (bool, error) {
	_, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.signaturesBucket),
		Key:    aws.String(zipFilePath),
	})
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s service) GetClaGroupCorporateContributors(ctx context.Context, claGroupID string, companyID string, searchTerm *string) (*models.CorporateContributorList, error) {
	f := logrus.Fields{
		"functionName": "GetClaGroupCorporateContributors",
		"claGroupID":   claGroupID,
		"companyID":    companyID,
	}
	if searchTerm != nil {
		f["searchTerm"] = *searchTerm
	}

	log.WithFields(f).Debug("querying CLA corporate contributors...")
	result, err := s.v1SignatureService.GetClaGroupCorporateContributors(ctx, claGroupID, &companyID, searchTerm)
	if err != nil {
		return nil, err
	}

	log.WithFields(f).Debug("converting to v2 response model...")
	var resp models.CorporateContributorList
	err = copier.Copy(&resp, result)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
