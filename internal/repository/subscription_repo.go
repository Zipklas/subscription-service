package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Zipklas/subscription-service/internal/logger"
	"github.com/Zipklas/subscription-service/internal/model"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error)
	CalculateTotalCost(ctx context.Context, filter model.SummaryFilter) (int, error)
}

type subscriptionRepo struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewSubscriptionRepository(db *sql.DB, logger *logger.Logger) SubscriptionRepository {
	return &subscriptionRepo{
		db:     db,
		logger: logger,
	}
}

func (r *subscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, monthly_cost, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	r.logger.Debug(ctx, "Creating subscription in database",
		"service_name", sub.ServiceName,
		"user_id", sub.UserID,
		"monthly_cost", sub.MonthlyCost,
	)

	err := r.db.QueryRowContext(ctx, query,
		sub.ServiceName,
		sub.MonthlyCost,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		r.logger.Error(ctx, "Failed to create subscription in database",
			"service_name", sub.ServiceName,
			"user_id", sub.UserID,
			"error", err,
		)
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	r.logger.Info(ctx, "Subscription created successfully",
		"subscription_id", sub.ID,
		"service_name", sub.ServiceName,
	)

	return nil
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions 
		WHERE id = $1
	`

	r.logger.Debug(ctx, "Getting subscription from database",
		"subscription_id", id,
	)

	var sub model.Subscription
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.MonthlyCost,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		r.logger.Debug(ctx, "Subscription not found in database",
			"subscription_id", id,
		)
		return nil, nil
	}
	if err != nil {
		r.logger.Error(ctx, "Failed to get subscription from database",
			"subscription_id", id,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	r.logger.Debug(ctx, "Subscription retrieved successfully",
		"subscription_id", id,
		"service_name", sub.ServiceName,
	)

	return &sub, nil
}

func (r *subscriptionRepo) Update(ctx context.Context, id uuid.UUID, sub *model.Subscription) error {
	query := `
		UPDATE subscriptions 
		SET service_name = $1, monthly_cost = $2, user_id = $3, start_date = $4, end_date = $5
		WHERE id = $6
	`

	r.logger.Info(ctx, "Updating subscription in database",
		"subscription_id", id,
		"service_name", sub.ServiceName,
		"user_id", sub.UserID,
	)

	result, err := r.db.ExecContext(ctx, query,
		sub.ServiceName,
		sub.MonthlyCost,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		id,
	)

	if err != nil {
		r.logger.Error(ctx, "Failed to update subscription in database",
			"subscription_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error(ctx, "Failed to get rows affected",
			"subscription_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn(ctx, "Subscription not found for update",
			"subscription_id", id,
		)
		return fmt.Errorf("subscription not found")
	}

	r.logger.Info(ctx, "Subscription updated successfully",
		"subscription_id", id,
	)
	return nil
}

func (r *subscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	r.logger.Info(ctx, "Deleting subscription from database",
		"subscription_id", id,
	)

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error(ctx, "Failed to delete subscription from database",
			"subscription_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error(ctx, "Failed to get rows affected",
			"subscription_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn(ctx, "Subscription not found for deletion",
			"subscription_id", id,
		)
		return fmt.Errorf("subscription not found")
	}

	r.logger.Info(ctx, "Subscription deleted successfully",
		"subscription_id", id,
	)
	return nil
}

func (r *subscriptionRepo) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	query := `
		SELECT id, service_name, monthly_cost, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions 
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	r.logger.Debug(ctx, "Listing subscriptions from database",
		"user_id", userID,
		"service_name", serviceName,
	)

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *userID)
		argPos++
	}

	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argPos)
		args = append(args, *serviceName)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error(ctx, "Failed to list subscriptions from database",
			"user_id", userID,
			"service_name", serviceName,
			"error", err,
		)
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*model.Subscription
	for rows.Next() {
		var sub model.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.MonthlyCost,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			r.logger.Error(ctx, "Failed to scan subscription row",
				"error", err,
			)
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, &sub)
	}

	r.logger.Debug(ctx, "Subscriptions listed successfully",
		"count", len(subscriptions),
		"user_id", userID,
	)

	return subscriptions, nil
}

func (r *subscriptionRepo) CalculateTotalCost(ctx context.Context, filter model.SummaryFilter) (int, error) {
	query := `
		SELECT 
			COALESCE(SUM(
				monthly_cost * (
					-- Количество месяцев, которые подписка активна в указанном периоде
					LEAST(
						EXTRACT(YEAR FROM age($1, start_date)) * 12 + EXTRACT(MONTH FROM age($1, start_date)),
						EXTRACT(YEAR FROM age(end_date, $2)) * 12 + EXTRACT(MONTH FROM age(end_date, $2)) + 1,
						EXTRACT(YEAR FROM age($1, $2)) * 12 + EXTRACT(MONTH FROM age($1, $2)) + 1
					)
				)
			), 0)
		FROM subscriptions 
		WHERE start_date <= $1  -- подписка началась до конца периода
			AND (end_date IS NULL OR end_date >= $2)  -- подписка активна после начала периода
	`

	r.logger.Debug(ctx, "Calculating total cost in database",
		"start_period", filter.StartPeriod,
		"end_period", filter.EndPeriod,
		"user_id", filter.UserID,
		"service_name", filter.ServiceName,
	)

	// Парсим периоды используя ParseMonthYear (формат "01-2006")
	startPeriod, err := model.ParseMonthYear(filter.StartPeriod)
	if err != nil {
		r.logger.Error(ctx, "Invalid start period format",
			"start_period", filter.StartPeriod,
			"error", err,
		)
		return 0, fmt.Errorf("invalid start period format, expected MM-YYYY: %w", err)
	}

	endPeriod, err := model.ParseMonthYear(filter.EndPeriod)
	if err != nil {
		r.logger.Error(ctx, "Invalid end period format",
			"end_period", filter.EndPeriod,
			"error", err,
		)
		return 0, fmt.Errorf("invalid end period format, expected MM-YYYY: %w", err)
	}

	// Начало и конец периода
	periodStart := time.Date(startPeriod.Year(), startPeriod.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(endPeriod.Year(), endPeriod.Month()+1, 0, 23, 59, 59, 0, time.UTC) // последний день месяца

	args := []interface{}{
		periodEnd,   // конец периода
		periodStart, // начало периода
	}

	argPos := 3

	// Добавляем фильтры
	conditions := []string{}
	if filter.UserID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPos))
		args = append(args, filter.UserID)
		argPos++
	}

	if filter.ServiceName != "" {
		conditions = append(conditions, fmt.Sprintf("service_name = $%d", argPos))
		args = append(args, filter.ServiceName)
		argPos++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var totalCost int
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&totalCost)
	if err != nil {
		r.logger.Error(ctx, "Failed to calculate total cost in database",
			"start_period", filter.StartPeriod,
			"end_period", filter.EndPeriod,
			"error", err,
		)
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	r.logger.Info(ctx, "Total cost calculated successfully",
		"total_cost", totalCost,
		"start_period", filter.StartPeriod,
		"end_period", filter.EndPeriod,
	)

	return totalCost, nil
}
