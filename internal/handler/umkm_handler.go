package handler

import (
    "errors"
    "net/http"
    "strconv"

    "umkm-api/internal/domain"
    "umkm-api/internal/middleware"
    "umkm-api/internal/service"
)

type UmkmHandler struct {
    svc       *service.UmkmService
    uploadDir string
}

func NewUmkmHandler(svc *service.UmkmService, uploadDir string) *UmkmHandler {
    return &UmkmHandler{svc: svc, uploadDir: uploadDir}
}

// List godoc
// @Summary      Lihat semua UMKM
// @Description  Mendukung pagination dan pencarian nama
// @Tags         umkm
// @Produce      json
// @Param        q query string false "Kata kunci pencarian nama UMKM"
// @Param        page query int false "Halaman (default 1)"
// @Param        limit query int false "Jumlah data per halaman (default 10, maksimal 100)"
// @Success      200 {object} handler.DataUmkmListResponse
// @Router       /umkm [get]
func (h *UmkmHandler) List(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("q")
    page, limit := parsePagination(r)

    list, total, usedPage, usedLimit, err := h.svc.List(r.Context(), q, page, limit)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
        return
    }
    writePaginated(w, http.StatusOK, list, usedPage, usedLimit, total)
}

// Detail godoc
// @Summary      Detail UMKM
// @Tags         umkm
// @Produce      json
// @Param        id path int true "ID UMKM"
// @Success      200 {object} handler.DataUmkmResponse
// @Failure      404 {object} handler.ErrorResponse
// @Router       /umkm/{id} [get]
func (h *UmkmHandler) Detail(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "id tidak valid", "BAD_REQUEST")
        return
    }

    umkm, err := h.svc.Detail(r.Context(), id)
    switch {
    case errors.Is(err, domain.ErrNotFound):
        writeError(w, http.StatusNotFound, "UMKM tidak ditemukan", "NOT_FOUND")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        writeJSON(w, http.StatusOK, umkm)
    }
}

// MyUmkm godoc
// @Summary      Lihat UMKM saya
// @Tags         umkm
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} handler.DataUmkmResponse
// @Failure      404 {object} handler.ErrorResponse
// @Router       /umkm/saya [get]
func (h *UmkmHandler) MyUmkm(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r)
    umkm, err := h.svc.MyUmkm(r.Context(), userID)
    switch {
    case errors.Is(err, domain.ErrNotFound):
        writeError(w, http.StatusNotFound, "Anda belum memiliki UMKM", "NOT_FOUND")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        writeJSON(w, http.StatusOK, umkm)
    }
}

// Create godoc
// @Summary      Buat UMKM sendiri
// @Description  1 user hanya boleh punya 1 UMKM. Foto opsional.
// @Tags         umkm
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        nama formData string true "Nama UMKM"
// @Param        deskripsi formData string false "Deskripsi UMKM"
// @Param        alamat formData string false "Alamat UMKM"
// @Param        foto formData file false "Foto UMKM"
// @Success      201 {object} handler.DataUmkmResponse
// @Failure      400 {object} handler.ErrorResponse
// @Failure      409 {object} handler.ErrorResponse
// @Router       /umkm [post]
func (h *UmkmHandler) Create(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r)

    r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
    if err := r.ParseMultipartForm(5 << 20); err != nil {
        writeError(w, http.StatusBadRequest, "form tidak valid atau file terlalu besar", "BAD_REQUEST")
        return
    }

    nama := r.FormValue("nama")
    deskripsi := r.FormValue("deskripsi")
    alamat := r.FormValue("alamat")

    fotoPath, err := saveUploadedFile(r, "foto", h.uploadDir, "umkm")
    if errors.Is(err, errInvalidFileType) {
        writeError(w, http.StatusBadRequest, "foto harus berformat jpg/jpeg/png", "BAD_REQUEST")
        return
    }
    if err != nil {
        writeError(w, http.StatusBadRequest, "gagal menyimpan foto", "BAD_REQUEST")
        return
    }

    umkm, err := h.svc.CreateOwn(r.Context(), userID, nama, deskripsi, alamat, fotoPath)
    switch {
    case errors.Is(err, domain.ErrInvalidInput):
        writeError(w, http.StatusBadRequest, "nama UMKM wajib diisi", "BAD_REQUEST")
    case errors.Is(err, domain.ErrConflict):
        writeError(w, http.StatusConflict, "Anda sudah memiliki UMKM", "CONFLICT")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        writeJSON(w, http.StatusCreated, umkm)
    }
}

// Update godoc
// @Summary      Update UMKM sendiri
// @Tags         umkm
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID UMKM"
// @Param        nama formData string true "Nama UMKM"
// @Param        deskripsi formData string false "Deskripsi UMKM"
// @Param        alamat formData string false "Alamat UMKM"
// @Param        foto formData file false "Foto UMKM baru"
// @Success      200 {object} handler.DataUmkmResponse
// @Failure      400 {object} handler.ErrorResponse
// @Failure      403 {object} handler.ErrorResponse
// @Router       /umkm/{id} [put]
func (h *UmkmHandler) Update(w http.ResponseWriter, r *http.Request) {
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
    alamat := r.FormValue("alamat")

    fotoBaru, err := saveUploadedFile(r, "foto", h.uploadDir, "umkm")
    if errors.Is(err, errInvalidFileType) {
        writeError(w, http.StatusBadRequest, "foto harus berformat jpg/jpeg/png", "BAD_REQUEST")
        return
    }
    if err != nil {
        writeError(w, http.StatusBadRequest, "gagal menyimpan foto", "BAD_REQUEST")
        return
    }

    umkm, fotoLama, err := h.svc.UpdateOwn(r.Context(), userID, id, nama, deskripsi, alamat, fotoBaru)
    switch {
    case errors.Is(err, domain.ErrInvalidInput):
        writeError(w, http.StatusBadRequest, "nama UMKM wajib diisi", "BAD_REQUEST")
    case errors.Is(err, domain.ErrForbidden):
        writeError(w, http.StatusForbidden, "Anda bukan pemilik UMKM ini", "FORBIDDEN")
    case errors.Is(err, domain.ErrNotFound):
        writeError(w, http.StatusNotFound, "UMKM tidak ditemukan", "NOT_FOUND")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        deleteFile(h.uploadDir, fotoLama)
        writeJSON(w, http.StatusOK, umkm)
    }
}

// Delete godoc
// @Summary      Hapus UMKM sendiri
// @Description  Produk pada UMKM ini ikut terhapus otomatis (ON DELETE CASCADE)
// @Tags         umkm
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID UMKM"
// @Success      200 {object} handler.MessageResponse
// @Failure      403 {object} handler.ErrorResponse
// @Failure      404 {object} handler.ErrorResponse
// @Router       /umkm/{id} [delete]
func (h *UmkmHandler) Delete(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r)
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "id tidak valid", "BAD_REQUEST")
        return
    }

    umkm, err := h.svc.DeleteOwn(r.Context(), userID, id)
    switch {
    case errors.Is(err, domain.ErrForbidden):
        writeError(w, http.StatusForbidden, "Anda bukan pemilik UMKM ini", "FORBIDDEN")
    case errors.Is(err, domain.ErrNotFound):
        writeError(w, http.StatusNotFound, "UMKM tidak ditemukan", "NOT_FOUND")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        deleteFile(h.uploadDir, umkm.Foto)
        writeJSON(w, http.StatusOK, map[string]string{"message": "UMKM berhasil dihapus"})
    }
}