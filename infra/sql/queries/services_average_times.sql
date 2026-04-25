-- name: InsertServiceAverageTime :one
INSERT INTO services_average_times(service_id, sub_service_id, pet_size, pet_kind, pet_temperament, average_time_minutes)
    VALUES (sqlc.arg('ServiceID'), sqlc.narg('SubServiceID'), sqlc.arg('PetSize'), sqlc.arg('PetKind'), sqlc.arg('PetTemperament'), sqlc.arg('AverageTimeMinutes'))
RETURNING
    *;

-- name: ListServiceAverageTimesByServiceID :many
SELECT
    id,
    service_id,
    sub_service_id,
    pet_size,
    pet_kind,
    pet_temperament,
    average_time_minutes,
    created_at,
    updated_at
FROM
    services_average_times
WHERE
    service_id = sqlc.arg('ServiceID')
ORDER BY
    sub_service_id NULLS FIRST,
    pet_size ASC,
    pet_kind ASC,
    pet_temperament ASC;

-- name: DeleteServiceAverageTimesByServiceID :execrows
DELETE FROM services_average_times
WHERE service_id = sqlc.arg('ServiceID');
