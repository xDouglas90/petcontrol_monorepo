package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/handler"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
	"github.com/xdouglas90/petcontrol_monorepo/test/integration"
)

type integrationTenantFixture struct {
	companyID pgtype.UUID
	userID    pgtype.UUID
	clientID  pgtype.UUID
	petID     pgtype.UUID
}

func TestScheduleEndpoints_ListIsTenantScoped(t *testing.T) {
	ctx := context.Background()
	pool := setupScheduleIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "tenant-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "tenant-b")
	moduleID := mustCreateScheduleModule(t, ctx, pool)
	mustAttachScheduleModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachScheduleModule(t, ctx, pool, tenantB.companyID, moduleID)

	scheduledAt := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Second)
	scheduleA := mustCreateScheduleRecord(t, ctx, queries, tenantA, scheduledAt, "tenant A schedule")
	scheduleB := mustCreateScheduleRecord(t, ctx, queries, tenantB, scheduledAt.Add(1*time.Hour), "tenant B schedule")

	router := setupScheduleRouterForTenant(queries, tenantA)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), scheduleA.String())
	require.NotContains(t, res.Body.String(), scheduleB.String())
}

func TestScheduleEndpoints_CreateUsesTenantFromContext(t *testing.T) {
	ctx := context.Background()
	pool := setupScheduleIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "tenant-create-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "tenant-create-b")
	moduleID := mustCreateScheduleModule(t, ctx, pool)
	mustAttachScheduleModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachScheduleModule(t, ctx, pool, tenantB.companyID, moduleID)

	router := setupScheduleRouterForTenant(queries, tenantA)
	scheduledAt := time.Now().UTC().Add(4 * time.Hour).Truncate(time.Second)
	estimatedEnd := scheduledAt.Add(75 * time.Minute)

	body, err := json.Marshal(map[string]any{
		"client_id":     tenantA.clientID.String(),
		"pet_id":        tenantA.petID.String(),
		"scheduled_at":  scheduledAt.Format(time.RFC3339),
		"estimated_end": estimatedEnd.Format(time.RFC3339),
		"notes":         "consulta",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/schedules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)

	listA, err := queries.ListSchedulesByCompanyID(ctx, tenantA.companyID)
	require.NoError(t, err)
	require.Len(t, listA, 1)
	require.Equal(t, tenantA.companyID, listA[0].CompanyID)
	require.Equal(t, tenantA.clientID, listA[0].ClientID)

	listB, err := queries.ListSchedulesByCompanyID(ctx, tenantB.companyID)
	require.NoError(t, err)
	require.Len(t, listB, 0)
}

func TestScheduleEndpoints_UpdateDeleteRespectTenantAndSoftDelete(t *testing.T) {
	ctx := context.Background()
	pool := setupScheduleIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "tenant-update-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "tenant-update-b")
	moduleID := mustCreateScheduleModule(t, ctx, pool)
	mustAttachScheduleModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachScheduleModule(t, ctx, pool, tenantB.companyID, moduleID)

	scheduledAt := time.Now().UTC().Add(6 * time.Hour).Truncate(time.Second)
	scheduleA := mustCreateScheduleRecord(t, ctx, queries, tenantA, scheduledAt, "original")
	scheduleB := mustCreateScheduleRecord(t, ctx, queries, tenantB, scheduledAt.Add(1*time.Hour), "other tenant")

	router := setupScheduleRouterForTenant(queries, tenantA)
	updateBody, err := json.Marshal(map[string]any{
		"notes":  "updated-note",
		"status": "confirmed",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/schedules/"+scheduleA.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "updated-note")
	require.Contains(t, res.Body.String(), "confirmed")

	req = httptest.NewRequest(http.MethodPut, "/api/v1/schedules/"+scheduleB.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/schedules/"+scheduleA.String(), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	var deletedAt pgtype.Timestamptz
	err = pool.QueryRow(ctx, "SELECT deleted_at FROM schedules WHERE id = $1", scheduleA).Scan(&deletedAt)
	require.NoError(t, err)
	require.True(t, deletedAt.Valid)
}

func setupScheduleIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	setup := integration.SetupPostgresWithMigrations(t)
	return setup.Pool
}

func setupScheduleRouterForTenant(queries sqlc.Querier, tenant integrationTenantFixture) *gin.Engine {
	gin.SetMode(gin.TestMode)

	scheduleService := service.NewScheduleService(queries)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", tenant.companyID)
		c.Set("auth_claims", appjwt.Claims{
			UserID:    tenant.userID.String(),
			CompanyID: tenant.companyID.String(),
			Role:      "admin",
			Kind:      "owner",
		})
		c.Next()
	})

	schedules := router.Group("/api/v1/schedules")
	schedules.Use(middleware.RequireModule(queries, "SCH"))
	schedules.GET("", scheduleHandler.List)
	schedules.POST("", scheduleHandler.Create)
	schedules.GET("/:id", scheduleHandler.GetByID)
	schedules.PUT("/:id", scheduleHandler.Update)
	schedules.DELETE("/:id", scheduleHandler.Delete)

	return router
}

func mustCreateTenantFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool, slug string) integrationTenantFixture {
	t.Helper()

	responsible := mustInsertPerson(t, ctx, pool, "responsible")
	companyID := mustInsertCompany(t, ctx, pool, slug, responsible)
	userID := mustInsertUser(t, ctx, pool, slug)
	clientPerson := mustInsertPerson(t, ctx, pool, "client")
	clientID := mustInsertClient(t, ctx, pool, clientPerson)
	mustLinkCompanyClient(t, ctx, pool, companyID, clientID)
	petID := mustInsertPet(t, ctx, pool, clientID)

	return integrationTenantFixture{
		companyID: companyID,
		userID:    userID,
		clientID:  clientID,
		petID:     petID,
	}
}

func mustInsertPerson(t *testing.T, ctx context.Context, pool *pgxpool.Pool, kind string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	err := pool.QueryRow(ctx, "INSERT INTO people(kind, is_active, has_system_user) VALUES ($1, TRUE, FALSE) RETURNING id", kind).Scan(&id)
	require.NoError(t, err)
	return id
}

func mustInsertCompany(t *testing.T, ctx context.Context, pool *pgxpool.Pool, slug string, responsibleID pgtype.UUID) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	cnpj := fmt.Sprintf("%014d", (time.Now().UnixNano()+int64(len(slug))*97)%100000000000000)
	err := pool.QueryRow(ctx, `
		INSERT INTO companies (slug, name, fantasy_name, cnpj, responsible_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, slug, "Company "+slug, "Fantasia "+slug, cnpj, responsibleID).Scan(&id)
	require.NoError(t, err)
	return id
}

func mustInsertUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, slug string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	email := slug + "@example.com"
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, email_verified, role, kind, is_active)
		VALUES ($1, TRUE, 'admin', 'owner', TRUE)
		RETURNING id
	`, email).Scan(&id)
	require.NoError(t, err)
	return id
}

func mustInsertClient(t *testing.T, ctx context.Context, pool *pgxpool.Pool, personID pgtype.UUID) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	err := pool.QueryRow(ctx, "INSERT INTO clients (person_id) VALUES ($1) RETURNING id", personID).Scan(&id)
	require.NoError(t, err)
	return id
}

func mustLinkCompanyClient(t *testing.T, ctx context.Context, pool *pgxpool.Pool, companyID pgtype.UUID, clientID pgtype.UUID) {
	t.Helper()
	_, err := pool.Exec(ctx, "INSERT INTO company_clients (company_id, client_id, is_active) VALUES ($1, $2, TRUE)", companyID, clientID)
	require.NoError(t, err)
}

func mustInsertPet(t *testing.T, ctx context.Context, pool *pgxpool.Pool, clientID pgtype.UUID) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO pets (name, size, kind, temperament, owner_id, is_active)
		VALUES ('Rex', 'small', 'dog', 'calm', $1, TRUE)
		RETURNING id
	`, clientID).Scan(&id)
	require.NoError(t, err)
	return id
}

func mustCreateScheduleModule(t *testing.T, ctx context.Context, pool *pgxpool.Pool) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO modules (code, name, description, min_package, is_active)
		VALUES ('SCH', 'Scheduling', 'Scheduling module', 'starter', TRUE)
		RETURNING id
	`).Scan(&id)
	require.NoError(t, err)
	return id
}

func mustAttachScheduleModule(t *testing.T, ctx context.Context, pool *pgxpool.Pool, companyID pgtype.UUID, moduleID pgtype.UUID) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO company_modules (company_id, module_id, is_active)
		VALUES ($1, $2, TRUE)
	`, companyID, moduleID)
	require.NoError(t, err)
}

func mustCreateScheduleRecord(t *testing.T, ctx context.Context, queries *sqlc.Queries, tenant integrationTenantFixture, at time.Time, notes string) uuid.UUID {
	t.Helper()

	created, err := queries.CreateSchedule(ctx, sqlc.CreateScheduleParams{
		CompanyID:   tenant.companyID,
		ClientID:    tenant.clientID,
		PetID:       tenant.petID,
		ScheduledAt: pgtype.Timestamptz{Time: at, Valid: true},
		Notes:       pgtype.Text{String: notes, Valid: true},
		CreatedBy:   tenant.userID,
	})
	require.NoError(t, err)

	_, err = queries.InsertScheduleStatusHistory(ctx, sqlc.InsertScheduleStatusHistoryParams{
		ScheduleID: created.ID,
		Status:     sqlc.ScheduleStatusWaiting,
		ChangedBy:  tenant.userID,
	})
	require.NoError(t, err)

	id, err := uuid.FromBytes(created.ID.Bytes[:])
	require.NoError(t, err)
	return id
}
