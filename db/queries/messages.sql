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