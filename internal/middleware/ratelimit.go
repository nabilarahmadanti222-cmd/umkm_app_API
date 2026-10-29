package middleware

import (
    "net"
    "net/http"
    "sync"

    "golang.org/x/time/rate"
)

// ipRateLimiter menyimpan satu rate.Limiter per alamat IP, supaya percobaan
// login yang bertubi-tubi dari satu IP (brute-force) dibatasi tanpa mengganggu
// pengguna lain yang mengakses endpoint yang sama secara wajar.
type ipRateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
    r        rate.Limit
    b        int
}

func newIPRateLimiter(r rate.Limit, b int) *ipRateLimiter {
    return &ipRateLimiter{
        limiters: make(map[string]*rate.Limiter),
        r:        r,
        b:        b,
    }
}

func (i *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    limiter, exists := i.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(i.r, i.b)
        i.limiters[ip] = limiter
    }
    return limiter
}

func clientIP(r *http.Request) string {
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return host
}

// RateLimit membatasi jumlah request per alamat IP. r adalah laju pengisian ulang
// token (request/detik) dan b adalah ukuran burst (jumlah request beruntun yang
// masih diizinkan sebelum limiter mulai menahan).
func RateLimit(r rate.Limit, b int) func(http.Handler) http.Handler {
    limiter := newIPRateLimiter(r, b)
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            if !limiter.getLimiter(clientIP(req)).Allow() {
                http.Error(w, `{"error":{"message":"terlalu banyak percobaan, coba lagi nanti","code":"TOO_MANY_REQUESTS"}}`, http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, req)
        })
    }
}