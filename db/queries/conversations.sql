-- name: CreateConversation :exec
INSERT INTO conversations (
    id
)
VALUES (?);

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
SELECT
    cm1.conversation_id
FROM conversation_members cm1
JOIN conversation_members cm2
    ON cm1.conversation_id = cm2.conversation_id
WHERE cm1.user_id = ?
    AND cm2.user_id = ?
LIMIT 1;