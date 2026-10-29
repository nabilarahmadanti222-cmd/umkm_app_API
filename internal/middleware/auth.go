package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"

func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            header := r.Header.Get("Authorization")
            if header == "" {
                http.Error(w, `{"error":{"message":"token tidak ditemukan","code":"UNAUTHORIZED"}}`, http.StatusUnauthorized)
                return
            }

            // Terima baik "Bearer <token>" maupun token mentah tanpa prefix,
            // supaya kotak Authorize di Swagger (skema apikey, tidak auto-tambah "Bearer ")
            // tidak selalu gagal kalau mahasiswa lupa mengetik kata "Bearer".
            tokenStr := header
            if len(header) > 7 && strings.EqualFold(header[:7], "Bearer ") {
                tokenStr = header[7:]
            }
            // jwt.WithValidMethods memaksa algoritma HS256 saja, agar token dengan
            // header "alg" lain (termasuk "none") selalu ditolak (algorithm confusion attack).
            token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
                return []byte(jwtSecret), nil
            }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
            if err != nil || !token.Valid {
                http.Error(w, `{"error":{"message":"token tidak valid","code":"UNAUTHORIZED"}}`, http.StatusUnauthorized)
                return
            }

            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok {
                http.Error(w, `{"error":{"message":"token tidak valid","code":"UNAUTHORIZED"}}`, http.StatusUnauthorized)
                return
            }
            sub, ok := claims["sub"].(float64)
            if !ok {
                http.Error(w, `{"error":{"message":"token tidak valid","code":"UNAUTHORIZED"}}`, http.StatusUnauthorized)
                return
            }
            userID := int64(sub)

            ctx := context.WithValue(r.Context(), UserIDKey, userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func GetUserID(r *http.Request) int64 {
    id, _ := r.Context().Value(UserIDKey).(int64)
    return id
}