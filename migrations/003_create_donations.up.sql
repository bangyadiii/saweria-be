CREATE TABLE donations (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    streamer_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    donor_name           VARCHAR(255) NOT NULL,
    amount               BIGINT NOT NULL,
    net_amount           BIGINT NOT NULL DEFAULT 0,
    platform_fee         BIGINT NOT NULL DEFAULT 0,
    message              TEXT NOT NULL DEFAULT '',
    media_url            TEXT,
    media_shown          BOOLEAN NOT NULL DEFAULT false,
    payment_method       VARCHAR(100) NOT NULL DEFAULT '',
    payment_status       VARCHAR(50) NOT NULL DEFAULT 'pending',
    midtrans_order_id    VARCHAR(255) UNIQUE NOT NULL,
    payment_token        TEXT NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_donations_streamer_id       ON donations(streamer_id);
CREATE INDEX idx_donations_midtrans_order_id ON donations(midtrans_order_id);
CREATE INDEX idx_donations_payment_status    ON donations(payment_status);
