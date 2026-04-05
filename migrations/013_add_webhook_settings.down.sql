ALTER TABLE users
    DROP COLUMN IF EXISTS webhook_enabled,
    DROP COLUMN IF EXISTS webhook_url,
    DROP COLUMN IF EXISTS webhook_token;
