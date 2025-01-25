// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	// "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/communitybridge/easycla/cla-backend-go/company"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v1/models"
	sigParams "github.com/communitybridge/easycla/cla-backend-go/gen/v1/restapi/operations/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/gerrits"
	"github.com/communitybridge/easycla/cla-backend-go/users"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	cla_group "github.com/communitybridge/easycla/cla-backend-go/project/repository"
	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"
	"github.com/communitybridge/easycla/cla-backend-go/repositories"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/sirupsen/logrus"
)

const (
	batchSize               = 100
	signatureBatchSize      = 100
	maxConcurrentGoroutines = 20 // Control concurrency
)

var (
	awsSession    = session.Must(session.NewSession())
	stage         string
	companyRepo   company.IRepository
	projectRepo   cla_group.ProjectRepository
	signatureRepo signatures.SignatureRepository
)

func init() {
	stage = os.Getenv("STAGE")
	if stage == "" {
		log.Fatal("stage not set")
	}
	log.Infof("STAGE set to %s\n", stage)
	companyRepo = company.NewRepository(awsSession, stage)
	ghRepo := repositories.NewRepository(awsSession, stage)
	gerritRepo := gerrits.NewRepository(awsSession, stage)
	pcgRepo := projects_cla_groups.NewRepository(awsSession, stage)
	projectRepo = cla_group.NewRepository(awsSession, stage, ghRepo, gerritRepo, pcgRepo)
	userRepo := users.NewRepository(awsSession, stage)
	signatureRepo = signatures.NewRepository(awsSession, stage, companyRepo, userRepo, nil, nil, nil, nil, nil)
}

func main() {
	f := logrus.Fields{
		"functionName": "main",
		"stage":        stage,
	}
	log.WithFields(f).Info("loading company data...")

	// Load the company data
	companyData, err := companyRepo.GetCompanies(context.Background())
	if err != nil {
		log.Warnf("Unable to load company data, error: %+v", err)
		return
	}

	if len(companyData.Companies) == 0 {
		log.Warn("No companies found")
		return
	}

	var companyDataList []CompanyData

	log.WithFields(f).Infof("processing %d companies...", len(companyData.Companies))

	processCompaniesBatch(companyData.Companies, &companyDataList)
	err = exportToCSV(companyDataList)
	if err != nil {
		log.Warnf("Unable to export company data to csv, error: %+v", err)
		return
	}

	// // Process the companies in batches
	// for i := 0; i < len(companyData.Companies); i += batchSize {
	// 	end := i + batchSize
	// 	if end > len(companyData.Companies) {
	// 		end = len(companyData.Companies)
	// 	}
	// 	processCompaniesBatch(companyData.Companies[i:end], &companyDataList)
	// 	log.WithFields(f).Info("exporting company data to csv...")
	// 	err = exportToCSV(companyDataList)
	// 	if err != nil {
	// 		log.Warnf("Unable to export company data to csv, error: %+v", err)
	// 		return
	// 	}
	// }

}

func processCompaniesBatch(companiesBatch []models.Company, companyDataList *[]CompanyData) {
	f := logrus.Fields{
		"functionName": "processCompaniesBatch",
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Semaphore channel to control the number of concurrent goroutines
	sem := make(chan struct{}, maxConcurrentGoroutines)

	for _, company := range companiesBatch {
		wg.Add(1)
		go func(companyID string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			log.WithFields(f).Infof("processing company: %s", companyID)
			data, err := processCompany(companyID)
			if err != nil {
				log.Warnf("Unable to process company data, error: %+v", err)
				return
			}
			// log.WithFields(f).Infof("processed company data: %+v", data)
			mu.Lock()
			if data != nil && len(data.ClaManagers) > 0 {
				*companyDataList = append(*companyDataList, *data)
			} else {
				log.Warn("No ccla signatures found for company")
			}
			mu.Unlock()
		}(company.CompanyID)
	}

	wg.Wait()
}

func processCompany(companyID string) (*CompanyData, error) {
	f := logrus.Fields{
		"functionName": "processCompany",
		"companyID":    companyID,
	}

	var companyData CompanyData

	log.WithFields(f).Info("loading company data...")
	companyModel, err := companyRepo.GetCompany(context.Background(), companyID)
	if err != nil {
		log.Warnf("Unable to load company data, error: %+v", err)
		return nil, err
	}

	params := sigParams.GetCompanySignaturesParams{
		CompanyID: companyModel.CompanyID,
	}

	companySignatures, err := signatureRepo.GetCompanySignatures(context.Background(), params, 10, false)
	if err != nil {
		log.Warnf("Unable to load CCLA signatures, error: %+v", err)
		return nil, err
	}

	if len(companySignatures.Signatures) == 0 {
		log.Warn("No CCLA signatures found")
		return nil, nil
	}

	companyData.CompanyID = companyModel.CompanyID
	companyData.CompanyName = companyModel.CompanyName
	companyData.CompanySFID = companyModel.CompanyExternalID


	log.WithFields(f).Info("processing CCLA signatures...")
	populateCompanyDataFromSignatures(&companyData, companySignatures.Signatures)

	return &companyData, nil
}

func populateCompanyDataFromSignatures(companyData *CompanyData, cclaSignatures []*models.Signature) {
	var mu sync.Mutex
	var wg sync.WaitGroup

	claGroupNames := []string{}
	addresses := []string{}
	cclaManagers := []string{}

	for _, sig := range cclaSignatures {
		if companyData.DateFirstSigned == "" || sig.SignedOn < companyData.DateFirstSigned {
			companyData.DateFirstSigned = sig.SignedOn
		}
		if companyData.DateLastSigned == "" || sig.SignedOn > companyData.DateLastSigned {
			companyData.DateLastSigned = sig.SignedOn
		}

		for _, manager := range sig.SignatureACL {
			if !utils.StringInSlice(manager.LfUsername, cclaManagers) {
				cclaManagers = append(cclaManagers, manager.LfUsername)
			}
		}

		// CLA Group Names and Corporation Addresses
		wg.Add(2)

		// CLA Group Name
		go func(claGroupID string) {
			defer wg.Done()
			loadDetails := false
			claGroupModel, err := projectRepo.GetCLAGroupByID(context.Background(), claGroupID, loadDetails)
			if err != nil {
				log.Warnf("unable to load CLA group, error: %+v", err)
				return
			}
			mu.Lock()
			claGroupNames = append(claGroupNames, claGroupModel.ProjectName)
			mu.Unlock()
		}(sig.ProjectID)

		// Corporation Address
		go func(xmlData string) {
			defer wg.Done()
			signature, err := signatureRepo.GetItemSignature(context.Background(), sig.SignatureID)
			if err != nil {
				log.Warnf("unable to load signature, error: %+v", err)
				return
			}

			if signature.UserDocusignRawXML == "" {
				log.Warn("user docusign raw xml is empty")
				return
			}

			log.Info("parsing xml data...")
			companyAddress, err := parseXML(signature.UserDocusignRawXML)
			if err != nil {
				log.Warnf("unable to parse xml data, error: %+v", err)
				return
			}
			mu.Lock()
			addresses = append(addresses, companyAddress)
			mu.Unlock()
		}(sig.SignatureID)
	}

	wg.Wait()
	companyData.ClaGroupNames = claGroupNames
	companyData.CoporationAddress = addresses
	companyData.ClaManagers = cclaManagers
}

func parseXML(xmlData string) (string, error) {
	f := logrus.Fields{
		"functionName": "parseXML",
	}
	var companyAddress string
	// Parse the XML data
	var envelopeInformation DocusignEnvelopeInformation
	log.WithFields(f).Info("unmarshalling xml data...")
	err := xml.Unmarshal([]byte(xmlData), &envelopeInformation)
	if err != nil {
		log.Warnf("unable to unmarshal xml data, error: %+v", err)
		return companyAddress, err
	}

	// Extract the corporation address
	var addressParts []string
	for _, recipientStatus := range envelopeInformation.EnvelopeStatus.RecipientStatuses.RecipientStatus {
		for _, field := range recipientStatus.FormData.XFDF.Fields.Field {
			switch field.Name {
			case "corporation_address1", "corporation_address2", "corporation_address3":
				if field.Value != "" && strings.ToLower(field.Value) != "none" {
					log.WithFields(f).Infof("adding address part: %s", field.Value)
					addressParts = append(addressParts, field.Value)
				}
			}
		}
	}

	companyAddress = strings.Join(addressParts, ", ")
	log.WithFields(f).Infof("company address: %s", companyAddress)
	return companyAddress, nil
}

func exportToCSV(companyData []CompanyData) error {
	f := logrus.Fields{
		"functionName": "exportToCSV",
	}
	log.WithFields(f).Info("exporting company data to csv...")
	// Export the data to CSV
	// create the file with a timestamp
	file, err := os.Create(fmt.Sprintf("company_data_%s.csv", time.Now().Format("2006-01-02T15:04:05")))
	if err != nil {
		log.Warnf("unable to create file, error: %+v", err)

		return err
	}

	// defer file.Close()
	defer func() {
		if err = file.Close(); err != nil {
			log.Warnf("unable to close file, error: %+v", err)
		}
	}()

	w := csv.NewWriter(file)
	defer w.Flush()

	// Write the header
	headers := []string{
		"Company ID",
		"Company Name",
		"Company SFID",
		"CCLA Signatures",
		"Date First Signed",
		"Date Last Signed",
		"Corporation Address",
		"CLA Managers",
		"CLA Group Names",
	}

	err = w.Write(headers)
	if err != nil {
		log.Warnf("unable to write headers, error: %+v", err)
		return err
	}

	// Write the data
	for _, data := range companyData {
		address := " "
		if len(data.CoporationAddress) > 0 {
			address = data.CoporationAddress[0]
		}
		record := []string{
			data.CompanyID,
			data.CompanyName,
			data.CompanySFID,
			fmt.Sprintf("%d", len(data.ClaManagers)),
			data.DateFirstSigned,
			data.DateLastSigned,
			address,
			strings.Join(data.ClaManagers, ", "),
			strings.Join(data.ClaGroupNames, ", "),
		}

		log.WithFields(f).Infof("writing record: %+v", record)
		err = w.Write(record)
		if err != nil {
			log.Warnf("unable to write record, error: %+v", err)
			return err
		}
	}

	log.WithFields(f).Info("exported company data to csv")
	return nil
}
