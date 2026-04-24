DELETE FROM pet_guardians pg
USING (
    SELECT
        pet_id,
        guardian_id,
        ROW_NUMBER() OVER (PARTITION BY pet_id ORDER BY created_at ASC, guardian_id ASC) AS row_num
    FROM
        pet_guardians
) duplicates
WHERE
    pg.pet_id = duplicates.pet_id
    AND pg.guardian_id = duplicates.guardian_id
    AND duplicates.row_num > 1;

ALTER TABLE pet_guardians
    DROP CONSTRAINT IF EXISTS pet_guardians_pkey;

ALTER TABLE pet_guardians
    ADD CONSTRAINT pet_guardians_pkey PRIMARY KEY (pet_id);

DROP INDEX IF EXISTS idx_pet_guardians_pet;
