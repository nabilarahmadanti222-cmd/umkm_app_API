package repository

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "umkm-api/internal/domain"
)

type ProdukRepository interface {
    Create(ctx context.Context, p *domain.Produk) error
    FindByID(ctx context.Context, id int64) (*domain.Produk, error)
    FindByUmkmID(ctx context.Context, umkmID int64, limit, offset int) ([]domain.Produk, error)
    CountByUmkmID(ctx context.Context, umkmID int64) (int, error)
    Update(ctx context.Context, p *domain.Produk) error
    Delete(ctx context.Context, id int64) error
}

type produkRepository struct {
    db *pgxpool.Pool
}

func NewProdukRepository(db *pgxpool.Pool) ProdukRepository {
    return &produkRepository{db: db}
}

func (r *produkRepository) Create(ctx context.Context, p *domain.Produk) error {
    query := `INSERT INTO produk (umkm_id, nama, harga, deskripsi, foto)
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
    return r.db.QueryRow(ctx, query, p.UmkmID, p.Nama, p.Harga, p.Deskripsi, p.Foto).Scan(&p.ID)
}

func (r *produkRepository) FindByID(ctx context.Context, id int64) (*domain.Produk, error) {
    var p domain.Produk
    query := `SELECT id, umkm_id, nama, harga, deskripsi, foto FROM produk WHERE id = $1`
    err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.UmkmID, &p.Nama, &p.Harga, &p.Deskripsi, &p.Foto)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &p, nil
}

func (r *produkRepository) FindByUmkmID(ctx context.Context, umkmID int64, limit, offset int) ([]domain.Produk, error) {
    rows, err := r.db.Query(ctx,
        `SELECT id, umkm_id, nama, harga, deskripsi, foto FROM produk
         WHERE umkm_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3`,
        umkmID, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    list := []domain.Produk{}
    for rows.Next() {
        var p domain.Produk
        if err := rows.Scan(&p.ID, &p.UmkmID, &p.Nama, &p.Harga, &p.Deskripsi, &p.Foto); err != nil {
            return nil, err
        }
        list = append(list, p)
    }
    return list, rows.Err()
}

func (r *produkRepository) CountByUmkmID(ctx context.Context, umkmID int64) (int, error) {
    var total int
    err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM produk WHERE umkm_id = $1`, umkmID).Scan(&total)
    return total, err
}

func (r *produkRepository) Update(ctx context.Context, p *domain.Produk) error {
    query := `UPDATE produk SET nama = $1, harga = $2, deskripsi = $3, foto = $4 WHERE id = $5`
    _, err := r.db.Exec(ctx, query, p.Nama, p.Harga, p.Deskripsi, p.Foto, p.ID)
    return err
}

func (r *produkRepository) Delete(ctx context.Context, id int64) error {
    _, err := r.db.Exec(ctx, `DELETE FROM produk WHERE id = $1`, id)
    return err
}