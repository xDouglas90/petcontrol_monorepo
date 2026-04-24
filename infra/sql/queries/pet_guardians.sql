-- name: UpsertPetGuardian :execrows
INSERT INTO pet_guardians(pet_id, guardian_id)
    VALUES (sqlc.arg('PetID'), sqlc.arg('GuardianID'))
ON CONFLICT (pet_id) DO UPDATE SET
    guardian_id = EXCLUDED.guardian_id;

-- name: DeletePetGuardiansByGuardianID :execrows
DELETE FROM pet_guardians
WHERE guardian_id = sqlc.arg('GuardianID');

-- name: ListGuardianPetsByCompanyID :many
SELECT
    pg.pet_id,
    p.name,
    p.kind,
    p.size,
    pi.full_name AS owner_name
FROM
    pet_guardians pg
    INNER JOIN pets p ON p.id = pg.pet_id
    INNER JOIN clients c ON c.id = p.owner_id
    INNER JOIN company_clients cc ON cc.client_id = c.id
    INNER JOIN people_identifications pi ON pi.person_id = c.person_id
WHERE
    pg.guardian_id = sqlc.arg('GuardianID')
    AND cc.company_id = sqlc.arg('CompanyID')
    AND cc.is_active = TRUE
    AND c.deleted_at IS NULL
    AND p.deleted_at IS NULL
    AND p.is_active = TRUE
ORDER BY
    p.name ASC;
