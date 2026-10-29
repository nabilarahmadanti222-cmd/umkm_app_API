package handler

import "umkm-api/internal/domain"

type DataUserResponse struct {
    Data domain.User `json:"data"`
}

type DataLoginResponse struct {
    Data struct {
        Token string      `json:"token"`
        User  domain.User `json:"user"`
    } `json:"data"`
}

type DataUmkmResponse struct {
    Data domain.Umkm `json:"data"`
}

type DataUmkmListResponse struct {
    Data []domain.Umkm `json:"data"`
}

type DataProdukResponse struct {
    Data domain.Produk `json:"data"`
}

type DataProdukListResponse struct {
    Data []domain.Produk `json:"data"`
}

type MessageResponse struct {
    Data struct {
        Message string `json:"message"`
    } `json:"data"`
}

type ErrorResponse struct {
    Error struct {
        Message string `json:"message"`
        Code    string `json:"code"`
    } `json:"error"`
}