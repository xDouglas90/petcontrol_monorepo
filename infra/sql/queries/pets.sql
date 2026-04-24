-- name: ListPetsByCompanyID :many
SELECT
    COUNT(*) OVER () AS total_count,
    p.id,
    p.owner_id,
    cc.company_id,
    pi.full_name AS owner_name,
    p.name,
    p.race,
    p.color,
    p.sex,
    p.size,
    p.kind,
    p.temperament,
    p.image_url,
    p.birth_date,
    p.is_active,
    p.is_deceased,
    p.is_vaccinated,
    p.is_neutered,
    p.is_microchipped,
    p.microchip_number,
    p.microchip_expiration_date,
    p.notes,
    p.created_at,
    p.updated_at,
    p.deleted_at
FROM
    pets p
    INNER JOIN clients c ON c.id = p.owner_id
    INNER JOIN company_clients cc ON cc.client_id = c.id
    INNER JOIN people_identifications pi ON pi.person_id = c.person_id
WHERE
    cc.company_id = sqlc.arg('CompanyID')
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL
    AND p.deleted_at IS NULL
    AND p.is_active = TRUE
    AND (sqlc.arg('Search')::text = ''
        OR p.name ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR pi.full_name ILIKE '%' || sqlc.arg('Search')::text || '%')
ORDER BY
    p.name ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: GetPetByIDAndCompanyID :one
SELECT
    p.id,
    p.owner_id,
    cc.company_id,
    pi.full_name AS owner_name,
    p.name,
    p.race,
    p.color,
    p.sex,
    p.size,
    p.kind,
    p.temperament,
    p.image_url,
    p.birth_date,
    p.is_active,
    p.is_deceased,
    p.is_vaccinated,
    p.is_neutered,
    p.is_microchipped,
    p.microchip_number,
    p.microchip_expiration_date,
    p.notes,
    p.created_at,
    p.updated_at,
    p.deleted_at
FROM
    pets p
    INNER JOIN clients c ON c.id = p.owner_id
    INNER JOIN company_clients cc ON cc.client_id = c.id
    INNER JOIN people_identifications pi ON pi.person_id = c.person_id
WHERE
    cc.company_id = sqlc.arg('CompanyID')
    AND p.id = sqlc.arg('ID')
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL
    AND p.deleted_at IS NULL
    AND p.is_active = TRUE
LIMIT 1;

-- name: CreatePet :one
INSERT INTO pets(name, race, color, sex, size, kind, temperament, image_url, birth_date, owner_id, is_active, is_deceased, is_vaccinated, is_neutered, is_microchipped, microchip_number, microchip_expiration_date, notes)
    VALUES (sqlc.arg('Name'), sqlc.arg('Race'), sqlc.arg('Color'), sqlc.arg('Sex'), sqlc.arg('Size'), sqlc.arg('Kind'), sqlc.arg('Temperament'), sqlc.narg('ImageUrl'), sqlc.narg('BirthDate'), sqlc.arg('OwnerID'), sqlc.narg('IsActive'), sqlc.narg('IsDeceased'), sqlc.narg('IsVaccinated'), sqlc.narg('IsNeutered'), sqlc.narg('IsMicrochipped'), sqlc.narg('MicrochipNumber'), sqlc.narg('MicrochipExpirationDate'), sqlc.narg('Notes'))
RETURNING
    *;

-- name: UpdatePet :execrows
UPDATE
    pets
SET
    owner_id = COALESCE(sqlc.narg('OwnerID'), owner_id),
    name = COALESCE(sqlc.narg('Name'), name),
    race = COALESCE(sqlc.narg('Race'), race),
    color = COALESCE(sqlc.narg('Color'), color),
    sex = COALESCE(sqlc.narg('Sex'), sex),
    size = COALESCE(sqlc.narg('Size'), size),
    kind = COALESCE(sqlc.narg('Kind'), kind),
    temperament = COALESCE(sqlc.narg('Temperament'), temperament),
    image_url = COALESCE(sqlc.narg('ImageUrl'), image_url),
    birth_date = COALESCE(sqlc.narg('BirthDate'), birth_date),
    is_deceased = COALESCE(sqlc.narg('IsDeceased'), is_deceased),
    is_vaccinated = COALESCE(sqlc.narg('IsVaccinated'), is_vaccinated),
    is_neutered = COALESCE(sqlc.narg('IsNeutered'), is_neutered),
    is_microchipped = COALESCE(sqlc.narg('IsMicrochipped'), is_microchipped),
    microchip_number = COALESCE(sqlc.narg('MicrochipNumber'), microchip_number),
    microchip_expiration_date = COALESCE(sqlc.narg('MicrochipExpirationDate'), microchip_expiration_date),
    notes = COALESCE(sqlc.narg('Notes'), notes),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
    AND is_active = TRUE;

-- name: DeletePet :execrows
UPDATE
    pets
SET
    is_active = FALSE,
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND deleted_at IS NULL
    AND is_active = TRUE;

-- name: ValidatePetOwnerByCompany :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            company_clients cc
            INNER JOIN clients c ON c.id = cc.client_id
        WHERE
            cc.company_id = sqlc.arg('CompanyID')
            AND cc.client_id = sqlc.arg('OwnerID')
            AND cc.is_active = TRUE
            AND c.deleted_at IS NULL) AS is_valid;

