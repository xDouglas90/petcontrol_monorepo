CREATE TABLE IF NOT EXISTS services_average_times(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_id uuid REFERENCES services(id) ON DELETE CASCADE,
  sub_service_id uuid REFERENCES sub_services(id) ON DELETE CASCADE,
  pet_size pet_size NOT NULL,
  pet_kind pet_kind NOT NULL,
  pet_temperament pet_temperament NOT NULL,
  average_time_minutes smallint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz,
  CONSTRAINT chk_service_or_sub_service CHECK (service_id IS NOT NULL OR sub_service_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_services_average_times_service ON services_average_times(service_id);

CREATE INDEX IF NOT EXISTS idx_services_average_times_sub_service ON services_average_times(sub_service_id);
