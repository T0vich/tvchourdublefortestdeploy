package repository

import (
	"context"
	"database/sql"
	"errors"
	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type reviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *domain.Review) (*domain.Review, error) {
	query := `
		INSERT INTO reviews (from_customer_id, to_customer_id, product_id, rating)
		VALUES ($1, $2, $3, $4)
		RETURNING review_id, from_customer_id, to_customer_id, product_id, rating, created_at, updated_at
	`

	var created domain.Review
	err := r.db.QueryRow(ctx, query,
		review.FromCustomerID,
		review.ToCustomerID,
		review.ProductID,
		review.Rating,
	).Scan(
		&created.ReviewID,
		&created.FromCustomerID,
		&created.ToCustomerID,
		&created.ProductID,
		&created.Rating,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *reviewRepository) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	query := `
		SELECT review_id, from_customer_id, to_customer_id, product_id, rating, created_at, updated_at
		FROM reviews
		WHERE review_id = $1
	`

	var review domain.Review
	err := r.db.QueryRow(ctx, query, id).Scan(
		&review.ReviewID,
		&review.FromCustomerID,
		&review.ToCustomerID,
		&review.ProductID,
		&review.Rating,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &review, nil
}

func (r *reviewRepository) GetByCustomerID(ctx context.Context, customerID string) ([]domain.Review, error) {
	query := `
		SELECT review_id, from_customer_id, to_customer_id, product_id, rating, created_at, updated_at
		FROM reviews
		WHERE to_customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var review domain.Review
		err := rows.Scan(
			&review.ReviewID,
			&review.FromCustomerID,
			&review.ToCustomerID,
			&review.ProductID,
			&review.Rating,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	return reviews, rows.Err()
}

func (r *reviewRepository) GetAverageRating(ctx context.Context, customerID string) (float64, error) {
	query := `
		SELECT COALESCE(AVG(rating), 0)::float
		FROM reviews
		WHERE to_customer_id = $1
	`

	var avg float64
	err := r.db.QueryRow(ctx, query, customerID).Scan(&avg)
	if err != nil {
		return 0, err
	}
	return avg, nil
}

func (r *reviewRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reviews WHERE review_id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}
