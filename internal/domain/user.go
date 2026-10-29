package domain

import "errors"

var (
    ErrNotFound     = errors.New("data tidak ditemukan")
    ErrConflict     = errors.New("data sudah ada")
    ErrForbidden    = errors.New("tidak memiliki akses")
    ErrInvalidInput = errors.New("input tidak valid")
)

type User struct {
    ID       int64  `json:"id"`
    Nama     string `json:"nama"`
    Email    string `json:"email"`
    Password string `json:"-"`
}