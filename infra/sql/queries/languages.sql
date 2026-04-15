-- name: InsertLanguage :execrows
INSERT INTO languages(code, "name", native_name)
    VALUES (sqlc.arg('Code'), sqlc.arg('Name'), sqlc.arg('NativeName'));

-- name: GetLanguageByCode :one
SELECT
    l.id,
    l.code,
    l."name",
    l.native_name,
    l.created_at,
    l.updated_at
FROM
    languages l
WHERE
    l.code = sqlc.arg('Code');

-- name: ListLanguages :many
SELECT
    l.id,
    l.code,
    l."name",
    l.native_name,
    l.created_at,
    l.updated_at
FROM
    languages l
ORDER BY
    l.code ASC
LIMIT sqlc.arg('Limit')
    OFFSET sqlc.arg('Offset');

-- name: UpdateLanguage :execrows
UPDATE
    languages
SET
    code = coalesce(sqlc.narg('Code'), code),
    "name" = coalesce(sqlc.narg('Name'), "name"),
    native_name = coalesce(sqlc.narg('NativeName'), native_name),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: DeleteLanguage :execrows
DELETE FROM languages
WHERE id = sqlc.arg('ID');

