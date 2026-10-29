package service

import (
    "context"
    "errors"

    "umkm-api/internal/domain"
    "umkm-api/internal/repository"
)

type UmkmService struct {
    umkmRepo repository.UmkmRepository
}

func NewUmkmService(umkmRepo repository.UmkmRepository) *UmkmService {
    return &UmkmService{umkmRepo: umkmRepo}
}

// List mengambil daftar UMKM dengan pagination, opsional difilter nama (q).
// Mengembalikan list, total data, serta page/limit yang benar-benar dipakai
// (setelah dinormalisasi) agar handler bisa menuliskan meta pagination yang akurat.
func (s *UmkmService) List(ctx context.Context, q string, page, limit int) (list []domain.Umkm, total, usedPage, usedLimit int, err error) {
    usedPage, usedLimit = normalizePagination(page, limit)
    offset := (usedPage - 1) * usedLimit

    list, err = s.umkmRepo.FindAll(ctx, q, usedLimit, offset)
    if err != nil {
        return nil, 0, usedPage, usedLimit, err
    }
    total, err = s.umkmRepo.Count(ctx, q)
    if err != nil {
        return nil, 0, usedPage, usedLimit, err
    }
    return list, total, usedPage, usedLimit, nil
}

func (s *UmkmService) Detail(ctx context.Context, id int64) (*domain.Umkm, error) {
    return s.umkmRepo.FindByID(ctx, id)
}

func (s *UmkmService) MyUmkm(ctx context.Context, userID int64) (*domain.Umkm, error) {
    return s.umkmRepo.FindByUserID(ctx, userID)
}

func (s *UmkmService) CreateOwn(ctx context.Context, userID int64, nama, deskripsi, alamat, foto string) (*domain.Umkm, error) {
    if nama == "" {
        return nil, domain.ErrInvalidInput
    }

    _, err := s.umkmRepo.FindByUserID(ctx, userID)
    if err == nil {
        return nil, domain.ErrConflict // setiap user hanya boleh punya 1 UMKM
    }
    if !errors.Is(err, domain.ErrNotFound) {
        return nil, err
    }

    umkm := &domain.Umkm{UserID: userID, Nama: nama, Deskripsi: deskripsi, Alamat: alamat, Foto: foto}
    if err := s.umkmRepo.Create(ctx, umkm); err != nil {
        return nil, err
    }
    return umkm, nil
}

// UpdateOwn mengembalikan UMKM hasil update beserta path foto lama (bisa kosong).
// Handler bertanggung jawab menghapus berkas foto lama dari disk setelah update berhasil.
func (s *UmkmService) UpdateOwn(ctx context.Context, userID, umkmID int64, nama, deskripsi, alamat, fotoBaru string) (umkm *domain.Umkm, fotoLama string, err error) {
    if nama == "" {
        return nil, "", domain.ErrInvalidInput
    }

    umkm, err = s.umkmRepo.FindByID(ctx, umkmID)
    if err != nil {
        return nil, "", err
    }
    if umkm.UserID != userID {
        return nil, "", domain.ErrForbidden
    }

    fotoLama = umkm.Foto
    umkm.Nama = nama
    umkm.Deskripsi = deskripsi
    umkm.Alamat = alamat
    if fotoBaru != "" {
        umkm.Foto = fotoBaru
    } else {
        fotoLama = "" // foto tidak diganti, tidak ada yang perlu dihapus
    }

    if err := s.umkmRepo.Update(ctx, umkm); err != nil {
        return nil, "", err
    }
    return umkm, fotoLama, nil
}

func (s *UmkmService) DeleteOwn(ctx context.Context, userID, umkmID int64) (*domain.Umkm, error) {
    umkm, err := s.umkmRepo.FindByID(ctx, umkmID)
    if err != nil {
        return nil, err
    }
    if umkm.UserID != userID {
        return nil, domain.ErrForbidden
    }
    if err := s.umkmRepo.Delete(ctx, umkmID); err != nil {
        return nil, err
    }
    return umkm, nil
}