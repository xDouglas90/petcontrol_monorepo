-- name: InsertPersonContacts :execrows
INSERT INTO people_contacts(person_id, email, phone, cellphone, has_whatsapp, instagram_user, emergency_contact, emergency_phone, is_primary)
    VALUES (sqlc.arg('PersonID'), sqlc.arg('Email'), sqlc.arg('Phone'), sqlc.arg('Cellphone'), sqlc.arg('HasWhatsapp'), sqlc.arg('InstagramUser'), sqlc.arg('EmergencyContact'), sqlc.arg('EmergencyPhone'), sqlc.arg('IsPrimary'));

-- name: UpdatePersonContacts :execrows
UPDATE
    people_contacts
SET
    email = COALESCE(sqlc.arg('Email'), email),
    phone = COALESCE(sqlc.arg('Phone'), phone),
    cellphone = COALESCE(sqlc.arg('Cellphone'), cellphone),
    has_whatsapp = COALESCE(sqlc.arg('HasWhatsapp'), has_whatsapp),
    instagram_user = COALESCE(sqlc.arg('InstagramUser'), instagram_user),
    emergency_contact = COALESCE(sqlc.arg('EmergencyContact'), emergency_contact),
    emergency_phone = COALESCE(sqlc.arg('EmergencyPhone'), emergency_phone),
    is_primary = COALESCE(sqlc.arg('IsPrimary'), is_primary),
    updated_at = now()
WHERE
    person_id = sqlc.arg('PersonID');

-- name: GetPersonContacts :one
SELECT
    pc.person_id,
    pc.email,
    pc.phone,
    pc.cellphone,
    pc.has_whatsapp,
    pc.instagram_user,
    pc.emergency_contact,
    pc.emergency_phone,
    pc.is_primary,
    pc.created_at,
    pc.updated_at
FROM
    people_contacts pc
WHERE
    pc.person_id = sqlc.arg('PersonID');

