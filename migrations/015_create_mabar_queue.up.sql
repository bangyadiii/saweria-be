CREATE TABLE mabar_queue (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    streamer_id      UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    donation_id      UUID         NOT NULL REFERENCES donations(id) ON DELETE CASCADE,
    donor_name       VARCHAR(255) NOT NULL DEFAULT '',
    ingame_username  VARCHAR(255) NOT NULL DEFAULT '',
    amount           BIGINT       NOT NULL DEFAULT 0,
    priority_tier    VARCHAR(20)  NOT NULL DEFAULT 'bronze',
    priority_order   INT          NOT NULL DEFAULT 0,
    status           VARCHAR(20)  NOT NULL DEFAULT 'waiting',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mabar_queue_streamer_status ON mabar_queue(streamer_id, status);
CREATE INDEX idx_mabar_queue_created_at      ON mabar_queue(created_at);
