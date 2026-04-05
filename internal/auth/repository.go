package auth

import (
	"context"
	"fmt"

	"saweria-be/internal/domain"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	UpsertGoogleUser(ctx context.Context, googleID, email, displayName, profileImage string) (*domain.User, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1`, email)
	if err != nil {
		return nil, fmt.Errorf("auth.FindByEmail: %w", err)
	}
	return &u, nil
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE username = $1`, username)
	if err != nil {
		return nil, fmt.Errorf("auth.FindByUsername: %w", err)
	}
	return &u, nil
}

func (r *repository) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE google_id = $1`, googleID)
	if err != nil {
		return nil, fmt.Errorf("auth.FindByGoogleID: %w", err)
	}
	return &u, nil
}

func (r *repository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (email, username, password_hash, display_name)
		VALUES ($1, $2, $3, $4)
		RETURNING *`
	var created domain.User
	err := r.db.QueryRowxContext(ctx, query, u.Email, u.Username, u.PasswordHash, u.DisplayName).StructScan(&created)
	if err != nil {
		return nil, fmt.Errorf("auth.Create: %w", err)
	}
	return &created, nil
}

func (r *repository) UpsertGoogleUser(ctx context.Context, googleID, email, displayName, profileImage string) (*domain.User, error) {
	query := `
		INSERT INTO users (email, username, google_id, display_name, profile_image)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (google_id) DO UPDATE
		  SET display_name   = EXCLUDED.display_name,
		      profile_image  = EXCLUDED.profile_image,
		      updated_at     = NOW()
		RETURNING *`
	// derive username from email prefix, truncated to 20 chars
	username := deriveUsername(email)
	var u domain.User
	err := r.db.QueryRowxContext(ctx, query, email, username, googleID, displayName, profileImage).StructScan(&u)
	if err != nil {
		return nil, fmt.Errorf("auth.UpsertGoogleUser: %w", err)
	}
	return &u, nil
}

func deriveUsername(email string) string {
	for i, c := range email {
		if c == '@' {
			if i > 20 {
				return email[:20]
			}
			return email[:i]
		}
	}
	if len(email) > 20 {
		return email[:20]
	}
	return email
}
