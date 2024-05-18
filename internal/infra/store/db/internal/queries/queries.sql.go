// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package queries

import (
	"context"
)

const getURL = `-- name: GetURL :one
SELECT url
FROM urls
WHERE slug = $1
`

func (q *Queries) GetURL(ctx context.Context, slug string) (string, error) {
	row := q.db.QueryRow(ctx, getURL, slug)
	var url string
	err := row.Scan(&url)
	return url, err
}

const insertURL = `-- name: InsertURL :one
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
LIMIT 1
`

type InsertURLParams struct {
	Url  string
	Slug string
}

type InsertURLRow struct {
	Url  string
	Slug string
}

func (q *Queries) InsertURL(ctx context.Context, arg InsertURLParams) (InsertURLRow, error) {
	row := q.db.QueryRow(ctx, insertURL, arg.Url, arg.Slug)
	var i InsertURLRow
	err := row.Scan(&i.Url, &i.Slug)
	return i, err
}
