-- Update all users with unique dates of birth between 18 and 80 years old
UPDATE users 
SET dob = CURRENT_DATE - (mod(id, 365 * 62) + 365 * 18) * INTERVAL '1 day'; 