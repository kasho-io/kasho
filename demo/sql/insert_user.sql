WITH random_ssn AS (
    SELECT 
        LPAD(FLOOR(RANDOM() * 1000)::TEXT, 3, '0') || '-' ||
        LPAD(FLOOR(RANDOM() * 100)::TEXT, 2, '0') || '-' ||
        LPAD(FLOOR(RANDOM() * 10000)::TEXT, 4, '0') as ssn
)
INSERT INTO users (name, email, ssn)
SELECT 
    'Nina' || FLOOR(RANDOM() * 1000)::TEXT,
    'nina' || FLOOR(RANDOM() * 1000)::TEXT || '@demo.com',
    ssn
FROM random_ssn;
