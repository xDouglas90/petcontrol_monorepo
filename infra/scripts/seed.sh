#!/usr/bin/env sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required"
  exit 1
fi

db_url="${DATABASE_URL}"
network_arg=""

if [ -n "${DOCKER_NETWORK:-}" ]; then
  network_arg="--network ${DOCKER_NETWORK}"
else
  case "$(uname -s)" in
    Linux*)
      network_arg="--network host"
      ;;
    *)
      db_url=$(printf '%s' "${db_url}" | sed 's/@localhost:/@host.docker.internal:/g; s/@127\.0\.0\.1:/@host.docker.internal:/g')
      ;;
  esac
fi

docker run --rm \
  ${network_arg} \
  -i \
  postgres:18-alpine \
  psql "${db_url}" -v ON_ERROR_STOP=1 <<'SQL'
-- Modules
INSERT INTO modules (code, name, description, min_package)
VALUES
  ('SCH', 'Scheduling', 'Core scheduling module', 'starter'),
  ('CRM', 'Customer Management', 'Customers and relationship management', 'starter'),
  ('FIN', 'Finance', 'Cashflow and finance controls', 'basic')
ON CONFLICT (code) DO NOTHING;

-- Plan types
INSERT INTO plan_types (name, description)
SELECT 'Monthly', 'Default monthly billing cycle'
WHERE NOT EXISTS (SELECT 1 FROM plan_types WHERE name = 'Monthly' AND deleted_at IS NULL);

INSERT INTO plan_types (name, description)
SELECT 'Annual', 'Default annual billing cycle'
WHERE NOT EXISTS (SELECT 1 FROM plan_types WHERE name = 'Annual' AND deleted_at IS NULL);

-- Plans
WITH monthly_type AS (
  SELECT id FROM plan_types WHERE name = 'Monthly' AND deleted_at IS NULL ORDER BY created_at ASC LIMIT 1
)
INSERT INTO plans (plan_type_id, name, description, package, price, billing_cycle_days, max_users, is_active)
SELECT mt.id, 'Starter Monthly', 'Initial starter monthly plan', 'starter', 99.90, 30, 5, TRUE
FROM monthly_type mt
WHERE NOT EXISTS (
  SELECT 1 FROM plans p WHERE p.name = 'Starter Monthly' AND p.deleted_at IS NULL
);

-- Plan modules
WITH starter_plan AS (
  SELECT id FROM plans WHERE name = 'Starter Monthly' AND deleted_at IS NULL ORDER BY created_at ASC LIMIT 1
), starter_modules AS (
  SELECT id FROM modules WHERE code IN ('SCH', 'CRM')
)
INSERT INTO plan_modules (plan_id, module_id, is_active)
SELECT sp.id, sm.id, TRUE
FROM starter_plan sp
CROSS JOIN starter_modules sm
ON CONFLICT (plan_id, module_id) DO NOTHING;

-- Responsible person (required by companies.responsible_id)
INSERT INTO people (kind, is_active, has_system_user)
SELECT 'responsible', TRUE, FALSE
WHERE NOT EXISTS (
  SELECT
    1
  FROM
    companies
  WHERE
    slug = 'petcontrol-dev'
);

-- Dev company
WITH responsible AS (
  SELECT
    id
  FROM
    people
  WHERE
    kind = 'responsible'
  ORDER BY
    created_at DESC
  LIMIT 1
)
INSERT INTO companies (slug, name, fantasy_name, cnpj, responsible_id, active_package, is_active)
SELECT
  'petcontrol-dev',
  'PetControl Desenvolvimento LTDA',
  'PetControl Dev',
  '12345678000195',
  r.id,
  'starter',
  TRUE
FROM
  responsible r
WHERE NOT EXISTS (
  SELECT
    1
  FROM
    companies c
  WHERE
    c.slug = 'petcontrol-dev'
);

-- Root user (dev bootstrap)
INSERT INTO users (email, email_verified, email_verified_at, role, is_active)
SELECT 'root@petcontrol.local', TRUE, NOW(), 'root', TRUE
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'root@petcontrol.local');

-- Root auth profile (password: password123, requires password change)
INSERT INTO user_auth (user_id, password_hash, must_change_password)
SELECT u.id, '$2a$12$HAtO6l.iXD27nYmeaFjSEeeiYPo0TVPANJzxhUG/DvC5xzdAp2QrG', TRUE
FROM users u
WHERE u.email = 'root@petcontrol.local'
ON CONFLICT (user_id) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  must_change_password = EXCLUDED.must_change_password,
  updated_at = NOW();

-- Admin user compatible with web default credentials
INSERT INTO users (email, email_verified, email_verified_at, role, is_active)
SELECT 'admin@petcontrol.local', TRUE, NOW(), 'admin', TRUE
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@petcontrol.local');

-- Admin auth profile (password: password123)
INSERT INTO user_auth (user_id, password_hash, must_change_password)
SELECT u.id, '$2a$12$HAtO6l.iXD27nYmeaFjSEeeiYPo0TVPANJzxhUG/DvC5xzdAp2QrG', FALSE
FROM users u
WHERE u.email = 'admin@petcontrol.local'
ON CONFLICT (user_id) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  must_change_password = EXCLUDED.must_change_password,
  updated_at = NOW();

-- Active memberships for seeded users
WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), seeded_users AS (
  SELECT id, email FROM users WHERE email IN ('root@petcontrol.local', 'admin@petcontrol.local')
)
INSERT INTO company_users (company_id, user_id, kind, is_owner, is_active)
SELECT
  dc.id,
  su.id,
  CASE WHEN su.email = 'admin@petcontrol.local' THEN 'owner'::user_kind ELSE 'employee'::user_kind END,
  CASE WHEN su.email = 'admin@petcontrol.local' THEN TRUE ELSE FALSE END,
  TRUE
FROM dev_company dc
CROSS JOIN seeded_users su
WHERE NOT EXISTS (
  SELECT 1
  FROM company_users cu
  WHERE cu.company_id = dc.id AND cu.user_id = su.id
);

-- Active subscription for current seeded plan
WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), starter_plan AS (
  SELECT id, price, billing_cycle_days FROM plans WHERE name = 'Starter Monthly' AND deleted_at IS NULL ORDER BY created_at ASC LIMIT 1
)
INSERT INTO company_subscriptions (company_id, plan_id, started_at, expires_at, is_active, price_paid, notes)
SELECT
  dc.id,
  sp.id,
  NOW() - INTERVAL '1 day',
  NOW() + make_interval(days => sp.billing_cycle_days),
  TRUE,
  sp.price,
  'Seeded development subscription'
FROM dev_company dc
CROSS JOIN starter_plan sp
WHERE NOT EXISTS (
  SELECT 1
  FROM company_subscriptions cs
  WHERE cs.company_id = dc.id
    AND cs.plan_id = sp.id
    AND cs.is_active = TRUE
    AND cs.canceled_at IS NULL
    AND cs.expires_at > NOW()
);

-- Active company modules for the seeded tenant
WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), starter_modules AS (
  SELECT id FROM modules WHERE code IN ('SCH', 'CRM')
)
INSERT INTO company_modules (company_id, module_id, is_active)
SELECT dc.id, sm.id, TRUE
FROM dev_company dc
CROSS JOIN starter_modules sm
ON CONFLICT (company_id, module_id) DO UPDATE SET
  is_active = EXCLUDED.is_active,
  updated_at = NOW();

-- Seeded client person for operational flows
INSERT INTO people (kind, is_active, has_system_user)
SELECT 'client', TRUE, FALSE
WHERE NOT EXISTS (
  SELECT 1
  FROM people_identifications pi
  WHERE pi.cpf = '12345678901'
);

WITH seeded_person AS (
  SELECT p.id
  FROM people p
  LEFT JOIN people_identifications pi ON pi.person_id = p.id
  WHERE pi.cpf = '12345678901'
     OR (
       pi.person_id IS NULL
       AND p.kind = 'client'
       AND p.has_system_user = FALSE
     )
  ORDER BY p.created_at DESC
  LIMIT 1
)
INSERT INTO people_identifications (
  person_id,
  full_name,
  short_name,
  gender_identity,
  marital_status,
  birth_date,
  cpf
)
SELECT
  sp.id,
  'Maria Silva',
  'Maria',
  'woman_cisgender',
  'single',
  DATE '1992-06-15',
  '12345678901'
FROM seeded_person sp
WHERE NOT EXISTS (
  SELECT 1 FROM people_identifications WHERE cpf = '12345678901'
);

WITH seeded_person AS (
  SELECT pi.person_id AS id
  FROM people_identifications pi
  WHERE pi.cpf = '12345678901'
  LIMIT 1
)
INSERT INTO people_contacts (
  person_id,
  email,
  phone,
  cellphone,
  has_whatsapp,
  is_primary
)
SELECT
  sp.id,
  'maria.silva@petcontrol.local',
  '+551130000000',
  '+5511999990001',
  TRUE,
  TRUE
FROM seeded_person sp
WHERE NOT EXISTS (
  SELECT 1
  FROM people_contacts pc
  WHERE pc.person_id = sp.id
    AND pc.email = 'maria.silva@petcontrol.local'
);

WITH seeded_person AS (
  SELECT pi.person_id AS id
  FROM people_identifications pi
  WHERE pi.cpf = '12345678901'
  LIMIT 1
)
INSERT INTO clients (
  person_id,
  client_since,
  notes
)
SELECT
  sp.id,
  CURRENT_DATE - INTERVAL '45 days',
  'Cliente seedado para fluxos operacionais locais'
FROM seeded_person sp
WHERE NOT EXISTS (
  SELECT 1
  FROM clients c
  WHERE c.person_id = sp.id
    AND c.deleted_at IS NULL
);

WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), seeded_client AS (
  SELECT c.id
  FROM clients c
  INNER JOIN people_identifications pi ON pi.person_id = c.person_id
  WHERE pi.cpf = '12345678901'
    AND c.deleted_at IS NULL
  LIMIT 1
)
INSERT INTO company_clients (
  company_id,
  client_id,
  is_active
)
SELECT
  dc.id,
  sc.id,
  TRUE
FROM dev_company dc
CROSS JOIN seeded_client sc
WHERE NOT EXISTS (
  SELECT 1
  FROM company_clients cc
  WHERE cc.company_id = dc.id
    AND cc.client_id = sc.id
);

WITH seeded_client AS (
  SELECT c.id
  FROM clients c
  INNER JOIN people_identifications pi ON pi.person_id = c.person_id
  WHERE pi.cpf = '12345678901'
    AND c.deleted_at IS NULL
  LIMIT 1
)
INSERT INTO pets (
  name,
  size,
  kind,
  temperament,
  birth_date,
  owner_id,
  is_active,
  notes
)
SELECT
  'Thor',
  'medium',
  'dog',
  'playful',
  DATE '2021-08-20',
  sc.id,
  TRUE,
  'Pet seedado para validar fluxos de agendamento'
FROM seeded_client sc
WHERE NOT EXISTS (
  SELECT 1
  FROM pets p
  WHERE p.owner_id = sc.id
    AND p.name = 'Thor'
    AND p.deleted_at IS NULL
);

INSERT INTO service_types (name, description)
SELECT 'Banho', 'Serviços de banho e higienização'
WHERE NOT EXISTS (
  SELECT 1
  FROM service_types st
  WHERE st.name = 'Banho'
    AND st.deleted_at IS NULL
);

WITH banho_type AS (
  SELECT id
  FROM service_types
  WHERE name = 'Banho'
    AND deleted_at IS NULL
  ORDER BY created_at ASC
  LIMIT 1
)
INSERT INTO services (
  type_id,
  title,
  description,
  notes,
  price,
  discount_rate,
  is_active
)
SELECT
  bt.id,
  'Banho completo',
  'Banho com secagem, perfume e escovação',
  'Serviço seedado para o catálogo local',
  89.90,
  0.00,
  TRUE
FROM banho_type bt
WHERE NOT EXISTS (
  SELECT 1
  FROM services s
  WHERE s.title = 'Banho completo'
    AND s.deleted_at IS NULL
);

WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), seeded_service AS (
  SELECT id
  FROM services
  WHERE title = 'Banho completo'
    AND deleted_at IS NULL
  ORDER BY created_at ASC
  LIMIT 1
)
INSERT INTO company_services (
  company_id,
  service_id,
  is_active
)
SELECT
  dc.id,
  ss.id,
  TRUE
FROM dev_company dc
CROSS JOIN seeded_service ss
WHERE NOT EXISTS (
  SELECT 1
  FROM company_services cs
  WHERE cs.company_id = dc.id
    AND cs.service_id = ss.id
);

WITH dev_company AS (
  SELECT id FROM companies WHERE slug = 'petcontrol-dev' LIMIT 1
), seeded_client AS (
  SELECT c.id
  FROM clients c
  INNER JOIN people_identifications pi ON pi.person_id = c.person_id
  WHERE pi.cpf = '12345678901'
    AND c.deleted_at IS NULL
  LIMIT 1
), seeded_pet AS (
  SELECT p.id
  FROM pets p
  INNER JOIN seeded_client sc ON sc.id = p.owner_id
  WHERE p.name = 'Thor'
    AND p.deleted_at IS NULL
  LIMIT 1
), seeded_admin AS (
  SELECT id
  FROM users
  WHERE email = 'admin@petcontrol.local'
  LIMIT 1
)
INSERT INTO schedules (
  company_id,
  client_id,
  pet_id,
  scheduled_at,
  estimated_end,
  notes,
  created_by
)
SELECT
  dc.id,
  sc.id,
  sp.id,
  TIMESTAMPTZ '2026-04-15 14:00:00+00',
  TIMESTAMPTZ '2026-04-15 15:00:00+00',
  'Agendamento seedado para fluxo local do módulo schedules',
  sa.id
FROM dev_company dc
CROSS JOIN seeded_client sc
CROSS JOIN seeded_pet sp
CROSS JOIN seeded_admin sa
WHERE NOT EXISTS (
  SELECT 1
  FROM schedules s
  WHERE s.company_id = dc.id
    AND s.client_id = sc.id
    AND s.pet_id = sp.id
    AND s.scheduled_at = TIMESTAMPTZ '2026-04-15 14:00:00+00'
    AND s.deleted_at IS NULL
);

WITH seeded_schedule AS (
  SELECT s.id
  FROM schedules s
  INNER JOIN companies c ON c.id = s.company_id
  INNER JOIN clients cl ON cl.id = s.client_id
  INNER JOIN people_identifications pi ON pi.person_id = cl.person_id
  INNER JOIN pets p ON p.id = s.pet_id
  WHERE c.slug = 'petcontrol-dev'
    AND pi.cpf = '12345678901'
    AND p.name = 'Thor'
    AND s.scheduled_at = TIMESTAMPTZ '2026-04-15 14:00:00+00'
    AND s.deleted_at IS NULL
  LIMIT 1
), seeded_admin AS (
  SELECT id
  FROM users
  WHERE email = 'admin@petcontrol.local'
  LIMIT 1
)
INSERT INTO schedule_status_history (
  schedule_id,
  status,
  changed_by,
  notes
)
SELECT
  ss.id,
  'confirmed',
  sa.id,
  'Status inicial do agendamento seedado'
FROM seeded_schedule ss
CROSS JOIN seeded_admin sa
WHERE NOT EXISTS (
  SELECT 1
  FROM schedule_status_history ssh
  WHERE ssh.schedule_id = ss.id
    AND ssh.status = 'confirmed'
);
SQL
