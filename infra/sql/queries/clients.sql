-- name: ListClientsByCompanyID :many
SELECT
    COUNT(*) OVER () AS total_count,
    c.id,
    c.person_id,
    cc.company_id,
    pi.full_name,
    pi.short_name,
    pi.gender_identity,
    pi.marital_status,
    pi.birth_date,
    pi.cpf,
    pc.email,
    pc.phone,
    pc.cellphone,
    pc.has_whatsapp,
    c.client_since,
    c.notes,
    cc.is_active,
    c.created_at,
    c.updated_at,
    cc.joined_at,
    cc.left_at
FROM
    company_clients cc
    INNER JOIN clients c ON c.id = cc.client_id
    INNER JOIN people p ON p.id = c.person_id
    INNER JOIN people_identifications pi ON pi.person_id = p.id
    INNER JOIN people_contacts pc ON pc.person_id = p.id
        AND pc.is_primary = TRUE
WHERE
    cc.company_id = sqlc.arg('CompanyID')
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL
    AND p.is_active = TRUE
    AND (sqlc.arg('Search')::text = ''
        OR pi.full_name ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR pi.cpf ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR pc.email ILIKE '%' || sqlc.arg('Search')::text || '%')
ORDER BY
    pi.full_name ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: GetClientByIDAndCompanyID :one
SELECT
    c.id,
    c.person_id,
    cc.company_id,
    pi.full_name,
    pi.short_name,
    pi.gender_identity,
    pi.marital_status,
    pi.birth_date,
    pi.cpf,
    pc.email,
    pc.phone,
    pc.cellphone,
    pc.has_whatsapp,
    c.client_since,
    c.notes,
    cc.is_active,
    c.created_at,
    c.updated_at,
    cc.joined_at,
    cc.left_at
FROM
    company_clients cc
    INNER JOIN clients c ON c.id = cc.client_id
    INNER JOIN people p ON p.id = c.person_id
    INNER JOIN people_identifications pi ON pi.person_id = p.id
    INNER JOIN people_contacts pc ON pc.person_id = p.id
        AND pc.is_primary = TRUE
WHERE
    cc.company_id = sqlc.arg('CompanyID')
    AND c.id = sqlc.arg('ID')
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL
    AND p.is_active = TRUE
LIMIT 1;

-- name: InsertClientPerson :one
INSERT INTO people(kind, is_active, has_system_user)
    VALUES ('client', TRUE, FALSE)
RETURNING
    *;

-- name: InsertClientIdentification :one
INSERT INTO people_identifications(person_id, full_name, short_name, gender_identity, marital_status, image_url, birth_date, cpf)
    VALUES (sqlc.arg('PersonID'), sqlc.arg('FullName'), sqlc.arg('ShortName'), sqlc.arg('GenderIdentity'), sqlc.arg('MaritalStatus'), sqlc.narg('ImageURL'), sqlc.arg('BirthDate'), sqlc.arg('CPF'))
RETURNING
    *;

-- name: InsertClientPrimaryContact :one
INSERT INTO people_contacts(person_id, email, phone, cellphone, has_whatsapp, is_primary)
    VALUES (sqlc.arg('PersonID'), sqlc.arg('Email'), sqlc.narg('Phone'), sqlc.arg('Cellphone'), sqlc.arg('HasWhatsapp'), TRUE)
RETURNING
    *;

-- name: InsertClientRecord :one
INSERT INTO clients(person_id, client_since, notes)
    VALUES (sqlc.arg('PersonID'), sqlc.narg('ClientSince'), sqlc.narg('Notes'))
RETURNING
    *;

-- name: CreateCompanyClient :one
INSERT INTO company_clients(company_id, client_id, is_active)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('ClientID'), TRUE)
RETURNING
    *;

-- name: UpdateClientIdentification :execrows
UPDATE
    people_identifications pi
SET
    full_name = COALESCE(sqlc.narg('FullName'), pi.full_name),
    short_name = COALESCE(sqlc.narg('ShortName'), pi.short_name),
    gender_identity = COALESCE(sqlc.narg('GenderIdentity'), pi.gender_identity),
    marital_status = COALESCE(sqlc.narg('MaritalStatus'), pi.marital_status),
    image_url = COALESCE(sqlc.narg('ImageURL'), pi.image_url),
    birth_date = COALESCE(sqlc.narg('BirthDate'), pi.birth_date),
    cpf = COALESCE(sqlc.narg('CPF'), pi.cpf),
    updated_at = now()
FROM
    clients c
    INNER JOIN company_clients cc ON cc.client_id = c.id
WHERE
    c.id = sqlc.arg('ID')
    AND cc.company_id = sqlc.arg('CompanyID')
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL
    AND pi.person_id = c.person_id;

-- name: UpdateClientPrimaryContact :execrows
UPDATE
    people_contacts pc
SET
    email = COALESCE(sqlc.narg('Email'), pc.email),
    phone = COALESCE(sqlc.narg('Phone'), pc.phone),
    cellphone = COALESCE(sqlc.narg('Cellphone'), pc.cellphone),
    has_whatsapp = COALESCE(sqlc.narg('HasWhatsapp'), pc.has_whatsapp),
    updated_at = now()
FROM
    clients c
    INNER JOIN company_clients cc ON cc.client_id = c.id
WHERE
    c.id = sqlc.arg('ID')
    AND cc.company_id = sqlc.arg('CompanyID')
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL
    AND pc.person_id = c.person_id
    AND pc.is_primary = TRUE;

-- name: UpdateClientRecord :execrows
UPDATE
    clients c
SET
    client_since = COALESCE(sqlc.narg('ClientSince'), c.client_since),
    notes = COALESCE(sqlc.narg('Notes'), c.notes),
    updated_at = now()
FROM
    company_clients cc
WHERE
    c.id = sqlc.arg('ID')
    AND cc.company_id = sqlc.arg('CompanyID')
    AND cc.client_id = c.id
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL;

-- name: DeactivateClient :execrows
UPDATE
    company_clients
SET
    is_active = FALSE,
    left_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND client_id = sqlc.arg('ClientID')
    AND is_active = TRUE;

