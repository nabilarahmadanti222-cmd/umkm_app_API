package service

import (
    "context"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "umkm-api/internal/domain"
    "umkm-api/internal/repository"
)

type AuthService struct {
    userRepo  repository.UserRepository
    jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) *AuthService {
    return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, nama, email, password string) (*domain.User, error) {
    if nama == "" || email == "" || len(password) < 6 {
        return nil, domain.ErrInvalidInput
    }

    _, err := s.userRepo.FindByEmail(ctx, email)
    if err == nil {
        return nil, domain.ErrConflict // email sudah dipakai
    }
    if !errors.Is(err, domain.ErrNotFound) {
        return nil, err
    }

    hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    user := &domain.User{Nama: nama, Email: email, Password: string(hashed)}
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil {
        return "", nil, domain.ErrNotFound
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return "", nil, domain.ErrInvalidInput // password salah
    }

    token, err := s.generateToken(user.ID)
    if err != nil {
        return "", nil, err
    }
    return token, user, nil
}

func (s *AuthService) Profile(ctx context.Context, userID int64) (*domain.User, error) {
    return s.userRepo.FindByID(ctx, userID)
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID int64, nama, email string) error {
    if nama == "" || email == "" {
        return domain.ErrInvalidInput
    }
    return s.userRepo.UpdateProfile(ctx, userID, nama, email)
}

func (s *AuthService) UpdatePassword(ctx context.Context, userID int64, lama, baru string) error {
    user, err := s.userRepo.FindByID(ctx, userID)
    if err != nil {
        return err
    }
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(lama)); err != nil {
        return domain.ErrInvalidInput // password lama tidak cocok
    }
    if len(baru) < 6 {
        return domain.ErrInvalidInput
    }
    hashed, err := bcrypt.GenerateFromPassword([]byte(baru), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    return s.userRepo.UpdatePassword(ctx, userID, string(hashed))
}

func (s *AuthService) generateToken(userID int64) (string, error) {
    claims := jwt.MapClaims{
        "sub": userID,
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}