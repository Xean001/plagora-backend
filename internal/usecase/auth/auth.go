package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/plagora/backend/config"
	"github.com/plagora/backend/internal/domain/entity"
	"github.com/plagora/backend/internal/domain/repository"
	ucDomain "github.com/plagora/backend/internal/domain/usecase"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type authUseCase struct {
	userRepo repository.UserRepository
	cfg      config.JWTConfig
}

func New(userRepo repository.UserRepository, cfg config.JWTConfig) ucDomain.AuthUseCase {
	return &authUseCase{userRepo: userRepo, cfg: cfg}
}

func (a *authUseCase) Login(ctx context.Context, input ucDomain.LoginInput) (*ucDomain.TokenPair, error) {
	user, err := a.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return a.generateTokenPair(user)
}

func (a *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*ucDomain.TokenPair, error) {
	c, err := a.parseToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if c.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(c.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return a.generateTokenPair(user)
}

func (a *authUseCase) SeedAdminIfNeeded(ctx context.Context, email, password string) error {
	exists, err := a.userRepo.ExistsAny(ctx)
	if err != nil {
		return fmt.Errorf("checking users: %w", err)
	}
	if exists {
		return nil // already seeded
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         "Admin",
		CreatedAt:    time.Now(),
	}
	return a.userRepo.Create(ctx, user)
}

func (a *authUseCase) generateTokenPair(user *entity.User) (*ucDomain.TokenPair, error) {
	accessToken, err := a.signToken(user, "access", time.Duration(a.cfg.ExpirationHours)*time.Hour)
	if err != nil {
		return nil, err
	}
	refreshToken, err := a.signToken(user, "refresh", time.Duration(a.cfg.RefreshExpirationHours)*time.Hour)
	if err != nil {
		return nil, err
	}
	return &ucDomain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authUseCase) signToken(user *entity.User, tokenType string, duration time.Duration) (string, error) {
	c := &claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(a.cfg.Secret))
}

func (a *authUseCase) parseToken(tokenStr string) (*claims, error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(a.cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return c, nil
}

// ParseAccessToken is used by middleware to validate access tokens and extract user ID.
func ParseAccessToken(tokenStr, secret string) (userID uuid.UUID, email string, err error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid || c.Type != "access" {
		return uuid.Nil, "", ErrInvalidToken
	}
	id, err := uuid.Parse(c.UserID)
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}
	return id, c.Email, nil
}
