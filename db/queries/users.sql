-- name: CreateUser :exec
INSERT INTO users (
    id,
    username,
    email,
    password_hash
)
VALUES (?, ?, ?, ?);


-- name: GetUserByID :one
SELECT
    id,
    username,
    email,
    password_hash,
    created_at
FROM users
WHERE id = ?
LIMIT 1;


-- name: GetUserByEmail :one
SELECT
    id,
    username,
    email,
    password_hash,
    created_at
FROM users
WHERE email = ?
LIMIT 1;


-- name: GetUserByUsername :one
SELECT
    id,
    username,
    email,
    password_hash,
    created_at
FROM users
WHERE username = ?
LIMIT 1;

-- name: ListUsers :many
SELECT
    id,
    username,
    email,
    password_hash,
    created_at
FROM users
ORDER BY username;