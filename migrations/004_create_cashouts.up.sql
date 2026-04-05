CREATE TABLE cashouts (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount         BIGINT NOT NULL,
    fee            BIGINT NOT NULL DEFAULT 0,
    net_amount     BIGINT NOT NULL,
    bank_name      VARCHAR(255) NOT NULL,
    account_number VARCHAR(255) NOT NULL,
    account_name   VARCHAR(255) NOT NULL,
    status         VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cashouts_user_id ON cashouts(user_id);
CREATE INDEX idx_cashouts_status  ON cashouts(status);
