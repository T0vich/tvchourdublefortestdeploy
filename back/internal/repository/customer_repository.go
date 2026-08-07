package repository

import (
	"context"
	"database/sql"
	"errors"
	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type customerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, customer *domain.CreateCustomerDTO) (*domain.Customer, error) {
	query := `
		INSERT INTO customers (email, password_hash)
		VALUES ($1, $2)
		RETURNING customer_id, email, password_hash, created_at, updated_at
	`

	var created domain.Customer
	err := r.db.QueryRow(ctx, query, customer.Email, customer.Password).Scan(
		&created.CustomerID,
		&created.Email,
		&created.PasswordHash,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *customerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	query := `
		SELECT customer_id, email, password_hash, created_at, updated_at
		FROM customers
		WHERE customer_id = $1 AND is_active = TRUE
	`

	var customer domain.Customer
	err := r.db.QueryRow(ctx, query, id).Scan(
		&customer.CustomerID,
		&customer.Email,
		&customer.PasswordHash,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &customer, nil
}

func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	query := `
		SELECT customer_id, email, password_hash, created_at, updated_at
		FROM customers
		WHERE email = $1 AND is_active = TRUE
	`

	var customer domain.Customer
	err := r.db.QueryRow(ctx, query, email).Scan(
		&customer.CustomerID,
		&customer.Email,
		&customer.PasswordHash,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &customer, nil
}

func (r *customerRepository) Update(ctx context.Context, id string, customer *domain.UpdateCustomerDTO) (*domain.Customer, error) {
	query := `
		UPDATE customers
		SET email = COALESCE($1, email),
		    password_hash = COALESCE($2, password_hash)
		WHERE customer_id = $3
		RETURNING customer_id, email, password_hash, created_at, updated_at
	`

	var updated domain.Customer
	err := r.db.QueryRow(ctx, query, customer.Email, customer.Password, id).Scan(
		&updated.CustomerID,
		&updated.Email,
		&updated.PasswordHash,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &updated, nil
}

func (r *customerRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE customers SET is_active = FALSE WHERE customer_id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *customerRepository) List(ctx context.Context, offset, limit int) ([]domain.Customer, error) {
	query := `
		SELECT customer_id, email, password_hash, created_at, updated_at
		FROM customers
		WHERE is_active = TRUE
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		var customer domain.Customer
		err := rows.Scan(
			&customer.CustomerID,
			&customer.Email,
			&customer.PasswordHash,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	return customers, rows.Err()
}
