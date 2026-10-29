package repository

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "umkm-api/internal/domain"
)

type UserRepository interface {
    Create(ctx context.Context, u *domain.User) error
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    FindByID(ctx context.Context, id int64) (*domain.User, error)
    UpdateProfile(ctx context.Context, id int64, nama, email string) error
    UpdatePassword(ctx context.Context, id int64, hashed string) error
}

type userRepository struct {
    db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *domain.User) error {
    query := `INSERT INTO users (nama, email, password) VALUES ($1, $2, $3) RETURNING id`
    return r.db.QueryRow(ctx, query, u.Nama, u.Email, u.Password).Scan(&u.ID)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    var u domain.User
    query := `SELECT id, nama, email, password FROM users WHERE email = $1`
    err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Nama, &u.Email, &u.Password)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
    var u domain.User
    query := `SELECT id, nama, email, password FROM users WHERE id = $1`
    err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Nama, &u.Email, &u.Password)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *userRepository) UpdateProfile(ctx context.Context, id int64, nama, email string) error {
    query := `UPDATE users SET nama = $1, email = $2 WHERE id = $3`
    _, err := r.db.Exec(ctx, query, nama, email, id)
    return err
}

func (r *userRepository) UpdatePassword(ctx context.Context, id int64, hashed string) error {
    query := `UPDATE users SET password = $1 WHERE id = $2`
    _, err := r.db.Exec(ctx, query, hashed, id)
    return err
}