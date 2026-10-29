package handler

import (
    "encoding/json"
    "net/http"
    "strconv"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]any{"data": data})
}

// paginationMeta adalah bentuk info halaman yang disisipkan pada respons list.
type paginationMeta struct {
    Page  int `json:"page"`
    Limit int `json:"limit"`
    Total int `json:"total"`
}

func writePaginated(w http.ResponseWriter, status int, data any, page, limit, total int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]any{
        "data": data,
        "meta": paginationMeta{Page: page, Limit: limit, Total: total},
    })
}

func writeError(w http.ResponseWriter, status int, message, code string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]any{
        "error": map[string]string{"message": message, "code": code},
    })
}

// parsePagination membaca query param "page" dan "limit". Nilai yang tidak ada
// atau tidak valid dibiarkan 0 agar dinormalisasi oleh lapisan service.
func parsePagination(r *http.Request) (page, limit int) {
    page, _ = strconv.Atoi(r.URL.Query().Get("page"))
    limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
    return page, limit
}