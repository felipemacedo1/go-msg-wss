-- name: GetRoom :one
SELECT
    "id", "theme"
FROM rooms
WHERE id = $1;

-- name: GetRooms :many
SELECT
    "id", "theme"
FROM rooms;

-- name: InsertRoom :one
INSERT INTO rooms
    ( "theme" ) VALUES
    ( $1 )
RETURNING "id";

-- name: GetMessage :one
SELECT
    "id", "room_id", "message", "reaction_count", "answered", "author_id", "author_name", "created_at"
FROM messages
WHERE
    id = $1;

-- name: GetRoomMessages :many
SELECT
    "id", "room_id", "message", "reaction_count", "answered", "author_id", "author_name", "created_at"
FROM messages
WHERE
    room_id = $1;

-- name: InsertMessage :one
INSERT INTO messages
    ( "room_id", "message", "reaction_count", "answered", "author_id", "author_name", "created_at" ) VALUES
    ( $1, $2, 0, false, $3, $4, CURRENT_TIMESTAMP )
RETURNING "id", "room_id", "message", "reaction_count", "answered", "author_id", "author_name", "created_at";

-- name: ReactToMessage :one
UPDATE messages
SET
    reaction_count = reaction_count + 1
WHERE
    id = $1
RETURNING reaction_count;

-- name: RemoveReactionFromMessage :one
UPDATE messages
SET
    reaction_count = GREATEST(0, reaction_count - 1)
WHERE
    id = $1
RETURNING reaction_count;

-- name: MarkMessageAsAnswered :exec
UPDATE messages
SET
    answered = true
WHERE
    id = $1;
