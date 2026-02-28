package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
)

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	u := &entity.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, name, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	u := &entity.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, name, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *userRepo) Create(ctx context.Context, u *entity.User) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, created_at) VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Email, u.PasswordHash, u.Name, u.CreatedAt,
	)
	return err
}

func (r *userRepo) ExistsAny(ctx context.Context) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count > 0, err
}

// --- Client ---

type clientRepo struct {
	db *pgxpool.Pool
}

func NewClientRepository(db *pgxpool.Pool) repository.ClientRepository {
	return &clientRepo{db: db}
}

func (r *clientRepo) FindAll(ctx context.Context) ([]*entity.Client, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, phone, email, notes, created_at FROM clients ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*entity.Client
	for rows.Next() {
		c := &entity.Client{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Email, &c.Notes, &c.CreatedAt); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, nil
}

func (r *clientRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Client, error) {
	c := &entity.Client{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, phone, email, notes, created_at FROM clients WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Phone, &c.Email, &c.Notes, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("client not found")
	}
	return c, err
}

func (r *clientRepo) Create(ctx context.Context, c *entity.Client) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO clients (id, name, phone, email, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		c.ID, c.Name, c.Phone, c.Email, c.Notes, c.CreatedAt,
	)
	return err
}

func (r *clientRepo) Update(ctx context.Context, c *entity.Client) error {
	_, err := r.db.Exec(ctx,
		`UPDATE clients SET name=$1, phone=$2, email=$3, notes=$4 WHERE id=$5`,
		c.Name, c.Phone, c.Email, c.Notes, c.ID,
	)
	return err
}

func (r *clientRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM clients WHERE id=$1`, id)
	return err
}
