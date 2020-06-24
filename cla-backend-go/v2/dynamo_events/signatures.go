package dynamo_events

import (
	"errors"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
)

// constants
const (
	CLASignatureType  = "cla"
	CCLASignatureType = "ccla"

	ICLASignatureType = "icla"
	ECLASignatureType = "ecla"
)

// Signature database model
type Signature struct {
	SignatureID                   string   `json:"signature_id"`
	DateCreated                   string   `json:"date_created"`
	DateModified                  string   `json:"date_modified"`
	SignatureApproved             bool     `json:"signature_approved"`
	SignatureSigned               bool     `json:"signature_signed"`
	SignatureDocumentMajorVersion string   `json:"signature_document_major_version"`
	SignatureDocumentMinorVersion string   `json:"signature_document_minor_version"`
	SignatureReferenceID          string   `json:"signature_reference_id"`
	SignatureReferenceName        string   `json:"signature_reference_name"`
	SignatureReferenceNameLower   string   `json:"signature_reference_name_lower"`
	SignatureProjectID            string   `json:"signature_project_id"`
	SignatureReferenceType        string   `json:"signature_reference_type"`
	SignatureType                 string   `json:"signature_type"`
	SignatureUserCompanyID        string   `json:"signature_user_ccla_company_id"`
	EmailWhitelist                []string `json:"email_whitelist"`
	DomainWhitelist               []string `json:"domain_whitelist"`
	GitHubWhitelist               []string `json:"github_whitelist"`
	GitHubOrgWhitelist            []string `json:"github_org_whitelist"`
	SignatureACL                  []string `json:"signature_acl"`
	SigtypeSignedApprovedID       string   `json:"sigtype_signed_approved_id"`
	UserGithubUsername            string   `json:"user_github_username"`
	UserLFUsername                string   `json:"user_lf_username"`
	UserName                      string   `json:"user_name"`
	UserEmail                     string   `json:"user_email"`
	SignedOn                      string   `json:"signed_on"`
}

// should be called when we modify signature
func (s *service) SignatureSignedEvent(event events.DynamoDBEventRecord) error {
	var newSignature, oldSignature Signature
	err := unmarshalStreamImage(event.Change.OldImage, &oldSignature)
	if err != nil {
		return err
	}
	err = unmarshalStreamImage(event.Change.NewImage, &newSignature)
	if err != nil {
		return err
	}
	// check if signature signed event is received
	if !oldSignature.SignatureSigned && newSignature.SignatureSigned {
		err = s.signatureRepo.AddSignedOn(newSignature.SignatureID)
		if err != nil {
			log.WithField("signature_id", newSignature.SignatureID).Warnf("failed to add signed_on on signature")
		}
		if newSignature.SignatureType == CCLASignatureType {
			err = s.SetInitialCLAManagerACSPermissions(newSignature.SignatureID)
			if err != nil {
				log.WithField("signature_id", newSignature.SignatureID).Warnf("failed to set initial cla manager")
			}
		}
	}
	return nil
}

// SignatureAdded function should be called when new icla, ecla signature added
func (s *service) SignatureAddSigTypeSignedApprovedID(event events.DynamoDBEventRecord) error {
	var newSig Signature
	var sigType string
	var id string
	err := unmarshalStreamImage(event.Change.NewImage, &newSig)
	if err != nil {
		return err
	}
	if newSig.SigtypeSignedApprovedID != "" {
		return nil
	}
	log.Debugf("setting sigtype_signed_approved_id for signature: %s", newSig.SignatureID)
	switch {
	case newSig.SignatureType == CCLASignatureType:
		sigType = CCLASignatureType
		id = newSig.SignatureReferenceID
	case newSig.SignatureType == CLASignatureType && newSig.SignatureUserCompanyID == "":
		sigType = ICLASignatureType
		id = newSig.SignatureReferenceID
	case newSig.SignatureType == CLASignatureType && newSig.SignatureUserCompanyID != "":
		sigType = ECLASignatureType
		id = newSig.SignatureUserCompanyID
	default:
		log.Warnf("setting sigtype_signed_approved_id for signature: %s failed", newSig.SignatureID)
		return errors.New("invalid signature in SignatureAddSigTypeSignedApprovedID")
	}
	err = s.signatureRepo.AddSigTypeSignedApprovedID(newSig.SignatureID, sigType, newSig.SignatureSigned, newSig.SignatureApproved, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) SignatureAddUsersDetails(event events.DynamoDBEventRecord) error {
	var newSig Signature
	err := unmarshalStreamImage(event.Change.NewImage, &newSig)
	if err != nil {
		return err
	}
	if newSig.SignatureReferenceType == "user" && newSig.UserLFUsername == "" && newSig.UserGithubUsername == "" {
		log.Debugf("adding users details in signature: %s", newSig.SignatureID)
		err = s.signatureRepo.AddUsersDetails(newSig.SignatureID, newSig.SignatureReferenceID)
		if err != nil {
			log.Debugf("adding users details in signature: %s failed. error = %s", newSig.SignatureID, err.Error())
		}
	}
	return nil
}
