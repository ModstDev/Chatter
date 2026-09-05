-- name: ListMessages :many
SELECT
    id,
    conversation_id,
    sender_id,
    content,
    created_at
FROM messages
WHERE conversation_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: ListMessagesBefore :many
SELECT
    id,
    conversation_id,
    sender_id,
    content,
    created_at
FROM messages
WHERE conversation_id = ?
  AND (
      created_at < ?
      OR (
          created_at = ?
          AND id < ?
      )
  )
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: CreateMessage :exec
INSERT INTO messages (
    id,
    conversation_id,
    sender_id,
    content
)
VALUES (?, ?, ?, ?);

-- name: GetMessageByID :one
SELECT
    id,
    conversation_id,
    sender_id,
    content,
    created_at
FROM messages
WHERE id = ?
LIMIT 1;

-- name: ListMessagesAfter :many
SELECT
    id,
    conversation_id,
    sender_id,
    content,
    created_at
FROM messages
WHERE conversation_id = ?
  AND (
      created_at > ?
      OR (
          created_at = ?
          AND id > ?
      )
  )
ORDER BY created_at ASC, id ASC
LIMIT ?;