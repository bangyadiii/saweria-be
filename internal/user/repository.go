package user

import (
	"context"
	"fmt"

	"saweria-be/internal/domain"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, id string, req UpdateFields) (*domain.User, error)
	UpdateWebhook(ctx context.Context, id string, f WebhookFields) (*domain.User, error)
}

type UpdateFields struct {
	DisplayName  string
	Bio          string
	ProfileImage *string
}

type WebhookFields struct {
	Enabled bool
	URL     *string
	Token   *string
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("user.FindByID: %w", err)
	}
	return &u, nil
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE username = $1`, username)
	if err != nil {
		return nil, fmt.Errorf("user.FindByUsername: %w", err)
	}
	return &u, nil
}

func (r *repository) Update(ctx context.Context, id string, f UpdateFields) (*domain.User, error) {
	query := `
		UPDATE users
		SET display_name  = $1,
		    bio           = $2,
		    profile_image = $3,
		    updated_at    = NOW()
		WHERE id = $4
		RETURNING *`
	var u domain.User
	err := r.db.QueryRowxContext(ctx, query, f.DisplayName, f.Bio, f.ProfileImage, id).StructScan(&u)
	if err != nil {
		return nil, fmt.Errorf("user.Update: %w", err)
	}
	return &u, nil
}

func (r *repository) UpdateWebhook(ctx context.Context, id string, f WebhookFields) (*domain.User, error) {
	query := `
		UPDATE users
		SET webhook_enabled = $1,
		    webhook_url     = $2,
		    webhook_token   = $3,
		    updated_at      = NOW()
		WHERE id = $4
		RETURNING *`
	var u domain.User
	err := r.db.QueryRowxContext(ctx, query, f.Enabled, f.URL, f.Token, id).StructScan(&u)
	if err != nil {
		return nil, fmt.Errorf("user.UpdateWebhook: %w", err)
	}
	return &u, nil
}
