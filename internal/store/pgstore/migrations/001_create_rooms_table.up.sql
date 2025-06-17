-- 001_create_rooms_table.up.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS rooms (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    theme TEXT NOT NULL
    );
