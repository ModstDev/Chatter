-- name: CreateConversation :exec
INSERT INTO conversations (
    id,
    user_low,
    user_high
)
VALUES (?, ?, ?);

-- name: AddConversationMember :exec
INSERT INTO conversation_members (
    conversation_id,
    user_id
)
VALUES (?, ?);

-- name: GetConversationByID :one
SELECT
    id,
    created_at
FROM conversations
WHERE id = ?
LIMIT 1;

-- name: FindDirectConversation :one
SELECT id
FROM conversations
WHERE user_low = ?
  AND user_high = ?
LIMIT 1;