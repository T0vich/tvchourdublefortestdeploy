package repository

import (
	"context"
	"database/sql"
	"errors"
	"trade-chain/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type chainRepository struct {
	db *pgxpool.Pool
}

func NewChainRepository(db *pgxpool.Pool) ChainRepository {
	return &chainRepository{db: db}
}

func (r *chainRepository) Create(ctx context.Context, chain *domain.Chain) (*domain.Chain, error) {
	query := `
		INSERT INTO chains (from_product_id, to_product_id, initiator_id, previous_chain_id, next_chain_id, status, message)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING chain_id, from_product_id, to_product_id, initiator_id, previous_chain_id, next_chain_id, status, message, created_at, updated_at
	`

	var created domain.Chain
	err := r.db.QueryRow(ctx, query,
		chain.FromProductID,
		chain.ToProductID,
		chain.InitiatorID,
		chain.PreviousChainID,
		chain.NextChainID,
		chain.Status,
		chain.Message,
	).Scan(
		&created.ChainID,
		&created.FromProductID,
		&created.ToProductID,
		&created.InitiatorID,
		&created.PreviousChainID,
		&created.NextChainID,
		&created.Status,
		&created.Message,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *chainRepository) GetByID(ctx context.Context, id string) (*domain.Chain, error) {
	query := `
		SELECT chain_id, from_product_id, to_product_id, initiator_id, previous_chain_id, next_chain_id, status, message, created_at, updated_at
		FROM chains
		WHERE chain_id = $1
	`

	var chain domain.Chain
	err := r.db.QueryRow(ctx, query, id).Scan(
		&chain.ChainID,
		&chain.FromProductID,
		&chain.ToProductID,
		&chain.InitiatorID,
		&chain.PreviousChainID,
		&chain.NextChainID,
		&chain.Status,
		&chain.Message,
		&chain.CreatedAt,
		&chain.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &chain, nil
}

func (r *chainRepository) GetByProductID(ctx context.Context, productID string) ([]domain.Chain, error) {
	query := `
		SELECT chain_id, from_product_id, to_product_id, initiator_id, previous_chain_id, next_chain_id, status, message, created_at, updated_at
		FROM chains
		WHERE from_product_id = $1 OR to_product_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chains []domain.Chain
	for rows.Next() {
		var chain domain.Chain
		err := rows.Scan(
			&chain.ChainID,
			&chain.FromProductID,
			&chain.ToProductID,
			&chain.InitiatorID,
			&chain.PreviousChainID,
			&chain.NextChainID,
			&chain.Status,
			&chain.Message,
			&chain.CreatedAt,
			&chain.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}
	return chains, rows.Err()
}

func (r *chainRepository) GetFullChain(ctx context.Context, chainID string) ([]domain.Chain, error) {
	query := `
		WITH RECURSIVE chain_path AS (
			SELECT chain_id, from_product_id, to_product_id, initiator_id, previous_chain_id, next_chain_id, status, message, created_at, updated_at
			FROM chains
			WHERE chain_id = $1
			UNION ALL
			SELECT c.chain_id, c.from_product_id, c.to_product_id, c.initiator_id, c.previous_chain_id, c.next_chain_id, c.status, c.message, c.created_at, c.updated_at
			FROM chains c
			INNER JOIN chain_path cp ON c.chain_id = cp.next_chain_id OR c.chain_id = cp.previous_chain_id
		)
		SELECT chain_id, from_product_id, to_product_id, initiator_id, previous_chain_id, next_chain_id, status, message, created_at, updated_at
		FROM chain_path
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chains []domain.Chain
	for rows.Next() {
		var chain domain.Chain
		err := rows.Scan(
			&chain.ChainID,
			&chain.FromProductID,
			&chain.ToProductID,
			&chain.InitiatorID,
			&chain.PreviousChainID,
			&chain.NextChainID,
			&chain.Status,
			&chain.Message,
			&chain.CreatedAt,
			&chain.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}
	return chains, rows.Err()
}

func (r *chainRepository) UpdateStatus(ctx context.Context, id string, status domain.ChainStatus) error {
	query := `UPDATE chains SET status = $1 WHERE chain_id = $2`
	result, err := r.db.Exec(ctx, query, string(status), id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CompleteExchange завершает обмен: меняет владельцев товаров и обновляет статус
func (r *chainRepository) CompleteExchange(ctx context.Context, chainID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Получить цепочку
	var chain domain.Chain
	err = tx.QueryRow(ctx, `
		SELECT chain_id, from_product_id, to_product_id, initiator_id, status
		FROM chains
		WHERE chain_id = $1
		FOR UPDATE
	`, chainID).Scan(
		&chain.ChainID,
		&chain.FromProductID,
		&chain.ToProductID,
		&chain.InitiatorID,
		&chain.Status,
	)
	if err != nil {
		return err
	}

	if chain.Status != string(domain.ChainActive) {
		return errors.New("chain must be active to complete")
	}

	// 2. Получить текущих владельцев
	var fromOwner, toOwner string
	err = tx.QueryRow(ctx, `
		SELECT customer_id FROM products WHERE product_id = $1
	`, chain.FromProductID).Scan(&fromOwner)
	if err != nil {
		return err
	}
	err = tx.QueryRow(ctx, `
		SELECT customer_id FROM products WHERE product_id = $1
	`, chain.ToProductID).Scan(&toOwner)
	if err != nil {
		return err
	}

	// 3. Обменять владельцев
	_, err = tx.Exec(ctx, `
		UPDATE products SET customer_id = $1 WHERE product_id = $2
	`, toOwner, chain.FromProductID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE products SET customer_id = $1 WHERE product_id = $2
	`, fromOwner, chain.ToProductID)
	if err != nil {
		return err
	}

	// 4. Обновить статус цепочки
	_, err = tx.Exec(ctx, `
		UPDATE chains SET status = $1 WHERE chain_id = $2
	`, string(domain.ChainCompleted), chainID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *chainRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM chains WHERE chain_id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}
	return nil
}
