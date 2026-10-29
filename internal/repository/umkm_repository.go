package repository

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "umkm-api/internal/domain"
)

type UmkmRepository interface {
    Create(ctx context.Context, u *domain.Umkm) error
    FindAll(ctx context.Context, q string, limit, offset int) ([]domain.Umkm, error)
    Count(ctx context.Context, q string) (int, error)
    FindByID(ctx context.Context, id int64) (*domain.Umkm, error)
    FindByUserID(ctx context.Context, userID int64) (*domain.Umkm, error)
    Update(ctx context.Context, u *domain.Umkm) error
    Delete(ctx context.Context, id int64) error
}

type umkmRepository struct {
    db *pgxpool.Pool
}

func NewUmkmRepository(db *pgxpool.Pool) UmkmRepository {
    return &umkmRepository{db: db}
}

func (r *umkmRepository) Create(ctx context.Context, u *domain.Umkm) error {
    query := `INSERT INTO umkm (user_id, nama, deskripsi, alamat, foto)
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
    return r.db.QueryRow(ctx, query, u.UserID, u.Nama, u.Deskripsi, u.Alamat, u.Foto).Scan(&u.ID)
}

// FindAll mengambil daftar UMKM dengan pagination, opsional difilter nama (q).
// q kosong berarti tidak difilter sama sekali.
func (r *umkmRepository) FindAll(ctx context.Context, q string, limit, offset int) ([]domain.Umkm, error) {
    query := `SELECT id, user_id, nama, deskripsi, alamat, foto FROM umkm
              WHERE ($1 = '' OR nama ILIKE '%' || $1 || '%')
              ORDER BY id DESC LIMIT $2 OFFSET $3`
    rows, err := r.db.Query(ctx, query, q, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    list := []domain.Umkm{} // inisialisasi slice kosong, bukan nil, agar JSON-nya "[]" bukan "null"
    for rows.Next() {
        var u domain.Umkm
        if err := rows.Scan(&u.ID, &u.UserID, &u.Nama, &u.Deskripsi, &u.Alamat, &u.Foto); err != nil {
            return nil, err
        }
        list = append(list, u)
    }
    return list, rows.Err()
}

func (r *umkmRepository) Count(ctx context.Context, q string) (int, error) {
    var total int
    query := `SELECT COUNT(*) FROM umkm WHERE ($1 = '' OR nama ILIKE '%' || $1 || '%')`
    err := r.db.QueryRow(ctx, query, q).Scan(&total)
    return total, err
}

func (r *umkmRepository) FindByID(ctx context.Context, id int64) (*domain.Umkm, error) {
    var u domain.Umkm
    query := `SELECT id, user_id, nama, deskripsi, alamat, foto FROM umkm WHERE id = $1`
    err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.UserID, &u.Nama, &u.Deskripsi, &u.Alamat, &u.Foto)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *umkmRepository) FindByUserID(ctx context.Context, userID int64) (*domain.Umkm, error) {
    var u domain.Umkm
    query := `SELECT id, user_id, nama, deskripsi, alamat, foto FROM umkm WHERE user_id = $1`
    err := r.db.QueryRow(ctx, query, userID).Scan(&u.ID, &u.UserID, &u.Nama, &u.Deskripsi, &u.Alamat, &u.Foto)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *umkmRepository) Update(ctx context.Context, u *domain.Umkm) error {
    query := `UPDATE umkm SET nama = $1, deskripsi = $2, alamat = $3, foto = $4 WHERE id = $5`
    _, err := r.db.Exec(ctx, query, u.Nama, u.Deskripsi, u.Alamat, u.Foto, u.ID)
    return err
}

func (r *umkmRepository) Delete(ctx context.Context, id int64) error {
    _, err := r.db.Exec(ctx, `DELETE FROM umkm WHERE id = $1`, id)
    return err
}