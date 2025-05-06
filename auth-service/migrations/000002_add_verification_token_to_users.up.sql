-- Add verification_token column to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS verification_token VARCHAR(255);

-- Create an index for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_verification_token ON users(verification_token);
