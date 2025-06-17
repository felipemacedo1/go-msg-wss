ALTER TABLE messages
DROP COLUMN IF EXISTS author_id,
    DROP COLUMN IF EXISTS author_name;