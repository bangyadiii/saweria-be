CREATE TABLE alert_queue (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    streamer_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    donation_id  UUID NOT NULL REFERENCES donations(id) ON DELETE CASCADE,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX idx_alert_queue_streamer_status ON alert_queue(streamer_id, status);
CREATE INDEX idx_alert_queue_created_at      ON alert_queue(created_at);
