package company

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/LF-Engineering/lfx-kit/auth"
	"github.com/communitybridge/easycla/cla-backend-go/company"

	"github.com/communitybridge/easycla/cla-backend-go/users"
	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/logging"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/aws/aws-sdk-go/aws"
	v1Models "github.com/communitybridge/easycla/cla-backend-go/gen/models"
	v1ProjectParams "github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/project"
	v1SignatureParams "github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	v2ProjectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
	v2ProjectServiceModels "github.com/communitybridge/easycla/cla-backend-go/v2/project-service/models"
	v2UserService "github.com/communitybridge/easycla/cla-backend-go/v2/user-service"
	v2UserServiceModels "github.com/communitybridge/easycla/cla-backend-go/v2/user-service/models"
)

// errors
var (
	ErrProjectNotFound = errors.New("project not found")
)

// constants
const (
	// used when we want to query all data from dependent service.
	HugePageSize        = int64(10000)
	LoadRepoDetails     = true
	DontLoadRepoDetails = false
)

// Service functions for company
type Service interface {
	GetCompanyCLAManagers(companyID string) (*models.CompanyClaManagers, error)
	GetCompanyActiveCLAs(companyID string) (*models.ActiveClaList, error)
	GetCompanyProjectContributors(projectSFID string, companySFID string, searchTerm string) (*models.CorporateContributorList, error)
	GetCompanyProjectCLA(authUser *auth.User, companySFID, projectSFID string) (*models.CompanyProjectClaList, error)
}

// ProjectRepo contains project repo methods
type ProjectRepo interface {
	GetProjectByID(projectID string) (*v1Models.Project, error)
	GetProjectsByExternalID(params *v1ProjectParams.GetProjectsByExternalIDParams, loadRepoDetails bool) (*v1Models.Projects, error)
}

type service struct {
	signatureRepo signatures.SignatureRepository
	projectRepo   ProjectRepo
	userRepo      users.UserRepository
	companyRepo   company.IRepository
}

type signatureResponse struct {
	companyID  string
	projectID  string
	signatures *v1Models.Signatures
	err        error
}

// NewService returns instance of company service
func NewService(sigRepo signatures.SignatureRepository, projectRepo ProjectRepo, usersRepo users.UserRepository, companyRepo company.IRepository) Service {
	return &service{
		signatureRepo: sigRepo,
		projectRepo:   projectRepo,
		userRepo:      usersRepo,
		companyRepo:   companyRepo,
	}
}

func signedCLAFilename(projectID string, claType string, identifier string, signatureID string) string {
	return strings.Join([]string{"contract-group", projectID, claType, identifier, signatureID}, "/") + ".pdf"
}

func (s *service) getAllCCLASignatures(companyID string) ([]*v1Models.Signature, error) {
	var sigs []*v1Models.Signature
	var lastScannedKey *string
	for {
		signatures, err := s.signatureRepo.GetCompanySignatures(v1SignatureParams.GetCompanySignaturesParams{
			CompanyID:     companyID,
			SignatureType: aws.String("ccla"),
			NextKey:       lastScannedKey,
		}, 1000, signatures.DontLoadACLDetails)
		if err != nil {
			return nil, err
		}
		sigs = append(sigs, signatures.Signatures...)
		if signatures.LastKeyScanned == "" {
			break
		}
		lastScannedKey = aws.String(signatures.LastKeyScanned)
	}
	return sigs, nil
}

// return list of all signature of the company for the projects
func (s *service) getCompanyProjectCCLASignatures(companyID string, projects *v1Models.Projects) ([]*v1Models.Signature, error) {
	var sigs []*v1Models.Signature
	res := make(chan *signatureResponse)
	var wg sync.WaitGroup
	wg.Add(len(projects.Projects))
	go func() {
		wg.Wait()
		close(res)
	}()
	for _, project := range projects.Projects {
		go func(companyID, projectID string, responseChan chan *signatureResponse) {
			defer wg.Done()
			signed, approved := true, true
			pageSize := HugePageSize
			sigs, err := s.signatureRepo.GetProjectCompanySignatures(companyID, projectID, &signed, &approved, nil, &pageSize)
			if err != nil {
				return
			}
			responseChan <- &signatureResponse{
				companyID:  companyID,
				projectID:  projectID,
				signatures: sigs,
				err:        err,
			}
		}(companyID, project.ProjectID, res)
	}
	var sigErr error
	for sigResp := range res {
		if sigResp.err != nil {
			log.WithFields(logrus.Fields{
				"project_id": sigResp.projectID,
				"company_id": sigResp.companyID,
			}).Error("unable to fetch ccla signatures for project")
			sigErr = sigResp.err
			continue
		}
		sigs = append(sigs, sigResp.signatures.Signatures...)
	}
	if sigErr != nil {
		return nil, sigErr
	}
	return sigs, nil
}

func (s *service) GetCompanyCLAManagers(companyID string) (*models.CompanyClaManagers, error) {
	sigs, err := s.getAllCCLASignatures(companyID)
	if err != nil {
		return nil, err
	}
	var claManagers []*models.CompanyClaManager
	lfUsernames := utils.NewStringSet()
	projectIDs := utils.NewStringSet()
	// Get CLA managers
	for _, sig := range sigs {
		for _, user := range sig.SignatureACL {
			claManagers = append(claManagers, &models.CompanyClaManager{
				// DB doesn't have approved_on value
				ApprovedOn: sig.SignatureCreated,
				LfUsername: user.LfUsername,
				ProjectID:  sig.ProjectID,
			})
			lfUsernames.Add(user.LfUsername)
			projectIDs.Add(sig.ProjectID)
		}
	}
	// get userinfo and project info
	var usermap map[string]*v2UserServiceModels.User
	var projects map[string]*v1Models.Project
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		usermap, err = getUsersInfo(lfUsernames.List())
	}()
	go func() {
		defer wg.Done()
		projects = s.getProjects(projectIDs.List())
	}()
	wg.Wait()
	if err != nil {
		return nil, err
	}
	// fill user info
	fillUsersInfo(claManagers, usermap)
	// fill project info
	fillProjectInfo(claManagers, projects)
	// sort result by cla manager name
	sort.Slice(claManagers, func(i, j int) bool {
		return claManagers[i].Name < claManagers[j].Name
	})
	return &models.CompanyClaManagers{List: claManagers}, nil
}

func getUsersInfo(lfUsernames []string) (map[string]*v2UserServiceModels.User, error) {
	userServiceClient := v2UserService.GetClient()
	users, err := userServiceClient.GetUsersByUsernames(lfUsernames)
	if err != nil {
		return nil, err
	}
	usermap := make(map[string]*v2UserServiceModels.User)
	for _, user := range users {
		usermap[user.Username] = user
	}
	return usermap, nil
}

func fillUsersInfo(claManagers []*models.CompanyClaManager, usermap map[string]*v2UserServiceModels.User) {
	for _, cm := range claManagers {
		user, ok := usermap[cm.LfUsername]
		if !ok {
			logging.Warnf("Unable to get user with username %s", cm.LfUsername)
			continue
		}
		cm.Name = user.Name
		cm.LogoURL = user.LogoURL
		cm.UserSfid = user.ID
		if user.Email != nil {
			cm.Email = *user.Email
		} else {
			if len(user.Emails) > 0 {
				cm.Email = utils.StringValue(user.Emails[0].EmailAddress)
			}
		}
	}
}

func (s *service) getProjects(projectIDs []string) map[string]*v1Models.Project {
	projects := make(map[string]*v1Models.Project)
	prChan := make(chan *v1Models.Project)
	for _, id := range projectIDs {
		go func(projectID string) {
			project, err := s.projectRepo.GetProjectByID(projectID)
			if err != nil {
				log.Warnf("Unable to fetch project details for project %s. error = %s", projectID, err)
			}
			prChan <- project
		}(id)
	}
	for range projectIDs {
		project := <-prChan
		if project != nil {
			projects[project.ProjectID] = project
		}
	}
	return projects
}

func fillProjectInfo(claManagers []*models.CompanyClaManager, projects map[string]*v1Models.Project) {
	projectSFIDs := utils.NewStringSet()
	for _, project := range projects {
		projectSFIDs.Add(project.ProjectExternalID)
	}
	pmap := getSFProjectDetails(projectSFIDs.List())
	for _, claManager := range claManagers {
		project, ok := projects[claManager.ProjectID]
		if !ok {
			continue
		}
		claManager.ClaGroupName = project.ProjectName
		claManager.ProjectSfid = project.ProjectExternalID
		if sfproject, ok := pmap[project.ProjectExternalID]; ok {
			claManager.ProjectName = sfproject.Name
		}
	}
}

func (s *service) GetCompanyActiveCLAs(companyID string) (*models.ActiveClaList, error) {
	var out models.ActiveClaList
	sigs, err := s.getAllCCLASignatures(companyID)
	if err != nil {
		return nil, err
	}
	out.List = make([]*models.ActiveCla, 0, len(sigs))
	if len(sigs) == 0 {
		return &out, nil
	}
	var wg sync.WaitGroup
	wg.Add(len(sigs))
	for _, sig := range sigs {
		activeCla := &models.ActiveCla{}
		out.List = append(out.List, activeCla)
		go func(swg *sync.WaitGroup, signature *v1Models.Signature, acla *models.ActiveCla) {
			s.fillActiveCLA(swg, signature, acla)
		}(&wg, sig, activeCla)
	}
	wg.Wait()
	return &out, nil
}

func (s *service) fillActiveCLA(wg *sync.WaitGroup, sig *v1Models.Signature, activeCla *models.ActiveCla) {
	defer wg.Done()
	p, err := s.projectRepo.GetProjectByID(sig.ProjectID)
	if err != nil {
		log.Error("fillActiveCLA : unable to get project", err)
		return
	}
	psc := v2ProjectService.GetClient()
	projectDetails, err := psc.GetProject(p.ProjectExternalID)
	if err != nil {
		log.Error("fillActiveCLA : unable to get project details", err)
		return
	}

	// fill details from dynamodb
	activeCla.ProjectID = sig.ProjectID
	activeCla.SignedOn = sig.SignatureCreated
	activeCla.ClaGroupName = p.ProjectName

	// fill details from project service
	activeCla.ProjectName = projectDetails.Name
	activeCla.ProjectSfid = p.ProjectExternalID
	activeCla.ProjectType = projectDetails.ProjectType
	activeCla.ProjectLogo = projectDetails.ProjectLogo
	var signatoryName string
	var cwg sync.WaitGroup
	cwg.Add(2)

	var cclaURL string
	go func() {
		defer cwg.Done()
		cclaURL, err = utils.GetDownloadLink(signedCLAFilename(sig.ProjectID, sig.SignatureType, sig.SignatureReferenceID, sig.SignatureID))
		if err != nil {
			log.Error("fillActiveCLA : unable to get ccla s3 link", err)
			return
		}
	}()

	go func() {
		defer cwg.Done()
		usc := v2UserService.GetClient()
		if len(sig.SignatureACL) == 0 {
			log.Warnf("signature : %s have empty signature_acl", sig.SignatureID)
			return
		}
		lfUsername := sig.SignatureACL[0].LfUsername
		user, err := usc.GetUserByUsername(lfUsername)
		if err != nil {
			log.Warnf("unable to get user with lf username : %s", lfUsername)
			return
		}
		signatoryName = user.Name
	}()

	cwg.Wait()

	activeCla.SignatoryName = signatoryName
	activeCla.CclaURL = cclaURL
}

// return projects output for which cla_group is present in cla
func (s *service) filterClaProjects(projects []*v2ProjectServiceModels.ProjectOutput) []*v2ProjectServiceModels.ProjectOutput { //nolint
	results := make([]*v2ProjectServiceModels.ProjectOutput, 0)
	prChan := make(chan *v2ProjectServiceModels.ProjectOutput)
	for _, v := range projects {
		go func(projectOutput *v2ProjectServiceModels.ProjectOutput) {
			project, err := s.projectRepo.GetProjectsByExternalID(&v1ProjectParams.GetProjectsByExternalIDParams{
				ProjectSFID: projectOutput.ID,
				PageSize:    aws.Int64(1),
			}, DontLoadRepoDetails)
			if err != nil {
				log.Warnf("Unable to fetch project details for project with external id %s. error = %s", projectOutput.ID, err)
				prChan <- nil
				return
			}
			if project.ResultCount == 0 {
				prChan <- nil
				return
			}
			prChan <- projectOutput
		}(v)
	}
	for range projects {
		project := <-prChan
		if project != nil {
			results = append(results, project)
		}
	}
	return results
}

func (s *service) GetCompanyProjectContributors(projectSFID string, companySFID string, searchTerm string) (*models.CorporateContributorList, error) {
	list := make([]*models.CorporateContributor, 0)
	sigs, err := s.getAllCompanyProjectEmployeeSignatures(companySFID, projectSFID)
	if err != nil {
		return nil, err
	}
	if len(sigs) == 0 {
		return &models.CorporateContributorList{
			List: list,
		}, nil
	}
	var wg sync.WaitGroup
	result := make(chan *models.CorporateContributor)
	wg.Add(len(sigs))
	go func() {
		wg.Wait()
		close(result)
	}()

	for _, sig := range sigs {
		go fillCorporateContributorModel(&wg, s.userRepo, sig, result, searchTerm)
	}

	for corpContributor := range result {
		list = append(list, corpContributor)
	}

	return &models.CorporateContributorList{
		List: list,
	}, nil
}

func fillCorporateContributorModel(wg *sync.WaitGroup, usersRepo users.UserRepository, sig *v1Models.Signature, result chan *models.CorporateContributor, searchTerm string) {
	defer wg.Done()
	user, err := usersRepo.GetUser(sig.SignatureReferenceID)
	if err != nil {
		log.Error("fillCorporateContributorModel: unable to get user info", err)
		return
	}
	if searchTerm != "" {
		ls := strings.ToLower(searchTerm)
		if !(strings.Contains(strings.ToLower(user.Username), ls) || strings.Contains(strings.ToLower(user.LfUsername), ls)) {
			return
		}
	}
	var contributor models.CorporateContributor
	var sigSignedTime = sig.SignatureCreated
	contributor.GithubID = user.GithubID
	contributor.LinuxFoundationID = user.LfUsername
	contributor.Name = user.Username
	t, err := utils.ParseDateTime(sig.SignatureCreated)
	if err != nil {
		log.Error("fillCorporateContributorModel: unable to parse time", err)
	} else {
		sigSignedTime = utils.TimeToString(t)
	}
	contributor.Timestamp = sigSignedTime
	contributor.SignatureVersion = fmt.Sprintf("v%s.%s", sig.SignatureMajorVersion, sig.SignatureMinorVersion)

	// send contributor struct on result channel
	result <- &contributor
}

func (s *service) getAllCompanyProjectEmployeeSignatures(companySFID string, projectSFID string) ([]*v1Models.Signature, error) {
	comp, projects, err := s.getCompanyAndProjects(companySFID, projectSFID)
	if err != nil {
		return nil, err
	}
	if len(projects.Projects) == 0 {
		return nil, nil
	}
	companyID := comp.CompanyID
	resp := make(chan *signatureResponse)
	var swg sync.WaitGroup
	swg.Add(len(projects.Projects))
	go func() {
		swg.Wait()
		close(resp)
	}()
	for _, project := range projects.Projects {
		go getCompanyProjectEmployeeSignatures(&swg, s.signatureRepo, companyID, project.ProjectID, resp)
	}
	var sigs []*v1Models.Signature
	for res := range resp {
		if res.err != nil {
			log.WithFields(logrus.Fields{
				"company_id": res.companyID,
				"project_id": res.projectID,
			}).Error("unable to get company project signatures", res.err)
			continue
		}
		sigs = append(sigs, res.signatures.Signatures...)
	}
	return sigs, nil
}

func getCompanyProjectEmployeeSignatures(wg *sync.WaitGroup, signatureRepo signatures.SignatureRepository, companyID, projectID string, resp chan<- *signatureResponse) {
	defer wg.Done()
	params := v1SignatureParams.GetProjectCompanyEmployeeSignaturesParams{
		HTTPRequest: nil,
		CompanyID:   companyID,
		ProjectID:   projectID,
	}
	sigs, err := signatureRepo.GetProjectCompanyEmployeeSignatures(params, HugePageSize)
	resp <- &signatureResponse{
		companyID:  companyID,
		projectID:  projectID,
		signatures: sigs,
		err:        err,
	}
}

func (s *service) GetCompanyProjectCLA(authUser *auth.User, companySFID, projectSFID string) (*models.CompanyProjectClaList, error) {
	var canSign bool
	resources := authUser.ResourceIDsByTypeAndRole(auth.ProjectOrganization, "cla-manager-designee")
	projectOrg := fmt.Sprintf("%s|%s", projectSFID, companySFID)
	for _, r := range resources {
		if r == projectOrg {
			canSign = true
			break
		}
	}
	// get company and projects
	companyModel, projects, err := s.getCompanyAndProjects(companySFID, projectSFID)
	if err != nil {
		return nil, err
	}
	if len(projects.Projects) == 0 {
		return nil, errors.New("project not found")
	}
	// get company project signatures
	sigs, err := s.getCompanyProjectCCLASignatures(companyModel.CompanyID, projects)
	if err != nil {
		return nil, err
	}
	var projectName string

	resp := &models.CompanyProjectClaList{
		SignedClaList:       make([]*models.ActiveCla, 0),
		UnsignedProjectList: make([]*models.UnsignedProject, 0),
	}
	// pmap will keep track of unsigned project
	pmap := make(map[string]v1Models.Project)
	for _, project := range projects.Projects {
		pmap[project.ProjectID] = project
	}
	// fill details for signed cla
	var wg sync.WaitGroup
	wg.Add(len(sigs))
	wg.Add(1)
	go func() {
		defer wg.Done()
		psc := v2ProjectService.GetClient()
		projectDetails, err := psc.GetProject(projectSFID)
		if err != nil {
			log.Error("GetCompanyProjectCLAStatus : unable to get project details", err)
		} else {
			projectName = projectDetails.Name
		}
	}()
	for _, sig := range sigs {
		activeCla := &models.ActiveCla{}
		// delete it from unsigned project
		delete(pmap, sig.ProjectID)

		resp.SignedClaList = append(resp.SignedClaList, activeCla)
		go func(swg *sync.WaitGroup, signature *v1Models.Signature, acla *models.ActiveCla) {
			s.fillActiveCLA(swg, signature, acla)
		}(&wg, sig, activeCla)
	}
	wg.Wait()
	// fill details for not signed cla
	for _, project := range pmap {
		unsignedProject := &models.UnsignedProject{
			CanSign:      canSign,
			ClaGroupID:   project.ProjectID,
			ClaGroupName: project.ProjectName,
			ProjectName:  projectName,
			ProjectSfid:  project.ProjectExternalID,
		}
		resp.UnsignedProjectList = append(resp.UnsignedProjectList, unsignedProject)
	}
	return resp, nil
}

// get company and project parallely
func (s *service) getCompanyAndProjects(companySFID, projectSFID string) (*v1Models.Company, *v1Models.Projects, error) {
	var comp *v1Models.Company
	var companyErr, projectErr error
	var projects *v1Models.Projects
	// query projects and company
	var cp sync.WaitGroup
	cp.Add(2)
	go func() {
		defer cp.Done()
		comp, companyErr = s.companyRepo.GetCompanyByExternalID(companySFID)
	}()
	go func() {
		defer cp.Done()
		t := time.Now()
		projects, projectErr = s.projectRepo.GetProjectsByExternalID(&v1ProjectParams.GetProjectsByExternalIDParams{
			ProjectSFID: projectSFID,
			PageSize:    aws.Int64(HugePageSize),
		}, DontLoadRepoDetails)
		log.WithField("time_taken", time.Since(t).String()).Debugf("getting project by external id : %s completed", projectSFID)
	}()
	cp.Wait()
	if companyErr != nil {
		return nil, nil, companyErr
	}
	if projectErr != nil {
		return nil, nil, projectErr
	}
	return comp, projects, nil
}

func getSFProjectDetails(sfProjectIDs []string) map[string]*v2ProjectServiceModels.ProjectOutputDetailed {
	pmap := make(map[string]*v2ProjectServiceModels.ProjectOutputDetailed)
	if len(sfProjectIDs) == 0 {
		return pmap
	}
	psc := v2ProjectService.GetClient()
	type sfProjectOutput struct {
		sfProjectID    string
		projectDetails *v2ProjectServiceModels.ProjectOutputDetailed
		err            error
	}
	responseChan := make(chan *sfProjectOutput)
	var wg sync.WaitGroup
	wg.Add(len(sfProjectIDs))
	go func() {
		wg.Wait()
		close(responseChan)
	}()
	for _, externalProjectID := range sfProjectIDs {
		go func(projectSFID string) {
			defer wg.Done()
			projectDetails, err := psc.GetProject(projectSFID)
			responseChan <- &sfProjectOutput{
				sfProjectID:    projectSFID,
				projectDetails: projectDetails,
				err:            err,
			}
		}(externalProjectID)
	}
	for resp := range responseChan {
		if resp.err != nil {
			log.WithField("project_sfid", resp.sfProjectID).Error("unable to get salesforce project details", resp.err)
			continue
		}
		pmap[resp.sfProjectID] = resp.projectDetails
	}
	return pmap
}
