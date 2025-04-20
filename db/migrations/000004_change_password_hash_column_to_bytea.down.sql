-- Down Migration
ALTER TABLE users
ALTER COLUMN password_hash TYPE TEXT;

ALTER TABLE users
ALTER COLUMN password_hash SET NOT NULL;
