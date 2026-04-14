-- name: InsertAddress :execrows
INSERT INTO addresses(person_id, zip_code, street, number, complement, district, city, state, country)
    VALUES (sqlc.arg('ZipCode'), sqlc.arg('Street'), sqlc.arg('Number'), sqlc.arg('Complement'), sqlc.arg('District'), sqlc.arg('City'), sqlc.arg('State'), sqlc.arg('Country'));

-- name: UpdateAddress :execrows
UPDATE
    addresses
SET
    zip_code = COALESCE(sqlc.arg('ZipCode'), zip_code),
    street = COALESCE(sqlc.arg('Street'), street),
    number = COALESCE(sqlc.arg('Number'), number),
    complement = COALESCE(sqlc.arg('Complement'), complement),
    district = COALESCE(sqlc.arg('District'), district),
    city = COALESCE(sqlc.arg('City'), city),
    state = COALESCE(sqlc.arg('State'), state),
    country = COALESCE(sqlc.arg('Country'), country),
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

-- name: GetAddress :one
SELECT
    a.zip_code,
    a.street,
    a.number,
    a.complement,
    a.district,
    a.city,
    a.state,
    a.country,
    a.created_at,
    a.updated_at
FROM
    addresses a
WHERE
    a.id = sqlc.arg('ID');

