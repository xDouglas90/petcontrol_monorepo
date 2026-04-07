#!/usr/bin/env sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required"
  exit 1
fi

docker run --rm \
  --network host \
  -i \
  postgres:16-alpine \
  psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 <<'SQL'
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

-- Root user (dev bootstrap)
INSERT INTO users (email, email_verified, email_verified_at, role, kind, is_active)
SELECT 'root@petcontrol.local', TRUE, NOW(), 'root', 'internal', TRUE
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'root@petcontrol.local');

-- Root auth profile (bcrypt hash placeholder for local development)
INSERT INTO user_auth (user_id, password_hash, must_change_password)
SELECT u.id, '$2a$12$HP0VOGM.j2Gm6rXtAdo2XOR4fN1fMCTM4xCEf7hL1g9lhH57jXkju', TRUE
FROM users u
WHERE u.email = 'root@petcontrol.local'
  AND NOT EXISTS (SELECT 1 FROM user_auth ua WHERE ua.user_id = u.id);
SQL