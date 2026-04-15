-- name: InsertPerson :one
INSERT INTO people(kind, is_active, has_system_user)
    VALUES (sqlc.arg('Kind'), sqlc.narg('IsActive'), sqlc.narg('HasSystemUser'))
RETURNING *;

-- name: UpdatePerson :execrows
UPDATE
    people
SET
    kind = COALESCE(sqlc.narg('Kind'), kind),
    is_active = COALESCE(sqlc.narg('IsActive'), is_active),
    has_system_user = COALESCE(sqlc.narg('HasSystemUser'), has_system_user),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: GetPerson :one
SELECT
    p.id,
    p.kind,
    p.is_active,
    p.has_system_user,
    p.created_at,
    p.updated_at,
    pi.full_name,
    PI.short_name,
    pi.gender_identity,
    pi.marital_status,
    pi.image_url,
    pi.birth_date,
    pi.cpf,
    pi.created_at AS identifications_created_at,
    pi.updated_at AS identifications_updated_at
FROM
    people p
    LEFT JOIN people_identifications pi ON p.id = pi.person_id
WHERE
    p.id = sqlc.arg('ID');

-- name: ListPeople :many
SELECT
    p.id,
    p.kind,
    p.is_active,
    p.has_system_user,
    p.created_at,
    p.updated_at,
    pi.full_name,
    PI.short_name,
    pi.gender_identity,
    pi.marital_status,
    pi.image_url,
    pi.birth_date,
    pi.cpf,
    pi.created_at AS identifications_created_at,
    pi.updated_at AS identifications_updated_at
FROM
    people p
    LEFT JOIN people_identifications pi ON p.id = pi.person_id
WHERE (sqlc.narg('Kind')::person_kind IS NULL
    OR p.kind = sqlc.narg('Kind')::person_kind)
AND (sqlc.narg('IsActive')::BOOLEAN IS NULL
    OR p.is_active = sqlc.narg('IsActive')::BOOLEAN)
AND (sqlc.narg('HasSystemUser')::BOOLEAN IS NULL
    OR p.has_system_user = sqlc.narg('HasSystemUser')::BOOLEAN)
ORDER BY
    p.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: DeletePerson :execrows
DELETE FROM people
WHERE id = sqlc.arg('ID');

