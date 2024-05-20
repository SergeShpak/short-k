-- name: InsertURL :one
WITH
new_entry AS (
    INSERT INTO urls(url, slug)
    VALUES($1, $2)
    ON CONFLICT(url) DO NOTHING
    RETURNING url, slug
),
old_entry AS (
    SELECT url, slug
    FROM urls
    WHERE url = $1
)
SELECT url, slug
FROM new_entry
UNION ALL
SELECT url, slug
FROM old_entry
LIMIT 1;

-- name: GetURL :one
SELECT url
FROM urls
WHERE slug = $1;