ALTER TABLE subscriptions DROP COLUMN IF EXISTS discount_percent;
DROP TABLE IF EXISTS password_resets;
ALTER TABLE users DROP COLUMN IF EXISTS updated_at;
