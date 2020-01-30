package metrics

import (
	"sync"

	"github.com/labstack/gommon/log"

	"github.com/communitybridge/easycla/cla-backend-go/company"
	"github.com/communitybridge/easycla/cla-backend-go/gen/models"
	"github.com/communitybridge/easycla/cla-backend-go/gen/restapi/operations/metrics"
	"github.com/communitybridge/easycla/cla-backend-go/project"
	"github.com/communitybridge/easycla/cla-backend-go/repositories"
	"github.com/communitybridge/easycla/cla-backend-go/signatures"
	"github.com/communitybridge/easycla/cla-backend-go/users"
)

// Service interface defines function of Metrics service
type Service interface {
	GetMetrics(params metrics.GetMetricsParams) (*models.Metrics, error)
}

type service struct {
	userRepo         users.Repository
	companyRepo      company.RepositoryService
	repositoriesRepo repositories.Repository
	signatureRepo    signatures.SignatureRepository
	projectRepo      project.Repository
}

// NewService creates new instance of metrics service
func NewService(
	userRepo users.Repository,
	companyRepo company.RepositoryService,
	repositoriesRepo repositories.Repository,
	signatureRepo signatures.SignatureRepository,
	projectRepo project.Repository,
) Service {
	return &service{
		userRepo:         userRepo,
		companyRepo:      companyRepo,
		repositoriesRepo: repositoriesRepo,
		signatureRepo:    signatureRepo,
		projectRepo:      projectRepo,
	}
}

func (s *service) GetMetrics(params metrics.GetMetricsParams) (*models.Metrics, error) {
	var out models.Metrics
	var wg sync.WaitGroup
	var userMetrics *models.UserMetrics
	var signatureMetrics *models.SignatureMetrics
	var companyMetrics *models.CompaniesMetrics
	var repositoriesMetrics *models.RepositoryMetrics
	var projectMetrics *models.ProjectMetrics
	wg.Add(5)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var err error
		userMetrics, err = s.userRepo.GetMetrics()
		if err != nil {
			log.Warnf("Unable to get user metrics. error = %v", err)
			return
		}
	}(&wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var err error
		signatureMetrics, err = s.signatureRepo.GetMetrics()
		if err != nil {
			log.Warnf("Unable to get signature metrics. error = %v", err)
			return
		}
	}(&wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var err error
		companyMetrics, err = s.companyRepo.GetMetrics()
		if err != nil {
			log.Warnf("Unable to get company metrics. error = %v", err)
			return
		}

	}(&wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var err error
		repositoriesMetrics, err = s.repositoriesRepo.GetMetrics()
		if err != nil {
			log.Warnf("Unable to get repository metrics. error = %v", err)
			return
		}
	}(&wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		var err error
		projectMetrics, err = s.projectRepo.GetMetrics()
		if err != nil {
			log.Warnf("Unable to get project metrics. error = %v", err)
			return
		}
	}(&wg)

	wg.Wait()

	if userMetrics != nil {
		out.Users = *userMetrics
	}
	if signatureMetrics != nil {
		out.Signatures = *signatureMetrics
	}
	if companyMetrics != nil {
		out.Companies = *companyMetrics
	}
	if repositoriesMetrics != nil {
		out.Repositories = *repositoriesMetrics
	}
	if projectMetrics != nil {
		out.Projects = *projectMetrics
	}
	return &out, nil
}
