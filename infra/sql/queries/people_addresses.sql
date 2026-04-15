-- name: InsertPersonAddress :execrows
INSERT INTO people_addresses(person_id, address_id, is_main, label)
    VALUES (sqlc.arg('PersonID'), sqlc.arg('AddressID'), sqlc.arg('IsMain'), sqlc.arg('Label'));

-- name: UpdatePersonAddress :execrows
UPDATE
    people_addresses
SET
    is_main = COALESCE(sqlc.narg('IsMain'), is_main),
    label = COALESCE(sqlc.narg('Label'), label),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: GetPersonAddress :one
SELECT
    pa.id,
    pa.person_id,
    pa.address_id,
    pa.is_main,
    pa.label,
    pa.created_at
FROM
    people_addresses pa
WHERE
    pa.person_id = sqlc.arg('PersonID')
    AND pa.address_id = sqlc.arg('AddressID');

-- name: ListPersonAddresses :many
SELECT
    pa.id,
    pa.person_id,
    pa.address_id,
    pa.is_main,
    pa.label,
    pa.created_at
FROM
    people_addresses pa
WHERE
    pa.person_id = sqlc.arg('PersonID')
ORDER BY
    pa.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: DeletePersonAddress :execrows
DELETE FROM people_addresses
WHERE person_id = sqlc.arg('PersonID')
    AND address_id = sqlc.arg('AddressID');

