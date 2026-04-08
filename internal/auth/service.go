package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"saweria-be/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken       = errors.New("email already registered")
	ErrUsernameTaken    = errors.New("username already taken")
	ErrUsernameReserved = errors.New("username is reserved")
	ErrInvalidCreds     = errors.New("invalid email or password")
	ErrUserNotFound     = errors.New("user not found")
)

// reservedUsernames lists usernames that must not be claimed by any user because
// they collide with frontend routes, backend API paths, or well-known system
// accounts. Add entries in lowercase; comparison is case-insensitive.
var reservedUsernames = map[string]struct{}{
	// ── frontend app routes ───────────────────────────────────────────────────
	"login":       {},
	"register":    {},
	"dashboard":   {},
	"profile":     {},
	"donations":   {},
	"overlay":     {},
	"integration": {},
	"mabar":       {},
	"widget":      {},
	"api":         {},
	// ── backend API path segments ─────────────────────────────────────────────
	"auth":      {},
	"users":     {},
	"me":        {},
	"payment":   {},
	"wallet":    {},
	"websocket": {},
	"ws":        {},
	"webhook":   {},
	// ── reserved admin / system accounts ─────────────────────────────────────
	"admin":      {},
	"superadmin": {},
	"root":       {},
	"system":     {},
	"support":    {},
	"help":       {},
	"info":       {},
	"contact":    {},
	"mail":       {},
	"noreply":    {},
	"no-reply":   {},
	"bot":        {},
	"staff":      {},
	"moderator":  {},
	"mod":        {},
	"owner":      {},
	// ── brand / product names ─────────────────────────────────────────────────
	"saweria":    {},
	"tako":       {},
	"saweria-be": {},
	"official":   {},
	// ── common abuse targets ──────────────────────────────────────────────────
	"null":      {},
	"undefined": {},
	"anonymous": {},
	"test":      {},
	"demo":      {},
}

type TokenPair struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type Service interface {
	Register(ctx context.Context, email, username, password string) (*domain.User, *TokenPair, error)
	Login(ctx context.Context, email, password string) (*domain.User, *TokenPair, error)
	LoginOrRegisterGoogle(ctx context.Context, googleID, email, displayName, profileImage string) (*domain.User, *TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}

type service struct {
	repo             Repository
	jwtSecret        string
	jwtRefreshSecret string
	jwtExpiryHours   int
	jwtRefreshDays   int
}

func NewService(repo Repository, jwtSecret, jwtRefreshSecret string, expiryHours, refreshDays int) Service {
	return &service{
		repo:             repo,
		jwtSecret:        jwtSecret,
		jwtRefreshSecret: jwtRefreshSecret,
		jwtExpiryHours:   expiryHours,
		jwtRefreshDays:   refreshDays,
	}
}

func (s *service) Register(ctx context.Context, email, username, password string) (*domain.User, *TokenPair, error) {
	if _, ok := reservedUsernames[strings.ToLower(username)]; ok {
		return nil, nil, ErrUsernameReserved
	}

	if _, err := s.repo.FindByEmail(ctx, email); err == nil {
		return nil, nil, ErrEmailTaken
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, fmt.Errorf("auth.Register: check email: %w", err)
	}

	if _, err := s.repo.FindByUsername(ctx, username); err == nil {
		return nil, nil, ErrUsernameTaken
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, fmt.Errorf("auth.Register: check username: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("auth.Register: hash password: %w", err)
	}

	hashStr := string(hash)
	user := &domain.User{
		Email:        email,
		Username:     username,
		PasswordHash: &hashStr,
		DisplayName:  username,
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("auth.Register: %w", err)
	}

	tokens, err := s.generateTokenPair(created)
	if err != nil {
		return nil, nil, err
	}
	return created, tokens, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*domain.User, *TokenPair, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrInvalidCreds
	}
	if err != nil {
		return nil, nil, fmt.Errorf("auth.Login: %w", err)
	}

	if user.PasswordHash == nil {
		return nil, nil, ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCreds
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

func (s *service) LoginOrRegisterGoogle(ctx context.Context, googleID, email, displayName, profileImage string) (*domain.User, *TokenPair, error) {
	user, err := s.repo.UpsertGoogleUser(ctx, googleID, email, displayName, profileImage)
	if err != nil {
		return nil, nil, fmt.Errorf("auth.LoginOrRegisterGoogle: %w", err)
	}
	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtRefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, _ := claims["user_id"].(string)
	username, _ := claims["username"].(string)

	return s.buildTokenPair(userID, username)
}

func (s *service) generateTokenPair(user *domain.User) (*TokenPair, error) {
	return s.buildTokenPair(user.ID, user.Username)
}

func (s *service) buildTokenPair(userID, username string) (*TokenPair, error) {
	now := time.Now()

	accessClaims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      now.Add(time.Duration(s.jwtExpiryHours) * time.Hour).Unix(),
		"iat":      now.Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("auth: sign access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      now.Add(time.Duration(s.jwtRefreshDays) * 24 * time.Hour).Unix(),
		"iat":      now.Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.jwtRefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("auth: sign refresh token: %w", err)
	}

	return &TokenPair{Token: accessToken, RefreshToken: refreshToken}, nil
}
