-- name: InsertAddress :execrows
INSERT INTO addresses(zip_code, street, number, complement, district, city, state, country)
    VALUES (sqlc.arg('ZipCode'), sqlc.arg('Street'), sqlc.arg('Number'), sqlc.arg('Complement'), sqlc.arg('District'), sqlc.arg('City'), sqlc.arg('State'), sqlc.arg('Country'));

-- name: CreateAddress :one
INSERT INTO addresses(zip_code, street, number, complement, district, city, state, country)
    VALUES (sqlc.arg('ZipCode'), sqlc.arg('Street'), sqlc.arg('Number'), sqlc.arg('Complement'), sqlc.arg('District'), sqlc.arg('City'), sqlc.arg('State'), sqlc.arg('Country'))
RETURNING
    id, zip_code, street, number, complement, district, city, state, country, created_at, updated_at;

-- name: UpdateAddress :execrows
UPDATE
    addresses
SET
    zip_code = COALESCE(sqlc.narg('ZipCode'), zip_code),
    street = COALESCE(sqlc.narg('Street'), street),
    number = COALESCE(sqlc.narg('Number'), number),
    complement = COALESCE(sqlc.narg('Complement'), complement),
    district = COALESCE(sqlc.narg('District'), district),
    city = COALESCE(sqlc.narg('City'), city),
    state = COALESCE(sqlc.narg('State'), state),
    country = COALESCE(sqlc.narg('Country'), country),
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
