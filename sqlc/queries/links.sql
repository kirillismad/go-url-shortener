-- name: GetLinkByHref :one
SELECT * FROM "links" WHERE "href" = $1;

-- name: GetLinkByShortID :one
SELECT * FROM "links" WHERE "short_id" = $1;

-- name: IsLinkExistByShortID :one
SELECT EXISTS(SELECT 1 FROM "links" WHERE "short_id" = $1);

-- name: CreateLink :one
INSERT INTO "links" ("short_id", "href") 
VALUES ($1, $2) 
RETURNING *;

-- name: UpdateLinkUsageInfo :exec
UPDATE "links" 
SET "usage_count" = "usage_count" + 1, "usage_at" = NOW()
WHERE "id" = $1;