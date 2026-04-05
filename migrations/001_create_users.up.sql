-- Create uuid extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    username      VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT,
    google_id     TEXT UNIQUE,
    profile_image TEXT,
    display_name  VARCHAR(255) NOT NULL DEFAULT '',
    bio           TEXT NOT NULL DEFAULT '',
    balance       BIGINT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
