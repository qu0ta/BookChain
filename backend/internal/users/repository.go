package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	const query = `SELECT * FROM users WHERE id = $1`
	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("collect users: %w", err)
	}

	return &user, nil
}

func (r *Repository) GetByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	const query = `SELECT * FROM users WHERE telegram_id = $1`
	rows, err := r.pool.Query(ctx, query, telegramID)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("collect users: %w", err)
	}

	return &user, nil
}

func (r *Repository) Create(ctx context.Context, user *User) error {
	const query = `INSERT INTO users (telegram_id, username, first_name, last_name)
VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(
		ctx, query, user.TelegramID, user.Username,
		user.FirstName, user.LastName,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}
