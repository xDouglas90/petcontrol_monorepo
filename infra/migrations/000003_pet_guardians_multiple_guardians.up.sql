ALTER TABLE pet_guardians
    DROP CONSTRAINT IF EXISTS pet_guardians_pkey;

ALTER TABLE pet_guardians
    ADD CONSTRAINT pet_guardians_pkey PRIMARY KEY (pet_id, guardian_id);

CREATE INDEX IF NOT EXISTS idx_pet_guardians_pet ON pet_guardians(pet_id);
