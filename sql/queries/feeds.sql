-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
  INSERT INTO feed_follows (user_id, feed_id) VALUES ($1, $2)
  RETURNING *
)
SELECT
  inserted_feed_follow.*,
  feeds.name as feed_name,
  users.name as user_name
FROM inserted_feed_follow
INNER JOIN feeds ON inserted_feed_follow.feed_id=feeds.id
INNER JOIN users ON inserted_feed_follow.user_id=users.id;
-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;
-- name: GetFeedFollowForUser :many
SELECT feed_follows.id, users.name AS user_name, feeds.name AS feed_name
FROM feed_follows
JOIN feeds ON feed_follows.feed_id = feeds.id
JOIN users ON feed_follows.user_id = users.id
WHERE feed_follows.user_id = $1;
-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1 AND feed_follows.feed_id = (SELECT id FROM feeds WHERE url = $2);
-- name: MarkFeedFetched :exec
UPDATE feeds
SET updated_at=CURRENT_TIMESTAMP, last_fetched_at=CURRENT_TIMESTAMP
WHERE feeds.id = $1;
-- name: GenNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at DESC NULLS FIRST
LIMIT 1;
