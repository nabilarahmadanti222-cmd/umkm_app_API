package service

import (
    "context"

    "umkm-api/internal/domain"
    "umkm-api/internal/repository"
)

type ProdukService struct {
    produkRepo repository.ProdukRepository
    umkmRepo   repository.UmkmRepository
}

func NewProdukService(produkRepo repository.ProdukRepository, umkmRepo repository.UmkmRepository) *ProdukService {
    return &ProdukService{produkRepo: produkRepo, umkmRepo: umkmRepo}
}

// ListByUmkm mengambil daftar produk milik satu UMKM dengan pagination.
// Mengembalikan list, total data, serta page/limit yang benar-benar dipakai.
func (s *ProdukService) ListByUmkm(ctx context.Context, umkmID int64, page, limit int) (list []domain.Produk, total, usedPage, usedLimit int, err error) {
    usedPage, usedLimit = normalizePagination(page, limit)
    offset := (usedPage - 1) * usedLimit

    list, err = s.produkRepo.FindByUmkmID(ctx, umkmID, usedLimit, offset)
    if err != nil {
        return nil, 0, usedPage, usedLimit, err
    }
    total, err = s.produkRepo.CountByUmkmID(ctx, umkmID)
    if err != nil {
        return nil, 0, usedPage, usedLimit, err
    }
    return list, total, usedPage, usedLimit, nil
}

func (s *ProdukService) Detail(ctx context.Context, id int64) (*domain.Produk, error) {
    return s.produkRepo.FindByID(ctx, id)
}

func (s *ProdukService) CreateForOwnUmkm(ctx context.Context, userID, umkmID int64, nama string, harga float64, deskripsi, foto string) (*domain.Produk, error) {
    if nama == "" || harga < 0 {
        return nil, domain.ErrInvalidInput
    }

    umkm, err := s.umkmRepo.FindByID(ctx, umkmID)
    if err != nil {
        return nil, err
    }
    if umkm.UserID != userID {
        return nil, domain.ErrForbidden // hanya pemilik UMKM yang boleh menambah produk
    }

    produk := &domain.Produk{UmkmID: umkmID, Nama: nama, Harga: harga, Deskripsi: deskripsi, Foto: foto}
    if err := s.produkRepo.Create(ctx, produk); err != nil {
        return nil, err
    }
    return produk, nil
}

// authorizeOwner memastikan userID adalah pemilik UMKM tempat produk berada.
func (s *ProdukService) authorizeOwner(ctx context.Context, userID, produkID int64) (*domain.Produk, error) {
    produk, err := s.produkRepo.FindByID(ctx, produkID)
    if err != nil {
        return nil, err
    }
    umkm, err := s.umkmRepo.FindByID(ctx, produk.UmkmID)
    if err != nil {
        return nil, err
    }
    if umkm.UserID != userID {
        return nil, domain.ErrForbidden
    }
    return produk, nil
}

func (s *ProdukService) UpdateOwn(ctx context.Context, userID, produkID int64, nama string, harga float64, deskripsi, fotoBaru string) (produk *domain.Produk, fotoLama string, err error) {
    if nama == "" || harga < 0 {
        return nil, "", domain.ErrInvalidInput
    }

    produk, err = s.authorizeOwner(ctx, userID, produkID)
    if err != nil {
        return nil, "", err
    }

    fotoLama = produk.Foto
    produk.Nama = nama
    produk.Harga = harga
    produk.Deskripsi = deskripsi
    if fotoBaru != "" {
        produk.Foto = fotoBaru
    } else {
        fotoLama = ""
    }

    if err := s.produkRepo.Update(ctx, produk); err != nil {
        return nil, "", err
    }
    return produk, fotoLama, nil
}

func (s *ProdukService) DeleteOwn(ctx context.Context, userID, produkID int64) (*domain.Produk, error) {
    produk, err := s.authorizeOwner(ctx, userID, produkID)
    if err != nil {
        return nil, err
    }
    if err := s.produkRepo.Delete(ctx, produkID); err != nil {
        return nil, err
    }
    return produk, nil
}