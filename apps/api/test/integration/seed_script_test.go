package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeedScriptCreatesOperationalSupportData(t *testing.T) {
	setup := SetupPostgresWithMigrations(t)
	scriptPath := filepath.Join(resolveRepoRoot(t), "infra", "scripts", "seed.sh")

	cmd := exec.Command("sh", scriptPath)
	cmd.Env = append(os.Environ(), "DATABASE_URL="+setup.ConnString)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "seed script failed: %s", string(output))

	cmd = exec.Command("sh", scriptPath)
	cmd.Env = append(os.Environ(), "DATABASE_URL="+setup.ConnString)
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "seed script should be idempotent: %s", string(output))

	var (
		clientCount                   int
		configCount                   int
		petCount                      int
		serviceCount                  int
		subServiceCount               int
		averageTimeCount              int
		confirmedStatusCount          int
		dashboardSeedCount            int
		permissionCount               int
		rootPermissionCount           int
		adminPermissionCount          int
		systemPermissionCount         int
		systemSettingsCount           int
		systemCompanyEditCount        int
		settingsModuleCount           int
		settingsModulePermissionCount int
		starterCompanyModuleCount     int
	)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM company_clients cc
		INNER JOIN companies c ON c.id = cc.company_id
		INNER JOIN clients cl ON cl.id = cc.client_id
		INNER JOIN people_identifications pi ON pi.person_id = cl.person_id
		WHERE c.slug = 'petcontrol-dev'
		  AND pi.cpf = '12345678901'
		  AND cc.is_active = TRUE
	`).Scan(&clientCount)
	require.NoError(t, err)
	require.Equal(t, 1, clientCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM permissions
	`).Scan(&permissionCount)
	require.NoError(t, err)
	require.Equal(t, 123, permissionCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM modules
		WHERE code IN ('CFG', 'UCR', 'CLI', 'SCH', 'SVC', 'PET', 'RPT', 'PRD', 'DLV', 'INV', 'PDC', 'PHO', 'CHT', 'NTF', 'FIN', 'SUP', 'EUA')
		  AND deleted_at IS NULL
	`).Scan(&settingsModuleCount)
	require.NoError(t, err)
	require.Equal(t, 17, settingsModuleCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM module_permissions mp
		INNER JOIN modules m ON m.id = mp.module_id
		INNER JOIN permissions p ON p.id = mp.permission_id
		WHERE (m.code, p.code) IN (
		  ('CFG', 'company_settings:edit'),
		  ('CFG', 'plan_settings:edit'),
		  ('UCR', 'users:view'),
		  ('CLI', 'clients:view'),
		  ('SCH', 'schedules:view'),
		  ('SVC', 'services:view'),
		  ('PET', 'pets:view'),
		  ('RPT', 'reports:view'),
		  ('PRD', 'products:view'),
		  ('DLV', 'pickup_delivery:view'),
		  ('INV', 'stock:view'),
		  ('PDC', 'daycare:view'),
		  ('PHO', 'hotel:view'),
		  ('CHT', 'chat:view'),
		  ('NTF', 'notifications:view'),
		  ('FIN', 'finances:view'),
		  ('SUP', 'suppliers:view'),
		  ('EUA', 'external_access:view'),
		  ('AUD', 'logs:view')
		)
	`).Scan(&settingsModulePermissionCount)
	require.NoError(t, err)
	require.Equal(t, 19, settingsModulePermissionCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM company_modules cm
		INNER JOIN companies c ON c.id = cm.company_id
		INNER JOIN modules m ON m.id = cm.module_id
		WHERE c.slug = 'petcontrol-dev'
		  AND cm.is_active = TRUE
		  AND m.code IN ('CFG', 'UCR', 'CLI', 'SCH', 'SVC', 'PET')
	`).Scan(&starterCompanyModuleCount)
	require.NoError(t, err)
	require.Equal(t, 6, starterCompanyModuleCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM user_permissions up
		INNER JOIN users u ON u.id = up.user_id
		WHERE u.email = 'root@petcontrol.local'
		  AND up.is_active = TRUE
		  AND up.revoked_at IS NULL
	`).Scan(&rootPermissionCount)
	require.NoError(t, err)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM permissions p
		WHERE 'root'::user_role_type = ANY(p.default_roles)
	`).Scan(&permissionCount)
	require.NoError(t, err)
	require.Equal(t, permissionCount, rootPermissionCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM user_permissions up
		INNER JOIN users u ON u.id = up.user_id
		WHERE u.email = 'admin@petcontrol.local'
		  AND up.is_active = TRUE
		  AND up.revoked_at IS NULL
	`).Scan(&adminPermissionCount)
	require.NoError(t, err)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM permissions p
		WHERE 'admin'::user_role_type = ANY(p.default_roles)
	`).Scan(&permissionCount)
	require.NoError(t, err)
	require.Equal(t, permissionCount, adminPermissionCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM user_permissions up
		INNER JOIN users u ON u.id = up.user_id
		WHERE u.email = 'system@petcontrol.local'
		  AND up.is_active = TRUE
		  AND up.revoked_at IS NULL
	`).Scan(&systemPermissionCount)
	require.NoError(t, err)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM permissions p
		WHERE 'system'::user_role_type = ANY(p.default_roles)
	`).Scan(&permissionCount)
	require.NoError(t, err)
	require.Equal(t, permissionCount, systemPermissionCount)
	require.Equal(t, 2, systemPermissionCount)

	rows, err := setup.Pool.Query(setup.Ctx, `
		SELECT p.code
		FROM permissions p
		WHERE 'system'::user_role_type = ANY(p.default_roles)
		ORDER BY p.code
	`)
	require.NoError(t, err)
	defer rows.Close()
	systemPermissionCodes := make([]string, 0, 2)
	for rows.Next() {
		var code string
		require.NoError(t, rows.Scan(&code))
		systemPermissionCodes = append(systemPermissionCodes, code)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []string{"services:view", "users:view"}, systemPermissionCodes)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM user_permissions up
		INNER JOIN users u ON u.id = up.user_id
		INNER JOIN permissions p ON p.id = up.permission_id
		WHERE u.email = 'system@petcontrol.local'
		  AND p.code = 'plan_settings:edit'
		  AND up.is_active = TRUE
		  AND up.revoked_at IS NULL
	`).Scan(&systemSettingsCount)
	require.NoError(t, err)
	require.Equal(t, 0, systemSettingsCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM user_permissions up
		INNER JOIN users u ON u.id = up.user_id
		INNER JOIN permissions p ON p.id = up.permission_id
		WHERE u.email = 'system@petcontrol.local'
		  AND p.code = 'company_settings:edit'
		  AND up.is_active = TRUE
		  AND up.revoked_at IS NULL
	`).Scan(&systemCompanyEditCount)
	require.NoError(t, err)
	require.Equal(t, 0, systemCompanyEditCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM company_system_configs csc
		INNER JOIN companies c ON c.id = csc.company_id
		WHERE c.slug = 'petcontrol-dev'
		  AND csc.schedule_init_time = TIME '08:00'
		  AND csc.schedule_end_time = TIME '18:00'
		  AND csc.min_schedules_per_day = 4
		  AND csc.max_schedules_per_day = 18
	`).Scan(&configCount)
	require.NoError(t, err)
	require.Equal(t, 1, configCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM pets p
		INNER JOIN clients cl ON cl.id = p.owner_id
		INNER JOIN people_identifications pi ON pi.person_id = cl.person_id
		WHERE pi.cpf = '12345678901'
		  AND p.name = 'Thor'
		  AND p.is_active = TRUE
		  AND p.deleted_at IS NULL
	`).Scan(&petCount)
	require.NoError(t, err)
	require.Equal(t, 1, petCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(DISTINCT s.id)
		FROM company_services cs
		INNER JOIN companies c ON c.id = cs.company_id
		INNER JOIN services s ON s.id = cs.service_id
		WHERE c.slug = 'petcontrol-dev'
		  AND s.title IN ('Banho completo', 'Tosa higiênica')
		  AND cs.is_active = TRUE
		  AND s.deleted_at IS NULL
	`).Scan(&serviceCount)
	require.NoError(t, err)
	require.Equal(t, 2, serviceCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(DISTINCT ss.id)
		FROM sub_services ss
		INNER JOIN services s ON s.id = ss.service_id
		INNER JOIN company_services cs ON cs.service_id = s.id
		INNER JOIN companies c ON c.id = cs.company_id
		WHERE c.slug = 'petcontrol-dev'
		  AND s.title IN ('Banho completo', 'Tosa higiênica')
		  AND ss.deleted_at IS NULL
	`).Scan(&subServiceCount)
	require.NoError(t, err)
	require.Equal(t, 4, subServiceCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(DISTINCT sat.id)
		FROM services_average_times sat
		INNER JOIN sub_services ss ON ss.id = sat.sub_service_id
		INNER JOIN services s ON s.id = ss.service_id
		INNER JOIN company_services cs ON cs.service_id = s.id
		INNER JOIN companies c ON c.id = cs.company_id
		WHERE c.slug = 'petcontrol-dev'
		  AND s.title IN ('Banho completo', 'Tosa higiênica')
		  AND ss.deleted_at IS NULL
	`).Scan(&averageTimeCount)
	require.NoError(t, err)
	require.Equal(t, 8, averageTimeCount)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM schedules s
		INNER JOIN companies c ON c.id = s.company_id
		INNER JOIN clients cl ON cl.id = s.client_id
		INNER JOIN people_identifications pi ON pi.person_id = cl.person_id
		INNER JOIN pets p ON p.id = s.pet_id
		INNER JOIN schedule_status_history ssh ON ssh.schedule_id = s.id
		WHERE c.slug = 'petcontrol-dev'
		  AND pi.cpf = '12345678901'
		  AND p.name = 'Thor'
		  AND ssh.status = 'confirmed'
		  AND s.deleted_at IS NULL
	`).Scan(&confirmedStatusCount)
	require.NoError(t, err)
	require.GreaterOrEqual(t, confirmedStatusCount, 1)

	err = setup.Pool.QueryRow(setup.Ctx, `
		SELECT COUNT(*)
		FROM schedules s
		INNER JOIN companies c ON c.id = s.company_id
		WHERE c.slug = 'petcontrol-dev'
		  AND s.notes LIKE 'Dashboard seed:%'
		  AND s.deleted_at IS NULL
	`).Scan(&dashboardSeedCount)
	require.NoError(t, err)
	require.GreaterOrEqual(t, dashboardSeedCount, 10)
}
