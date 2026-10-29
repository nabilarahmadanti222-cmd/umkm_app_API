// @title           UMKM API
// @version         1.0
// @description     REST API untuk pendataan UMKM: registrasi, login, kelola profil, kelola UMKM milik sendiri, dan kelola produk.
// @BasePath        /api
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Ketik "Bearer" diikuti spasi dan token JWT. Contoh: "Bearer eyJhbGciOi..."
package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    httpSwagger "github.com/swaggo/http-swagger/v2"
    "golang.org/x/time/rate"

    _ "umkm-api/docs"
    "umkm-api/internal/config"
    "umkm-api/internal/handler"
    "umkm-api/internal/middleware"
    "umkm-api/internal/repository"
    "umkm-api/internal/service"
)

func main() {
    cfg := config.Load()

    ctx := context.Background()
    db, err := pgxpool.New(ctx, cfg.DatabaseURL)
    if err != nil {
        slog.Error("gagal konek ke database", "err", err)
        os.Exit(1)
    }
    defer db.Close()

    if err := db.Ping(ctx); err != nil {
        slog.Error("database tidak dapat dijangkau", "err", err)
        os.Exit(1)
    }
    slog.Info("berhasil terhubung ke PostgreSQL (Neon)")

    userRepo := repository.NewUserRepository(db)
    umkmRepo := repository.NewUmkmRepository(db)
    produkRepo := repository.NewProdukRepository(db)

    authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
    umkmSvc := service.NewUmkmService(umkmRepo)
    produkSvc := service.NewProdukService(produkRepo, umkmRepo)

    authHandler := handler.NewAuthHandler(authSvc)
    umkmHandler := handler.NewUmkmHandler(umkmSvc, cfg.UploadDir)
    produkHandler := handler.NewProdukHandler(produkSvc, cfg.UploadDir)

    auth := middleware.RequireAuth(cfg.JWTSecret)

    // maksimal 5 percobaan login per menit per alamat IP, agar brute-force password dibatasi
    loginLimiter := middleware.RateLimit(rate.Every(time.Minute/5), 5)

    mux := http.NewServeMux()

    // auth
    mux.HandleFunc("POST /api/auth/register", authHandler.Register)
    mux.Handle("POST /api/auth/login", loginLimiter(http.HandlerFunc(authHandler.Login)))

    // profile (butuh login)
    mux.Handle("GET /api/profile", auth(http.HandlerFunc(authHandler.Profile)))
    mux.Handle("PUT /api/profile", auth(http.HandlerFunc(authHandler.UpdateProfile)))
    mux.Handle("PUT /api/profile/password", auth(http.HandlerFunc(authHandler.UpdatePassword)))

    // umkm
    mux.HandleFunc("GET /api/umkm", umkmHandler.List)
    mux.Handle("GET /api/umkm/saya", auth(http.HandlerFunc(umkmHandler.MyUmkm)))
    mux.HandleFunc("GET /api/umkm/{id}", umkmHandler.Detail)
    mux.Handle("POST /api/umkm", auth(http.HandlerFunc(umkmHandler.Create)))
    mux.Handle("PUT /api/umkm/{id}", auth(http.HandlerFunc(umkmHandler.Update)))
    mux.Handle("DELETE /api/umkm/{id}", auth(http.HandlerFunc(umkmHandler.Delete)))

    // produk
    mux.HandleFunc("GET /api/umkm/{id}/produk", produkHandler.ListByUmkm)
    mux.Handle("POST /api/umkm/{id}/produk", auth(http.HandlerFunc(produkHandler.Create)))
    mux.HandleFunc("GET /api/produk/{id}", produkHandler.Detail)
    mux.Handle("PUT /api/produk/{id}", auth(http.HandlerFunc(produkHandler.Update)))
    mux.Handle("DELETE /api/produk/{id}", auth(http.HandlerFunc(produkHandler.Delete)))

    // static file untuk foto yang sudah diupload
    mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

    // dokumentasi Swagger
    mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

    // health check
    mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    srv := &http.Server{
        Addr:              ":" + cfg.Port,
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       10 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    go func() {
        slog.Info("server berjalan", "port", cfg.Port, "swagger", "/swagger/index.html")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server berhenti", "err", err)
            os.Exit(1)
        }
    }()

    // graceful shutdown
    stop, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()
    <-stop.Done()

    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer shutdownCancel()
    _ = srv.Shutdown(shutdownCtx)
    slog.Info("server dimatikan dengan aman")
}