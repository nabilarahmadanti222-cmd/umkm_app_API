package handler

import (
    "encoding/json"
    "errors"
    "net/http"

    "umkm-api/internal/domain"
    "umkm-api/internal/middleware"
    "umkm-api/internal/service"
)

type AuthHandler struct {
    svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
    return &AuthHandler{svc: svc}
}

type registerRequest struct {
    Nama     string `json:"nama"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

// Register godoc
// @Summary      Register akun baru
// @Description  Mendaftarkan pengguna baru dengan nama, email, dan password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body registerRequest true "Data registrasi"
// @Success      201 {object} handler.DataUserResponse
// @Failure      400 {object} handler.ErrorResponse
// @Failure      409 {object} handler.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "body tidak valid", "BAD_REQUEST")
        return
    }

    user, err := h.svc.Register(r.Context(), req.Nama, req.Email, req.Password)
    switch {
    case errors.Is(err, domain.ErrInvalidInput):
        writeError(w, http.StatusBadRequest, "nama, email, dan password (min 6 karakter) wajib diisi", "BAD_REQUEST")
    case errors.Is(err, domain.ErrConflict):
        writeError(w, http.StatusConflict, "email sudah terdaftar", "CONFLICT")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        writeJSON(w, http.StatusCreated, user)
    }
}

type loginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// Login godoc
// @Summary      Login
// @Description  Login dengan email dan password, mengembalikan JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body loginRequest true "Kredensial login"
// @Success      200 {object} handler.DataLoginResponse
// @Failure      401 {object} handler.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "body tidak valid", "BAD_REQUEST")
        return
    }

    token, user, err := h.svc.Login(r.Context(), req.Email, req.Password)
    switch {
    case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrInvalidInput):
        writeError(w, http.StatusUnauthorized, "email atau password salah", "UNAUTHORIZED")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
    }
}

// Profile godoc
// @Summary      Lihat profil sendiri
// @Description  Mengambil data profil pengguna yang sedang login
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} handler.DataUserResponse
// @Failure      401 {object} handler.ErrorResponse
// @Router       /profile [get]
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r)
    user, err := h.svc.Profile(r.Context(), userID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
        return
    }
    writeJSON(w, http.StatusOK, user)
}

type updateProfileRequest struct {
    Nama  string `json:"nama"`
    Email string `json:"email"`
}

// UpdateProfile godoc
// @Summary      Update profil
// @Description  Mengubah nama dan email akun yang sedang login
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body updateProfileRequest true "Data profil baru"
// @Success      200 {object} handler.MessageResponse
// @Failure      400 {object} handler.ErrorResponse
// @Router       /profile [put]
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r)
    var req updateProfileRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "body tidak valid", "BAD_REQUEST")
        return
    }

    if err := h.svc.UpdateProfile(r.Context(), userID, req.Nama, req.Email); err != nil {
        writeError(w, http.StatusBadRequest, "gagal memperbarui profil", "BAD_REQUEST")
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"message": "profil berhasil diperbarui"})
}

type updatePasswordRequest struct {
    PasswordLama string `json:"password_lama"`
    PasswordBaru string `json:"password_baru"`
}

// UpdatePassword godoc
// @Summary      Update password
// @Description  Mengubah password akun yang sedang login, wajib menyertakan password lama
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body updatePasswordRequest true "Password lama dan baru"
// @Success      200 {object} handler.MessageResponse
// @Failure      400 {object} handler.ErrorResponse
// @Router       /profile/password [put]
func (h *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r)
    var req updatePasswordRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "body tidak valid", "BAD_REQUEST")
        return
    }

    err := h.svc.UpdatePassword(r.Context(), userID, req.PasswordLama, req.PasswordBaru)
    switch {
    case errors.Is(err, domain.ErrInvalidInput):
        writeError(w, http.StatusBadRequest, "password lama salah atau password baru terlalu pendek", "BAD_REQUEST")
    case err != nil:
        writeError(w, http.StatusInternalServerError, "kesalahan server", "INTERNAL")
    default:
        writeJSON(w, http.StatusOK, map[string]string{"message": "password berhasil diperbarui"})
    }
}