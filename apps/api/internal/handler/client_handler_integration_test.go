package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/handler"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
	"github.com/xdouglas90/petcontrol_monorepo/test/integration"
)

type stubClientUploadResolver struct {
	publicURL string
	err       error
}

func (s *stubClientUploadResolver) ResolveObjectKey(_ context.Context, resource string, field string, objectKey string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.publicURL, nil
}

func TestClientEndpoints_ListIsTenantScoped(t *testing.T) {
	ctx := context.Background()
	pool := setupClientIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "client-list-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "client-list-b")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachClientModule(t, ctx, pool, tenantB.companyID, moduleID)

	clientA := mustCreateClientRecord(t, ctx, pool, queries, tenantA.companyID, "Ana Lima", "12345678902")
	clientB := mustCreateClientRecord(t, ctx, pool, queries, tenantB.companyID, "Bruno Costa", "12345678903")

	router := setupClientRouterForTenant(pool, queries, tenantA)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), clientA.String())
	require.NotContains(t, res.Body.String(), clientB.String())
	require.Contains(t, res.Body.String(), "Ana Lima")
}

func TestClientEndpoints_CreateUsesTenantContextAndAudits(t *testing.T) {
	ctx := context.Background()
	pool := setupClientIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "client-create-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "client-create-b")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachClientModule(t, ctx, pool, tenantB.companyID, moduleID)

	router := setupClientRouterForTenant(pool, queries, tenantA)
	body, err := json.Marshal(map[string]any{
		"full_name":       "Maria Souza",
		"short_name":      "Maria",
		"gender_identity": "woman_cisgender",
		"marital_status":  "single",
		"birth_date":      "1992-06-15",
		"cpf":             "123.456.789-04",
		"email":           "maria.souza@petcontrol.local",
		"phone":           "+551130000000",
		"cellphone":       "+5511999990004",
		"has_whatsapp":    true,
		"client_since":    "2026-04-01",
		"notes":           "Cliente criada via teste",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Contains(t, res.Body.String(), "Maria Souza")
	require.Contains(t, res.Body.String(), "\"company_id\":\""+tenantA.companyID.String()+"\"")

	listA, err := queries.ListClientsByCompanyID(ctx, sqlc.ListClientsByCompanyIDParams{
		CompanyID: tenantA.companyID,
		Limit:     10,
	})
	require.NoError(t, err)
	require.Len(t, listA, 1)
	require.Equal(t, "Maria Souza", listA[0].FullName)
	require.Equal(t, "12345678904", listA[0].Cpf)

	listB, err := queries.ListClientsByCompanyID(ctx, sqlc.ListClientsByCompanyIDParams{
		CompanyID: tenantB.companyID,
		Limit:     10,
	})
	require.NoError(t, err)
	require.Len(t, listB, 0)

	var auditCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE entity_table = 'clients' AND action = 'create' AND company_id = $1`, tenantA.companyID).Scan(&auditCount)
	require.NoError(t, err)
	require.Equal(t, 1, auditCount)
}

func TestClientEndpoints_GetUpdateDeleteRespectTenant(t *testing.T) {
	ctx := context.Background()
	pool := setupClientIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "client-update-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "client-update-b")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachClientModule(t, ctx, pool, tenantB.companyID, moduleID)

	clientA := mustCreateClientRecord(t, ctx, pool, queries, tenantA.companyID, "Ana Lima", "12345678905")
	clientB := mustCreateClientRecord(t, ctx, pool, queries, tenantB.companyID, "Bruno Costa", "12345678906")

	router := setupClientRouterForTenant(pool, queries, tenantA)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientA.String(), nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "Ana Lima")

	req = httptest.NewRequest(http.MethodGet, "/api/v1/clients/"+clientB.String(), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)

	updateBody, err := json.Marshal(map[string]any{
		"short_name":   "Aninha",
		"email":        "ana.lima@petcontrol.local",
		"cellphone":    "+5511999990005",
		"notes":        "Atualizada",
		"has_whatsapp": true,
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientA.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "Aninha")
	require.Contains(t, res.Body.String(), "ana.lima@petcontrol.local")

	req = httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientB.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/clients/"+clientA.String(), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	_, err = queries.GetClientByIDAndCompanyID(ctx, sqlc.GetClientByIDAndCompanyIDParams{
		CompanyID: tenantA.companyID,
		ID:        clientA,
	})
	require.Error(t, err)

	var updateAuditCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE entity_table = 'clients' AND action = 'update' AND company_id = $1`, tenantA.companyID).Scan(&updateAuditCount)
	require.NoError(t, err)
	require.Equal(t, 1, updateAuditCount)

	var deactivateAuditCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE entity_table = 'clients' AND action = 'deactivate' AND company_id = $1`, tenantA.companyID).Scan(&deactivateAuditCount)
	require.NoError(t, err)
	require.Equal(t, 1, deactivateAuditCount)
}

func TestClientEndpoints_RequireModuleForAccess(t *testing.T) {
	ctx := context.Background()
	pool := setupClientIntegrationPool(t)
	queries := sqlc.New(pool)

	tenant := mustCreateTenantFixture(t, ctx, pool, "client-no-module")
	router := setupClientRouterForTenant(pool, queries, tenant)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "module not available for company")
}

func TestClientEndpoints_CreateAndUpdateUsingUploadObjectKey(t *testing.T) {
	ctx := context.Background()
	pool := setupClientIntegrationPool(t)
	queries := sqlc.New(pool)

	tenant := mustCreateTenantFixture(t, ctx, pool, "client-upload-key")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenant.companyID, moduleID)

	uploadURL := "https://cdn.example.com/clients/avatar.png"
	router := setupClientRouterForTenantWithUploadResolver(pool, queries, tenant, &stubClientUploadResolver{publicURL: uploadURL})

	createBody, err := json.Marshal(map[string]any{
		"full_name":         "Maria Souza",
		"short_name":        "Maria",
		"gender_identity":   "woman_cisgender",
		"marital_status":    "single",
		"birth_date":        "1992-06-15",
		"cpf":               "123.456.789-11",
		"email":             "maria.upload@petcontrol.local",
		"cellphone":         "+5511999990099",
		"has_whatsapp":      true,
		"upload_object_key": "uploads/people_identifications/image_url/2026/04/avatar.png",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)

	items, err := queries.ListClientsByCompanyID(ctx, sqlc.ListClientsByCompanyIDParams{
		CompanyID: tenant.companyID,
		Limit:     10,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)

	var persistedImageURL string
	err = pool.QueryRow(ctx, `SELECT image_url FROM people_identifications WHERE person_id = $1`, items[0].PersonID).Scan(&persistedImageURL)
	require.NoError(t, err)
	require.Equal(t, uploadURL, persistedImageURL)

	updatedURL := "https://cdn.example.com/clients/avatar-updated.png"
	router = setupClientRouterForTenantWithUploadResolver(pool, queries, tenant, &stubClientUploadResolver{publicURL: updatedURL})

	updateBody, err := json.Marshal(map[string]any{
		"upload_object_key": "uploads/people_identifications/image_url/2026/04/avatar-updated.png",
		"has_whatsapp":      true,
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+items[0].ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	err = pool.QueryRow(ctx, `SELECT image_url FROM people_identifications WHERE person_id = $1`, items[0].PersonID).Scan(&persistedImageURL)
	require.NoError(t, err)
	require.Equal(t, updatedURL, persistedImageURL)
}

func TestClientEndpoints_UpdateRejectsInvalidUploadObjectKey(t *testing.T) {
	ctx := context.Background()
	pool := setupClientIntegrationPool(t)
	queries := sqlc.New(pool)

	tenant := mustCreateTenantFixture(t, ctx, pool, "client-upload-invalid")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenant.companyID, moduleID)
	clientID := mustCreateClientRecord(t, ctx, pool, queries, tenant.companyID, "Ana Lima", "12345678912")

	router := setupClientRouterForTenantWithUploadResolver(pool, queries, tenant, &stubClientUploadResolver{err: apperror.ErrUnprocessableEntity})
	body, err := json.Marshal(map[string]any{
		"upload_object_key": "uploads/people_identifications/image_url/2026/04/invalid.png",
		"has_whatsapp":      true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/clients/"+clientID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnprocessableEntity, res.Code)
	require.Contains(t, res.Body.String(), "invalid_upload_object_key")
}

func setupClientIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	setup := integration.SetupPostgresWithMigrations(t)
	return setup.Pool
}

func setupClientRouterForTenant(pool *pgxpool.Pool, queries *sqlc.Queries, tenant integrationTenantFixture) *gin.Engine {
	return setupClientRouterForTenantWithUploadResolver(pool, queries, tenant)
}

func setupClientRouterForTenantWithUploadResolver(pool *pgxpool.Pool, queries *sqlc.Queries, tenant integrationTenantFixture, uploadResolver ...interface {
	ResolveObjectKey(ctx context.Context, resource string, field string, objectKey string) (string, error)
}) *gin.Engine {
	gin.SetMode(gin.TestMode)

	clientService := service.NewClientService(pool, queries)
	clientHandler := handler.NewClientHandler(clientService)
	if len(uploadResolver) > 0 {
		clientHandler = handler.NewClientHandler(clientService, uploadResolver[0])
	}

	router := gin.New()
	router.Use(middleware.RequestContext())
	router.Use(func(c *gin.Context) {
		c.Set("company_id", tenant.companyID)
		c.Next()
	})
	router.Use(middleware.Audit(queries, nil))

	clients := router.Group("/api/v1/clients")
	clients.Use(middleware.RequireModule(queries, "CLI"))
	clients.GET("", clientHandler.List)
	clients.POST("", clientHandler.Create)
	clients.GET("/:id", clientHandler.GetByID)
	clients.PUT("/:id", clientHandler.Update)
	clients.DELETE("/:id", clientHandler.Delete)

	return router
}

func mustCreateClientModule(t *testing.T, ctx context.Context, pool *pgxpool.Pool) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO modules (code, name, description, min_package, is_active)
		VALUES ('CLI', 'Clients', 'Clients module', 'starter', TRUE)
		RETURNING id
	`).Scan(&id)
	require.NoError(t, err)
	return id
}

func mustCreatePetModule(t *testing.T, ctx context.Context, pool *pgxpool.Pool) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO modules (code, name, description, min_package, is_active)
		VALUES ('PET', 'Pets', 'Pets module', 'starter', TRUE)
		RETURNING id
	`).Scan(&id)
	require.NoError(t, err)
	return id
}

func mustAttachClientModule(t *testing.T, ctx context.Context, pool *pgxpool.Pool, companyID pgtype.UUID, moduleID pgtype.UUID) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO company_modules (company_id, module_id, is_active)
		VALUES ($1, $2, TRUE)
	`, companyID, moduleID)
	require.NoError(t, err)
}

func mustCreateClientRecord(t *testing.T, ctx context.Context, pool *pgxpool.Pool, queries *sqlc.Queries, companyID pgtype.UUID, fullName string, cpf string) pgtype.UUID {
	t.Helper()

	birthDate := pgtype.Date{Time: time.Date(1992, 6, 15, 0, 0, 0, 0, time.UTC), Valid: true}
	clientSince := pgtype.Date{Time: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), Valid: true}

	clientService := service.NewClientService(pool, queries)
	created, err := clientService.CreateClient(ctx, service.CreateClientInput{
		CompanyID:      companyID,
		FullName:       fullName,
		ShortName:      fullName,
		GenderIdentity: sqlc.GenderIdentityWomanCisgender,
		MaritalStatus:  sqlc.MaritalStatusSingle,
		BirthDate:      birthDate,
		CPF:            cpf,
		Email:          uuid.NewString() + "@petcontrol.local",
		Phone:          pgtype.Text{String: "+551130000000", Valid: true},
		Cellphone:      "+5511999991111",
		HasWhatsapp:    true,
		ClientSince:    clientSince,
		Notes:          pgtype.Text{String: "Criado no teste", Valid: true},
	})
	require.NoError(t, err)
	return created.ID
}
