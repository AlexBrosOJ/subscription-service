package service

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"subscription-service/internal/models"
	"subscription-service/internal/repository"
)

type SubscriptionService struct {
	repo   *repository.SubscriptionRepository
	logger *logrus.Logger
}

func NewSubscriptionService(repo *repository.SubscriptionRepository, logger *logrus.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SubscriptionService) Create(req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, expected MM-YYYY: %w", err)
	}

	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsedEnd, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format, expected MM-YYYY: %w", err)
		}
		endDate = &parsedEnd

		if endDate.Before(startDate) {
			return nil, fmt.Errorf("end_date cannot be before start_date")
		}
	}

	sub := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	err = s.repo.Create(sub)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) GetByID(id string) (*models.Subscription, error) {
	return s.repo.GetByID(id)
}

func (s *SubscriptionService) Update(id string, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	return s.repo.Update(id, req)
}

func (s *SubscriptionService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *SubscriptionService) List(limit, offset int) ([]models.Subscription, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(limit, offset)
}

func (s *SubscriptionService) GetTotalPrice(req *models.SummaryRequest) (int, error) {
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return 0, fmt.Errorf("invalid start_date format: %w", err)
	}

	endDate, err := time.Parse("01-2006", req.EndDate)
	if err != nil {
		return 0, fmt.Errorf("invalid end_date format: %w", err)
	}

	if endDate.Before(startDate) {
		return 0, fmt.Errorf("end_date cannot be before start_date")
	}

	return s.repo.GetTotalPrice(req.UserID, req.ServiceName, startDate, endDate)
}
