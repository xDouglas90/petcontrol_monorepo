-- name: InsertTranslation :one
INSERT INTO translations (language_code, entity_table, entity_id, field, content)
    VALUES (
        sqlc.arg('LanguageCode'),
        sqlc.arg('EntityTable'),
        sqlc.arg('EntityID'),
        sqlc.arg('Field'),
        sqlc.arg('Content')
    )
RETURNING *;

-- name: GetTranslation :one
SELECT
    t.id,
    t.language_code,
    t.entity_table,
    t.entity_id,
    t.field,
    t.content,
    t.created_at,
    t.updated_at
FROM
    translations t
WHERE
    t.language_code = sqlc.arg('LanguageCode')
    AND t.entity_table = sqlc.arg('EntityTable')
    AND t.entity_id = sqlc.arg('EntityID')
    AND t.field = sqlc.arg('Field')
LIMIT 1;

-- name: UpsertTranslation :one
INSERT INTO translations (language_code, entity_table, entity_id, field, content)
    VALUES (
        sqlc.arg('LanguageCode'),
        sqlc.arg('EntityTable'),
        sqlc.arg('EntityID'),
        sqlc.arg('Field'),
        sqlc.arg('Content')
    )
ON CONFLICT (language_code, entity_table, entity_id, field)
    DO UPDATE SET
        content = EXCLUDED.content,
        updated_at = now()
RETURNING *;

-- name: DeleteTranslation :execrows
DELETE FROM translations
WHERE language_code = sqlc.arg('LanguageCode')
    AND entity_table = sqlc.arg('EntityTable')
    AND entity_id = sqlc.arg('EntityID')
    AND field = sqlc.arg('Field');

-- name: DeleteTranslationsByEntity :execrows
DELETE FROM translations
WHERE entity_table = sqlc.arg('EntityTable')
    AND entity_id = sqlc.arg('EntityID');

-- name: ListTranslationsByEntity :many
SELECT
    t.id,
    t.language_code,
    t.entity_table,
    t.entity_id,
    t.field,
    t.content,
    t.created_at,
    t.updated_at
FROM
    translations t
WHERE
    t.entity_table = sqlc.arg('EntityTable')
    AND t.entity_id = sqlc.arg('EntityID')
ORDER BY
    t.language_code ASC,
    t.field ASC;

-- name: ListTranslationsByEntityAndLanguage :many
SELECT
    t.id,
    t.language_code,
    t.entity_table,
    t.entity_id,
    t.field,
    t.content,
    t.created_at,
    t.updated_at
FROM
    translations t
WHERE
    t.entity_table = sqlc.arg('EntityTable')
    AND t.entity_id = sqlc.arg('EntityID')
    AND t.language_code = sqlc.arg('LanguageCode')
ORDER BY
    t.field ASC;

