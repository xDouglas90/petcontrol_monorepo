-- name: InsertPersonIdentifications :execrows
INSERT INTO people_identifications(person_id, full_name, short_name, gender_identity, marital_status, image_url, birth_date, cpf)
    VALUES (sqlc.arg('PersonID'), sqlc.arg('FullName'), sqlc.arg('ShortName'), sqlc.arg('GenderIdentity'), sqlc.arg('MaritalStatus'), sqlc.arg('ImageURL'), sqlc.arg('BirthDate'), sqlc.arg('CPF'));

-- name: UpdatePersonIdentifications :execrows
UPDATE
    people_identifications
SET
    full_name = COALESCE(sqlc.arg('FullName'), full_name),
    short_name = COALESCE(sqlc.arg('ShortName'), short_name),
    gender_identity = COALESCE(sqlc.arg('GenderIdentity'), gender_identity),
    marital_status = COALESCE(sqlc.arg('MaritalStatus'), marital_status),
    image_url = COALESCE(sqlc.arg('ImageURL'), image_url),
    birth_date = COALESCE(sqlc.arg('BirthDate'), birth_date),
    cpf = COALESCE(sqlc.arg('CPF'), cpf),
    updated_at = now()
WHERE
    person_id = sqlc.arg('PersonID');

-- name: GetPersonIdentifications :one
SELECT
    pi.person_id,
    pi.full_name,
    pi.short_name,
    pi.gender_identity,
    pi.marital_status,
    pi.image_url,
    pi.birth_date,
    pi.cpf,
    pi.created_at,
    pi.updated_at
FROM
    people_identifications pi
WHERE
    pi.person_id = sqlc.arg('PersonID');

