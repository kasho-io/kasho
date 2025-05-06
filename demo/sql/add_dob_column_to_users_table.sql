-- Add date of birth column to users table
ALTER TABLE users ADD COLUMN dob DATE;

-- Add a comment to the column for documentation
COMMENT ON COLUMN users.dob IS 'Date of birth of the user'; 