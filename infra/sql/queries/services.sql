-- name: ListServicesByCompanyID :many
SELECT
    COUNT(*) OVER () AS total_count,
    s.id,
    s.type_id,
    st.name AS type_name,
    s.title,
    s.description,
    s.notes,
    s.price,
    s.discount_rate,
    s.image_url,
    cs.is_active,
    COALESCE(sub_services.count, 0)::bigint AS sub_services_count,
    COALESCE(average_times.count, 0)::bigint AS average_times_count
FROM
    company_services cs
    INNER JOIN services s ON s.id = cs.service_id
    INNER JOIN service_types st ON st.id = s.type_id
    LEFT JOIN LATERAL (
        SELECT
            COUNT(*) AS count
        FROM
            sub_services ss
        WHERE
            ss.service_id = s.id
            AND ss.deleted_at IS NULL) sub_services ON TRUE
    LEFT JOIN LATERAL (
        SELECT
            COUNT(*) AS count
        FROM
            services_average_times sat
        WHERE
            sat.service_id = s.id) average_times ON TRUE
WHERE
    cs.company_id = sqlc.arg('CompanyID')
    AND (sqlc.narg('IsActive')::boolean IS NULL
        OR cs.is_active = sqlc.narg('IsActive')::boolean)
    AND s.deleted_at IS NULL
    AND st.deleted_at IS NULL
    AND (sqlc.arg('Search')::text = ''
        OR s.title ILIKE '%' || sqlc.arg('Search')::text || '%'
        OR s.description ILIKE '%' || sqlc.arg('Search')::text || '%')
    AND (sqlc.arg('TypeName')::text = ''
        OR st.name = sqlc.arg('TypeName')::text)
    AND (sqlc.narg('MinPrice')::numeric IS NULL
        OR s.price >= sqlc.narg('MinPrice')::numeric)
    AND (sqlc.narg('MaxPrice')::numeric IS NULL
        OR s.price <= sqlc.narg('MaxPrice')::numeric)
ORDER BY
    s.title ASC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

-- name: GetServiceByIDAndCompanyID :one
SELECT
    s.id,
    s.type_id,
    st.name AS type_name,
    s.title,
    s.description,
    s.notes,
    s.price,
    s.discount_rate,
    s.image_url,
    cs.is_active,
    COALESCE(sub_services.count, 0)::bigint AS sub_services_count,
    COALESCE(average_times.count, 0)::bigint AS average_times_count
FROM
    company_services cs
    INNER JOIN services s ON s.id = cs.service_id
    INNER JOIN service_types st ON st.id = s.type_id
    LEFT JOIN LATERAL (
        SELECT
            COUNT(*) AS count
        FROM
            sub_services ss
        WHERE
            ss.service_id = s.id
            AND ss.deleted_at IS NULL) sub_services ON TRUE
    LEFT JOIN LATERAL (
        SELECT
            COUNT(*) AS count
        FROM
            services_average_times sat
        WHERE
            sat.service_id = s.id) average_times ON TRUE
WHERE
    cs.company_id = sqlc.arg('CompanyID')
    AND s.id = sqlc.arg('ID')
    AND s.deleted_at IS NULL
    AND st.deleted_at IS NULL
LIMIT 1;

-- name: FindServiceTypeByName :one
SELECT
    id,
    name,
    description,
    created_at,
    updated_at,
    deleted_at
FROM
    service_types
WHERE
    lower(name) = lower(sqlc.arg('Name'))
    AND deleted_at IS NULL
LIMIT 1;

-- name: CreateServiceType :one
INSERT INTO service_types(name, description)
    VALUES (sqlc.arg('Name'), sqlc.narg('Description'))
RETURNING
    *;

-- name: CreateService :one
INSERT INTO services(type_id, title, description, notes, price, discount_rate, image_url, is_active)
    VALUES (sqlc.arg('TypeID'), sqlc.arg('Title'), sqlc.arg('Description'), sqlc.narg('Notes'), sqlc.arg('Price'), sqlc.arg('DiscountRate'), sqlc.narg('ImageURL'), sqlc.arg('IsActive'))
RETURNING
    *;

-- name: CreateCompanyService :one
INSERT INTO company_services(company_id, service_id, is_active)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('ServiceID'), TRUE)
RETURNING
    *;

-- name: UpdateServiceByIDAndCompanyID :one
UPDATE
    services s
SET
    type_id = COALESCE(sqlc.narg('TypeID'), s.type_id),
    title = COALESCE(sqlc.narg('Title'), s.title),
    description = COALESCE(sqlc.narg('Description'), s.description),
    notes = COALESCE(sqlc.narg('Notes'), s.notes),
    price = COALESCE(sqlc.narg('Price'), s.price),
    discount_rate = COALESCE(sqlc.narg('DiscountRate'), s.discount_rate),
    image_url = COALESCE(sqlc.narg('ImageURL'), s.image_url),
    is_active = COALESCE(sqlc.narg('IsActive'), s.is_active),
    updated_at = now()
FROM
    company_services cs
WHERE
    s.id = cs.service_id
    AND s.id = sqlc.arg('ID')
    AND cs.company_id = sqlc.arg('CompanyID')
    AND cs.is_active = TRUE
    AND s.deleted_at IS NULL
RETURNING
    s.*;

-- name: DeactivateCompanyService :one
UPDATE
    company_services
SET
    is_active = FALSE,
    updated_at = now()
WHERE
    company_id = sqlc.arg('CompanyID')
    AND service_id = sqlc.arg('ServiceID')
    AND is_active = TRUE
RETURNING
    *;

-- name: ValidateServiceByIDAndCompanyID :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            company_services cs
            INNER JOIN services s ON s.id = cs.service_id
        WHERE
            cs.company_id = sqlc.arg('CompanyID')
            AND cs.service_id = sqlc.arg('ServiceID')
            AND cs.is_active = TRUE
            AND s.deleted_at IS NULL) AS is_valid;
