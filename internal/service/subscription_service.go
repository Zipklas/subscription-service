package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Zipklas/subscription-service/internal/logger"
	"github.com/Zipklas/subscription-service/internal/model"
	"github.com/Zipklas/subscription-service/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, req model.CreateSubscriptionRequest) (*model.Subscription, error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, req model.UpdateSubscriptionRequest) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error)
	CalculateTotalCost(ctx context.Context, filter model.SummaryFilter) (*model.SummaryResponse, error)
}

type subscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *logger.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *logger.Logger) SubscriptionService {
	return &subscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, req model.CreateSubscriptionRequest) (*model.Subscription, error) {
	s.logger.Info(ctx, "Creating subscription",
		"user_id", req.UserID,
		"service_name", req.ServiceName,
		"monthly_cost", req.MonthlyCost,
	)

	// Парсим даты из строк в формате "01-2006" (месяц-год)
	startDate, err := model.ParseMonthYear(req.StartDate)
	if err != nil {
		s.logger.Error(ctx, "Invalid start date format",
			"start_date", req.StartDate,
			"error", err,
		)
		return nil, fmt.Errorf("invalid start date format, expected MM-YYYY: %w", err)
	}

	endDate, err := model.ParseMonthYearPtr(req.EndDate)
	if err != nil {
		s.logger.Error(ctx, "Invalid end date format",
			"end_date", req.EndDate,
			"error", err,
		)
		return nil, fmt.Errorf("invalid end date format, expected MM-YYYY: %w", err)
	}

	// Валидация дат
	if err := validateDates(startDate, endDate); err != nil {
		s.logger.Error(ctx, "Date validation failed",
			"start_date", startDate,
			"end_date", endDate,
			"error", err,
		)
		return nil, err
	}

	subscription := &model.Subscription{
		ServiceName: req.ServiceName,
		MonthlyCost: req.MonthlyCost,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := s.repo.Create(ctx, subscription); err != nil {
		s.logger.Error(ctx, "Failed to create subscription in repository",
			"user_id", req.UserID,
			"service_name", req.ServiceName,
			"error", err,
		)
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	s.logger.Info(ctx, "Subscription created successfully",
		"subscription_id", subscription.ID,
		"user_id", req.UserID,
	)

	return subscription, nil
}

func (s *subscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req model.UpdateSubscriptionRequest) error {
	s.logger.Info(ctx, "Updating subscription", "subscription_id", id)

	// Парсим даты из строк в формате "01-2006" (месяц-год)
	startDate, err := model.ParseMonthYear(req.StartDate)
	if err != nil {
		s.logger.Error(ctx, "Invalid start date format",
			"start_date", req.StartDate,
			"error", err,
		)
		return fmt.Errorf("invalid start date format, expected MM-YYYY: %w", err)
	}

	endDate, err := model.ParseMonthYearPtr(req.EndDate)
	if err != nil {
		s.logger.Error(ctx, "Invalid end date format",
			"end_date", req.EndDate,
			"error", err,
		)
		return fmt.Errorf("invalid end date format, expected MM-YYYY: %w", err)
	}

	// Валидация дат
	if err := validateDates(startDate, endDate); err != nil {
		s.logger.Error(ctx, "Date validation failed",
			"start_date", startDate,
			"end_date", endDate,
			"error", err,
		)
		return err
	}

	// Проверяем существование подписки
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to check subscription existence",
			"subscription_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to check subscription: %w", err)
	}
	if existing == nil {
		s.logger.Warn(ctx, "Subscription not found for update", "subscription_id", id)
		return fmt.Errorf("subscription not found")
	}

	subscription := &model.Subscription{
		ServiceName: req.ServiceName,
		MonthlyCost: req.MonthlyCost,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := s.repo.Update(ctx, id, subscription); err != nil {
		s.logger.Error(ctx, "Failed to update subscription in repository",
			"subscription_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logger.Info(ctx, "Subscription updated successfully", "subscription_id", id)
	return nil
}

func (s *subscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	s.logger.Debug(ctx, "Getting subscription", "subscription_id", id)

	subscription, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to get subscription from repository",
			"subscription_id", id,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		s.logger.Warn(ctx, "Subscription not found", "subscription_id", id)
		return nil, fmt.Errorf("subscription not found")
	}

	s.logger.Debug(ctx, "Subscription retrieved successfully",
		"subscription_id", id,
		"service_name", subscription.ServiceName,
	)

	return subscription, nil
}

func (s *subscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	s.logger.Info(ctx, "Deleting subscription", "subscription_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to delete subscription from repository",
			"subscription_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	s.logger.Info(ctx, "Subscription deleted successfully", "subscription_id", id)
	return nil
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	s.logger.Debug(ctx, "Listing subscriptions",
		"user_id", userID,
		"service_name", serviceName,
	)

	subscriptions, err := s.repo.List(ctx, userID, serviceName)
	if err != nil {
		s.logger.Error(ctx, "Failed to list subscriptions from repository",
			"user_id", userID,
			"service_name", serviceName,
			"error", err,
		)
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	s.logger.Debug(ctx, "Subscriptions listed successfully",
		"count", len(subscriptions),
		"user_id", userID,
	)

	return subscriptions, nil
}

func (s *subscriptionService) CalculateTotalCost(ctx context.Context, filter model.SummaryFilter) (*model.SummaryResponse, error) {
	s.logger.Info(ctx, "Calculating total cost",
		"start_period", filter.StartPeriod,
		"end_period", filter.EndPeriod,
		"user_id", filter.UserID,
		"service_name", filter.ServiceName,
	)

	total, err := s.repo.CalculateTotalCost(ctx, filter)
	if err != nil {
		s.logger.Error(ctx, "Failed to calculate total cost",
			"start_period", filter.StartPeriod,
			"end_period", filter.EndPeriod,
			"error", err,
		)
		return nil, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	s.logger.Info(ctx, "Total cost calculated successfully",
		"total_cost", total,
		"start_period", filter.StartPeriod,
		"end_period", filter.EndPeriod,
	)

	return &model.SummaryResponse{TotalCost: total}, nil
}

func validateDates(startDate time.Time, endDate *time.Time) error {
	if startDate.IsZero() {
		return fmt.Errorf("start date is required")
	}

	if endDate != nil && !endDate.IsZero() {
		if endDate.Before(startDate) {
			return fmt.Errorf("end date cannot be before start date")
		}
	}

	return nil
}
