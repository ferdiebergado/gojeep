-- Up Migration
ALTER TABLE users
ALTER COLUMN password_hash TYPE BYTEA
USING password_hash::BYTEA;

ALTER TABLE users
ALTER COLUMN password_hash SET NOT NULL;
