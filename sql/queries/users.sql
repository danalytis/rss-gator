-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;
-- name: GetUser :one
SELECT * FROM users WHERE name = $1;
-- name: GetUsers :many
SELECT * FROM users;
-- name: Reset :exec
DELETE FROM users WHERE name IN ('kahya', 'holgith', 'ballan');
-- name: GetFeedsWithUserName :many
SELECT
  f.name AS feed_name,
  f.url AS feed_url,
  u.name AS user_name
FROM
  feeds f
JOIN
  users u ON f.user_id = u.id;
