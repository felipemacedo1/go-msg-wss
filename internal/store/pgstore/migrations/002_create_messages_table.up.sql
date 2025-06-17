-- 002_create_messages_table.up.sql

CREATE TABLE IF NOT EXISTS messages (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    reaction_count BIGINT NOT NULL DEFAULT 0,
    answered BOOLEAN NOT NULL DEFAULT FALSE,
    author_id TEXT NOT NULL DEFAULT 'guest',
    author_name TEXT NOT NULL DEFAULT 'Guest',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );
