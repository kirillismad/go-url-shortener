-- name: GetLinkByHref :one
SELECT * FROM "links" WHERE "href" = $1;

-- name: CreateLink :one
INSERT INTO "links" ("short_id", "href") 
VALUES ($1, $2) 
RETURNING *;

-- name: IsLinkExistByShortID :one
SELECT EXISTS(SELECT 1 FROM "links" WHERE "short_id" = $1);