package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"subscription-service/internal/models"
)

type ISubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id string) (*models.Subscription, error)
	Update(ctx context.Context, id string, req *models.UpdateSubscriptionRequest) (*models.Subscription, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]models.Subscription, error)
	GetTotalPrice(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error)
}

type SubscriptionRepository struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

func NewSubscriptionRepository(db *sqlx.DB, logger *logrus.Logger) ISubscriptionRepository {
	return &SubscriptionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowxContext(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create subscription")
		return err
	}

	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	var sub models.Subscription
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at 
	          FROM subscriptions WHERE id = $1`

	err := r.db.GetContext(ctx, &sub, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.WithError(err).Error("Failed to get subscription by ID")
		return nil, err
	}

	return &sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, id string, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	sub, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, nil
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		sub.Price = *req.Price
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			sub.EndDate = nil
		} else {
			endDate, err := time.Parse("01-2006", *req.EndDate)
			if err != nil {
				return nil, fmt.Errorf("invalid end_date format: %w", err)
			}
			sub.EndDate = &endDate
		}
	}

	query := `
		UPDATE subscriptions 
		SET service_name = $1, price = $2, end_date = $3
		WHERE id = $4
		RETURNING updated_at
	`

	err = r.db.QueryRowxContext(ctx, query, sub.ServiceName, sub.Price, sub.EndDate, id).Scan(&sub.UpdatedAt)
	if err != nil {
		r.logger.WithError(err).Error("Failed to update subscription")
		return nil, err
	}

	return sub, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete subscription")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *SubscriptionRepository) List(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	err := r.db.SelectContext(ctx, &subscriptions, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list subscriptions")
		return nil, err
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) GetTotalPrice(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
	var totalPrice int

	query := `
		SELECT COALESCE(SUM(
			price * (
				EXTRACT(YEAR FROM age_months) * 12 + EXTRACT(MONTH FROM age_months)
			)
		), 0)
		FROM (
			SELECT 
				price,
				AGE(
					LEAST(COALESCE(end_date, $4), $4),
					GREATEST(start_date, $3)
				) as age_months
			FROM subscriptions
			WHERE ($1 = '' OR user_id = $1::uuid)
			  AND ($2 = '' OR service_name ILIKE '%' || $2 || '%')
			  AND start_date <= $4
			  AND (end_date IS NULL OR end_date >= $3)
		) AS periods
	`

	err := r.db.GetContext(ctx, &totalPrice, query, userID, serviceName, startDate, endDate)
	if err != nil {
		r.logger.WithError(err).Error("Failed to calculate total price")
		return 0, err
	}

	return totalPrice, nil
}
