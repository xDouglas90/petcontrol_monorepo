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
		clientCount          int
		configCount          int
		petCount             int
		serviceCount         int
		confirmedStatusCount int
		dashboardSeedCount   int
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
		SELECT COUNT(*)
		FROM company_services cs
		INNER JOIN companies c ON c.id = cs.company_id
		INNER JOIN services s ON s.id = cs.service_id
		WHERE c.slug = 'petcontrol-dev'
		  AND s.title = 'Banho completo'
		  AND cs.is_active = TRUE
		  AND s.deleted_at IS NULL
	`).Scan(&serviceCount)
	require.NoError(t, err)
	require.Equal(t, 1, serviceCount)

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
