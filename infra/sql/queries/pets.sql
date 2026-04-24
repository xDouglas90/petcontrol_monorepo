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
    AND (sqlc.arg('Search')::text = ''
        OR p.name ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR pi.full_name ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR p.race ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR p.kind::text ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR p.size::text ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR p.temperament::text ILIKE '%' || sqlc.arg('Search')::text || '%')
    AND (sqlc.narg('Size')::pet_size IS NULL
        OR p.size = sqlc.narg('Size')::pet_size)
    AND (sqlc.narg('Kind')::pet_kind IS NULL
        OR p.kind = sqlc.narg('Kind')::pet_kind)
    AND (sqlc.narg('Temperament')::pet_temperament IS NULL
        OR p.temperament = sqlc.narg('Temperament')::pet_temperament)
    AND (sqlc.narg('Race')::text IS NULL
        OR p.race = sqlc.narg('Race')::text)
    AND (sqlc.narg('IsActive')::boolean IS NULL
        OR p.is_active = sqlc.narg('IsActive')::boolean)
ORDER BY
    p.is_active DESC,
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

-- name: GetPetDetailByIDAndCompanyID :one
SELECT
    p.id,
    p.owner_id,
    cc.company_id,
    c.person_id AS owner_person_id,
    pi.full_name AS owner_name,
    pi.short_name AS owner_short_name,
    pi.image_url AS owner_image_url,
    poc.email AS owner_email,
    poc.cellphone AS owner_cellphone,
    poc.has_whatsapp AS owner_has_whatsapp,
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
    LEFT JOIN LATERAL (
        SELECT
            contact.email,
            contact.cellphone,
            contact.has_whatsapp
        FROM
            people_contacts contact
        WHERE
            contact.person_id = c.person_id
            AND contact.is_primary = TRUE
        ORDER BY
            contact.created_at ASC
        LIMIT 1
    ) poc ON TRUE
WHERE
    cc.company_id = sqlc.arg('CompanyID')
    AND p.id = sqlc.arg('ID')
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL
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
    is_active = COALESCE(sqlc.narg('IsActive'), is_active),
    is_deceased = COALESCE(sqlc.narg('IsDeceased'), is_deceased),
    is_vaccinated = COALESCE(sqlc.narg('IsVaccinated'), is_vaccinated),
    is_neutered = COALESCE(sqlc.narg('IsNeutered'), is_neutered),
    is_microchipped = COALESCE(sqlc.narg('IsMicrochipped'), is_microchipped),
    microchip_number = COALESCE(sqlc.narg('MicrochipNumber'), microchip_number),
    microchip_expiration_date = COALESCE(sqlc.narg('MicrochipExpirationDate'), microchip_expiration_date),
    notes = COALESCE(sqlc.narg('Notes'), notes),
    deleted_at = CASE
        WHEN sqlc.narg('IsActive')::boolean = TRUE THEN NULL
        WHEN sqlc.narg('IsActive')::boolean = FALSE THEN COALESCE(deleted_at, now())
        ELSE deleted_at
    END,
    updated_at = now()
WHERE
    id = sqlc.arg('ID');

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
