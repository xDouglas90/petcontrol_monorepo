-- name: InsertEmployeeDocuments :execrows
INSERT INTO employee_documents(person_id, rg, issuing_body, issuing_date, ctps, ctps_series, ctps_state, pis, voter_registration, vote_zone, vote_section, military_certificate, military_series, military_category, has_special_needs, has_children, children_qty, has_children_under_18, has_family_special_needs, graduation, has_cnh, cnh_type, cnh_number, cnh_expiration_date)
    VALUES (sqlc.arg('PersonID'), sqlc.arg('RG'), sqlc.arg('IssuingBody'), sqlc.arg('IssuingDate'), sqlc.arg('CTPS'), sqlc.arg('CTPSSeries'), sqlc.arg('CTPSState'), sqlc.arg('PIS'), sqlc.arg('VoterRegistration'), sqlc.arg('VoteZone'), sqlc.arg('VoteSection'), sqlc.arg('MilitaryCertificate'), sqlc.arg('MilitarySeries'), sqlc.arg('MilitaryCategory'), sqlc.narg('HasSpecialNeeds'), sqlc.narg('HasChildren'), sqlc.arg('ChildrenQty'), sqlc.narg('HasChildrenUnder18'), sqlc.narg('HasFamilySpecialNeeds'), sqlc.arg('Graduation'), sqlc.narg('HasCNH'), sqlc.arg('CNHType'), sqlc.arg('CNHNumber'), sqlc.arg('CNHExpirationDate'));

-- name: UpdateEmployeeDocuments :execrows
UPDATE
    employee_documents
SET
    rg = COALESCE(sqlc.narg('RG'), rg),
    issuing_body = COALESCE(sqlc.narg('IssuingBody'), issuing_body),
    issuing_date = COALESCE(sqlc.narg('IssuingDate'), issuing_date),
    ctps = COALESCE(sqlc.narg('CTPS'), ctps),
    ctps_series = COALESCE(sqlc.narg('CTPSSeries'), ctps_series),
    ctps_state = COALESCE(sqlc.narg('CTPSState'), ctps_state),
    pis = COALESCE(sqlc.narg('PIS'), pis),
    voter_registration = COALESCE(sqlc.narg('VoterRegistration'), voter_registration),
    vote_zone = COALESCE(sqlc.narg('VoteZone'), vote_zone),
    vote_section = COALESCE(sqlc.narg('VoteSection'), vote_section),
    military_certificate = COALESCE(sqlc.narg('MilitaryCertificate'), military_certificate),
    military_series = COALESCE(sqlc.narg('MilitarySeries'), military_series),
    military_category = COALESCE(sqlc.narg('MilitaryCategory'), military_category),
    has_special_needs = COALESCE(sqlc.narg('HasSpecialNeeds'), has_special_needs),
    has_children = COALESCE(sqlc.narg('HasChildren'), has_children),
    children_qty = COALESCE(sqlc.narg('ChildrenQty'), children_qty),
    has_children_under_18 = COALESCE(sqlc.narg('HasChildrenUnder18'), has_children_under_18),
    has_family_special_needs = COALESCE(sqlc.narg('HasFamilySpecialNeeds'), has_family_special_needs),
    graduation = COALESCE(sqlc.narg('Graduation'), graduation),
    has_cnh = COALESCE(sqlc.narg('HasCNH'), has_cnh),
    cnh_type = COALESCE(sqlc.narg('CNHType'), cnh_type),
    cnh_number = COALESCE(sqlc.narg('CNHNumber'), cnh_number),
    cnh_expiration_date = COALESCE(sqlc.narg('CNHExpirationDate'), cnh_expiration_date),
    updated_at = now()
WHERE
    person_id = sqlc.arg('PersonID');

-- name: GetEmployeeDocuments :one
SELECT
    ed.person_id,
    ed.rg,
    ed.issuing_body,
    ed.issuing_date,
    ed.ctps,
    ed.ctps_series,
    ed.ctps_state,
    ed.pis,
    ed.voter_registration,
    ed.vote_zone,
    ed.vote_section,
    ed.military_certificate,
    ed.military_series,
    ed.military_category,
    ed.has_special_needs,
    ed.has_children,
    ed.children_qty,
    ed.has_children_under_18,
    ed.has_family_special_needs,
    ed.graduation,
    ed.has_cnh,
    ed.cnh_type,
    ed.cnh_number,
    ed.cnh_expiration_date,
    ed.created_at,
    ed.updated_at
FROM
    employee_documents ed
WHERE
    ed.person_id = sqlc.arg('PersonID');

