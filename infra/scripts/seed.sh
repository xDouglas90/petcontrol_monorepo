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
INSERT INTO users (email, email_verified, email_verified_at, role, kind, is_active)
SELECT 'root@petcontrol.local', TRUE, NOW(), 'root', 'internal', TRUE
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
INSERT INTO users (email, email_verified, email_verified_at, role, kind, is_active)
SELECT 'admin@petcontrol.local', TRUE, NOW(), 'admin', 'owner', TRUE
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
INSERT INTO company_users (company_id, user_id, is_owner, is_active)
SELECT
  dc.id,
  su.id,
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
), active_subscription AS (
  SELECT id FROM company_subscriptions
  WHERE company_id = (SELECT id FROM dev_company)
    AND is_active = TRUE
    AND canceled_at IS NULL
    AND expires_at > NOW()
  ORDER BY started_at DESC
  LIMIT 1
), starter_modules AS (
  SELECT id FROM modules WHERE code IN ('SCH', 'CRM')
)
INSERT INTO company_modules (company_id, module_id, subscription_id, granted_manually, is_active)
SELECT dc.id, sm.id, s.id, FALSE, TRUE
FROM dev_company dc
JOIN active_subscription s ON TRUE
CROSS JOIN starter_modules sm
ON CONFLICT (company_id, module_id) DO UPDATE SET
  subscription_id = EXCLUDED.subscription_id,
  granted_manually = EXCLUDED.granted_manually,
  is_active = EXCLUDED.is_active,
  updated_at = NOW();
SQL
