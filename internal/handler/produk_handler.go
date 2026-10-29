package handler

import (
    "errors"
    "net/http"
    "strconv"

    "umkm-api/internal/domain"
    "umkm-api/internal/middleware"
    "umkm-api/internal/service"
)

type ProdukHandler struct {
    svc       *service.ProdukService
    uploadDir string
}

func NewProdukHandler(svc *service.ProdukService, uploadDir string) *ProdukHandler {
    return &ProdukHandler{svc: svc, uploadDir: uploadDir}
}

// Detail godoc
// @Summary      Detail produk
// @Tags         produk
// @Produce      json
// @Param        id path int true "ID Produk"
// @Success      200 {object} handler.DataProdukResponse
// @Failure      404 {object} handler.ErrorResponse
// @Router       /produk/{id} [get]
func (h *ProdukHandler) Detail(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "id tidak valid", "BAD_REQUEST")
        return
    }

    produk, err := h.svc.Detail(r.Context(), id)
    switch {
    case errors.Is(err, domain.ErrNotFound):
        writeError(w, http.StatusNotFound, "produk tidak ditemukan", "NOT_FOUND")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        writeJSON(w, http.StatusOK, produk)
    }
}

// ListByUmkm godoc
// @Summary      Lihat produk pada satu UMKM
// @Description  Mendukung pagination
// @Tags         produk
// @Produce      json
// @Param        id path int true "ID UMKM"
// @Param        page query int false "Halaman (default 1)"
// @Param        limit query int false "Jumlah data per halaman (default 10, maksimal 100)"
// @Success      200 {object} handler.DataProdukListResponse
// @Router       /umkm/{id}/produk [get]
func (h *ProdukHandler) ListByUmkm(w http.ResponseWriter, r *http.Request) {
    umkmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "id tidak valid", "BAD_REQUEST")
        return
    }
    page, limit := parsePagination(r)

    list, total, usedPage, usedLimit, err := h.svc.ListByUmkm(r.Context(), umkmID, page, limit)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
        return
    }
    writePaginated(w, http.StatusOK, list, usedPage, usedLimit, total)
}

// Create godoc
// @Summary      Tambah produk pada UMKM sendiri
// @Tags         produk
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID UMKM"
// @Param        nama formData string true "Nama produk"
// @Param        harga formData number true "Harga produk"
// @Param        deskripsi formData string false "Deskripsi produk"
// @Param        foto formData file false "Foto produk"
// @Success      201 {object} handler.DataProdukResponse
// @Failure      400 {object} handler.ErrorResponse
// @Failure      403 {object} handler.ErrorResponse
// @Router       /umkm/{id}/produk [post]
func (h *ProdukHandler) Create(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r)
    umkmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "id tidak valid", "BAD_REQUEST")
        return
    }

    r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
    if err := r.ParseMultipartForm(5 << 20); err != nil {
        writeError(w, http.StatusBadRequest, "form tidak valid atau file terlalu besar", "BAD_REQUEST")
        return
    }

    nama := r.FormValue("nama")
    deskripsi := r.FormValue("deskripsi")
    harga, _ := strconv.ParseFloat(r.FormValue("harga"), 64)

    fotoPath, err := saveUploadedFile(r, "foto", h.uploadDir, "produk")
    if errors.Is(err, errInvalidFileType) {
        writeError(w, http.StatusBadRequest, "foto harus berformat jpg/jpeg/png", "BAD_REQUEST")
        return
    }
    if err != nil {
        writeError(w, http.StatusBadRequest, "gagal menyimpan foto", "BAD_REQUEST")
        return
    }

    produk, err := h.svc.CreateForOwnUmkm(r.Context(), userID, umkmID, nama, harga, deskripsi, fotoPath)
    switch {
    case errors.Is(err, domain.ErrInvalidInput):
        writeError(w, http.StatusBadRequest, "nama dan harga produk wajib valid", "BAD_REQUEST")
    case errors.Is(err, domain.ErrForbidden):
        writeError(w, http.StatusForbidden, "Anda bukan pemilik UMKM ini", "FORBIDDEN")
    case errors.Is(err, domain.ErrNotFound):
        writeError(w, http.StatusNotFound, "UMKM tidak ditemukan", "NOT_FOUND")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        writeJSON(w, http.StatusCreated, produk)
    }
}

// Update godoc
// @Summary      Update produk sendiri
// @Tags         produk
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID Produk"
// @Param        nama formData string true "Nama produk"
// @Param        harga formData number true "Harga produk"
// @Param        deskripsi formData string false "Deskripsi produk"
// @Param        foto formData file false "Foto produk baru"
// @Success      200 {object} handler.DataProdukResponse
// @Failure      400 {object} handler.ErrorResponse
// @Failure      403 {object} handler.ErrorResponse
// @Router       /produk/{id} [put]
func (h *ProdukHandler) Update(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r)
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "id tidak valid", "BAD_REQUEST")
        return
    }

    r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
    if err := r.ParseMultipartForm(5 << 20); err != nil {
        writeError(w, http.StatusBadRequest, "form tidak valid atau file terlalu besar", "BAD_REQUEST")
        return
    }

    nama := r.FormValue("nama")
    deskripsi := r.FormValue("deskripsi")
    harga, _ := strconv.ParseFloat(r.FormValue("harga"), 64)

    fotoBaru, err := saveUploadedFile(r, "foto", h.uploadDir, "produk")
    if errors.Is(err, errInvalidFileType) {
        writeError(w, http.StatusBadRequest, "foto harus berformat jpg/jpeg/png", "BAD_REQUEST")
        return
    }
    if err != nil {
        writeError(w, http.StatusBadRequest, "gagal menyimpan foto", "BAD_REQUEST")
        return
    }

    produk, fotoLama, err := h.svc.UpdateOwn(r.Context(), userID, id, nama, harga, deskripsi, fotoBaru)
    switch {
    case errors.Is(err, domain.ErrInvalidInput):
        writeError(w, http.StatusBadRequest, "nama dan harga produk wajib valid", "BAD_REQUEST")
    case errors.Is(err, domain.ErrForbidden):
        writeError(w, http.StatusForbidden, "Anda bukan pemilik produk ini", "FORBIDDEN")
    case errors.Is(err, domain.ErrNotFound):
        writeError(w, http.StatusNotFound, "produk tidak ditemukan", "NOT_FOUND")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        deleteFile(h.uploadDir, fotoLama)
        writeJSON(w, http.StatusOK, produk)
    }
}

// Delete godoc
// @Summary      Hapus produk sendiri
// @Tags         produk
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID Produk"
// @Success      200 {object} handler.MessageResponse
// @Failure      403 {object} handler.ErrorResponse
// @Failure      404 {object} handler.ErrorResponse
// @Router       /produk/{id} [delete]
func (h *ProdukHandler) Delete(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r)
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "id tidak valid", "BAD_REQUEST")
        return
    }

    produk, err := h.svc.DeleteOwn(r.Context(), userID, id)
    switch {
    case errors.Is(err, domain.ErrForbidden):
        writeError(w, http.StatusForbidden, "Anda bukan pemilik produk ini", "FORBIDDEN")
    case errors.Is(err, domain.ErrNotFound):
        writeError(w, http.StatusNotFound, "produk tidak ditemukan", "NOT_FOUND")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        deleteFile(h.uploadDir, produk.Foto)
        writeJSON(w, http.StatusOK, map[string]string{"message": "produk berhasil dihapus"})
    }
}