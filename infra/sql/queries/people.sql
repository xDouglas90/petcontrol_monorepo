-- name: InsertPerson :execrows
INSERT INTO people(kind, is_active, has_system_user)
    VALUES (sqlc.arg('Kind'), sqlc.narg('IsActive'), sqlc.narg('HasSystemUser'));

-- name: UpdatePerson :execrows
UPDATE
    people
SET
    kind = COALESCE(sqlc.arg('Kind'), kind),
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
WHERE (sqlc.arg('Kind') IS NULL
    OR p.kind = sqlc.arg('Kind'))
AND (sqlc.narg('IsActive') IS NULL
    OR p.is_active = sqlc.narg('IsActive'))
AND (sqlc.narg('HasSystemUser') IS NULL
    OR p.has_system_user = sqlc.narg('HasSystemUser'))
ORDER BY
    p.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: DeletePerson :execrows
DELETE FROM people
WHERE id = sqlc.arg('ID');

