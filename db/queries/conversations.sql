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

-- name: ListDirectConversationsForUser :many
SELECT
    c.id,
    c.created_at,
    u.id AS other_user_id,
    u.username AS other_username
FROM conversations c
JOIN conversation_members cm
    ON cm.conversation_id = c.id
JOIN conversation_members other_cm
    ON other_cm.conversation_id = c.id
JOIN users u
    ON u.id = other_cm.user_id
WHERE cm.user_id = ?
  AND other_cm.user_id != ?
  AND c.user_low IS NOT NULL
  AND c.user_high IS NOT NULL
ORDER BY c.created_at DESC;

-- name: IsConversationMember :one
SELECT EXISTS (
    SELECT 1
    FROM conversation_members
    WHERE conversation_id = ?
        AND user_id = ?
);