-- name: CreatePost :exec
INSERT INTO posts (title, url, description, published_at, feed_id) VALUES ($1, $2, $3, $4, $5);
-- name: GetPostsForUser :many
SELECT DISTINCT ON (url) * FROM posts
WHERE feed_id IN (
    SELECT feed_id FROM feed_follows WHERE user_id = $1
)
ORDER BY url, published_at DESC NULLS LAST
LIMIT $2;
