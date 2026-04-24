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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/handler"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
	"github.com/xdouglas90/petcontrol_monorepo/test/integration"
)

type stubPetUploadResolver struct {
	publicURL string
	err       error
}

func (s *stubPetUploadResolver) ResolveObjectKey(_ context.Context, resource string, field string, objectKey string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.publicURL, nil
}

func TestPetEndpoints_ListIsTenantScoped(t *testing.T) {
	ctx := context.Background()
	pool := setupPetIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "pet-list-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "pet-list-b")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachClientModule(t, ctx, pool, tenantB.companyID, moduleID)
	clientA := mustCreateClientRecord(t, ctx, pool, queries, tenantA.companyID, "Ana Lima", "12345678922")
	clientB := mustCreateClientRecord(t, ctx, pool, queries, tenantB.companyID, "Bruno Costa", "12345678923")

	petA := mustCreatePetRecord(t, ctx, queries, clientA, "Thor")
	petB := mustCreatePetRecord(t, ctx, queries, clientB, "Mingau")

	router := setupPetRouterForTenant(queries, tenantA)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pets", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), petA.String())
	require.Contains(t, res.Body.String(), "Thor")
	require.NotContains(t, res.Body.String(), petB.String())
	require.NotContains(t, res.Body.String(), "Mingau")
}

func TestPetEndpoints_CreateRejectsOwnerFromAnotherTenant(t *testing.T) {
	ctx := context.Background()
	pool := setupPetIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "pet-create-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "pet-create-b")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachClientModule(t, ctx, pool, tenantB.companyID, moduleID)
	clientB := mustCreateClientRecord(t, ctx, pool, queries, tenantB.companyID, "Bruno Costa", "12345678925")

	router := setupPetRouterForTenant(queries, tenantA)
	body, err := json.Marshal(map[string]any{
		"owner_id":    clientB.String(),
		"name":        "Thor",
		"race":        "SRD",
		"color":       "Preto",
		"sex":         "M",
		"size":        "medium",
		"kind":        "dog",
		"temperament": "playful",
		"birth_date":  "2021-08-20",
		"notes":       "Cross tenant should fail",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnprocessableEntity, res.Code)

	items, err := queries.ListPetsByCompanyID(ctx, sqlc.ListPetsByCompanyIDParams{
		CompanyID: tenantA.companyID,
		Limit:     10,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Rex", items[0].Name)
}

func TestPetEndpoints_CreateUpdateDeleteRespectTenant(t *testing.T) {
	ctx := context.Background()
	pool := setupPetIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "pet-crud-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "pet-crud-b")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachClientModule(t, ctx, pool, tenantB.companyID, moduleID)
	clientA := mustCreateClientRecord(t, ctx, pool, queries, tenantA.companyID, "Ana Lima", "12345678926")
	clientB := mustCreateClientRecord(t, ctx, pool, queries, tenantB.companyID, "Bruno Costa", "12345678927")

	petB := mustCreatePetRecord(t, ctx, queries, clientB, "Mingau")
	router := setupPetRouterForTenant(queries, tenantA)

	createBody, err := json.Marshal(map[string]any{
		"owner_id":    clientA.String(),
		"name":        "Thor",
		"race":        "SRD",
		"color":       "Preto",
		"sex":         "M",
		"size":        "medium",
		"kind":        "dog",
		"temperament": "playful",
		"birth_date":  "2021-08-20",
		"notes":       "Criado no teste",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pets", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusCreated, res.Code)
	require.Contains(t, res.Body.String(), "Thor")
	require.Contains(t, res.Body.String(), "\"owner_name\":")

	items, err := queries.ListPetsByCompanyID(ctx, sqlc.ListPetsByCompanyIDParams{
		CompanyID: tenantA.companyID,
		Limit:     10,
	})
	require.NoError(t, err)
	require.Len(t, items, 2)

	var petAID pgtype.UUID
	for _, item := range items {
		if item.Name == "Thor" {
			petAID = item.ID
		}
	}
	require.True(t, petAID.Valid)

	updateBody, err := json.Marshal(map[string]any{
		"name":        "Thor Atualizado",
		"temperament": "loving",
		"notes":       "Atualizado no teste",
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/pets/"+petAID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "Thor Atualizado")
	require.Contains(t, res.Body.String(), "loving")

	req = httptest.NewRequest(http.MethodPut, "/api/v1/pets/"+petB.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/pets/"+petAID.String(), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	_, err = queries.GetPetByIDAndCompanyID(ctx, sqlc.GetPetByIDAndCompanyIDParams{
		CompanyID: tenantA.companyID,
		ID:        petAID,
	})
	require.Error(t, err)

	var updateAuditCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE entity_table = 'pets' AND action = 'update' AND company_id = $1`, tenantA.companyID).Scan(&updateAuditCount)
	require.NoError(t, err)
	require.Equal(t, 1, updateAuditCount)

	var deleteAuditCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE entity_table = 'pets' AND action = 'delete' AND company_id = $1`, tenantA.companyID).Scan(&deleteAuditCount)
	require.NoError(t, err)
	require.Equal(t, 1, deleteAuditCount)
}

func TestPetEndpoints_RequireModuleForAccess(t *testing.T) {
	ctx := context.Background()
	pool := setupPetIntegrationPool(t)
	queries := sqlc.New(pool)

	tenant := mustCreateTenantFixture(t, ctx, pool, "pet-no-module")
	router := setupPetRouterForTenant(queries, tenant)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pets", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "module not available for company")
}

func TestPetEndpoints_CreateAndUpdateUsingUploadObjectKey(t *testing.T) {
	ctx := context.Background()
	pool := setupPetIntegrationPool(t)
	queries := sqlc.New(pool)

	tenant := mustCreateTenantFixture(t, ctx, pool, "pet-upload-key")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenant.companyID, moduleID)
	clientID := mustCreateClientRecord(t, ctx, pool, queries, tenant.companyID, "Ana Lima", "12345678928")

	uploadURL := "https://cdn.example.com/pets/thor.png"
	router := setupPetRouterForTenantWithUploadResolver(queries, tenant, &stubPetUploadResolver{publicURL: uploadURL})

	createBody, err := json.Marshal(map[string]any{
		"owner_id":          clientID.String(),
		"name":              "Thor",
		"race":              "SRD",
		"color":             "Preto",
		"sex":               "M",
		"size":              "medium",
		"kind":              "dog",
		"temperament":       "playful",
		"upload_object_key": "uploads/pets/image_url/2026/04/thor.png",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pets", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Contains(t, res.Body.String(), uploadURL)

	items, err := queries.ListPetsByCompanyID(ctx, sqlc.ListPetsByCompanyIDParams{
		CompanyID: tenant.companyID,
		Limit:     10,
	})
	require.NoError(t, err)
	require.Len(t, items, 2)

	var petID pgtype.UUID
	for _, item := range items {
		if item.Name == "Thor" {
			petID = item.ID
			require.Equal(t, uploadURL, item.ImageUrl.String)
		}
	}
	require.True(t, petID.Valid)

	updatedURL := "https://cdn.example.com/pets/thor-updated.png"
	router = setupPetRouterForTenantWithUploadResolver(queries, tenant, &stubPetUploadResolver{publicURL: updatedURL})
	updateBody, err := json.Marshal(map[string]any{
		"upload_object_key": "uploads/pets/image_url/2026/04/thor-updated.png",
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/pets/"+petID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), updatedURL)

	item, err := queries.GetPetByIDAndCompanyID(ctx, sqlc.GetPetByIDAndCompanyIDParams{
		CompanyID: tenant.companyID,
		ID:        petID,
	})
	require.NoError(t, err)
	require.Equal(t, updatedURL, item.ImageUrl.String)
}

func TestPetEndpoints_CreateAndUpdateWithGuardianIDs(t *testing.T) {
	ctx := context.Background()
	pool := setupPetIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "pet-guardian-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "pet-guardian-b")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachClientModule(t, ctx, pool, tenantB.companyID, moduleID)

	clientID := mustCreateClientRecord(t, ctx, pool, queries, tenantA.companyID, "Ana Lima", "12345678931")
	guardianA := mustCreateGuardianPersonRecord(t, ctx, pool, tenantA.companyID, "Guardiao A", "12345678932")
	guardianB := mustCreateGuardianPersonRecord(t, ctx, pool, tenantA.companyID, "Guardiao B", "12345678933")
	guardianOtherTenant := mustCreateGuardianPersonRecord(t, ctx, pool, tenantB.companyID, "Guardiao Externo", "12345678934")

	router := setupPetRouterForTenant(queries, tenantA)

	rejectBody, err := json.Marshal(map[string]any{
		"owner_id":     clientID.String(),
		"guardian_ids": []string{guardianOtherTenant.String()},
		"name":         "Thor",
		"race":         "SRD",
		"color":        "Preto",
		"sex":          "M",
		"size":         "medium",
		"kind":         "dog",
		"temperament":  "playful",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pets", bytes.NewReader(rejectBody))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusUnprocessableEntity, res.Code)

	createBody, err := json.Marshal(map[string]any{
		"owner_id":     clientID.String(),
		"guardian_ids": []string{guardianA.String()},
		"name":         "Thor",
		"race":         "SRD",
		"color":        "Preto",
		"sex":          "M",
		"size":         "medium",
		"kind":         "dog",
		"temperament":  "playful",
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/pets", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusCreated, res.Code)
	require.Contains(t, res.Body.String(), guardianA.String())

	items, err := queries.ListPetsByCompanyID(ctx, sqlc.ListPetsByCompanyIDParams{
		CompanyID: tenantA.companyID,
		Limit:     10,
	})
	require.NoError(t, err)

	var petID pgtype.UUID
	for _, item := range items {
		if item.Name == "Thor" {
			petID = item.ID
			break
		}
	}
	require.True(t, petID.Valid)

	updateBody, err := json.Marshal(map[string]any{
		"guardian_ids": []string{guardianB.String()},
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPut, "/api/v1/pets/"+petID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), guardianB.String())
	require.NotContains(t, res.Body.String(), guardianA.String())

	guardians, err := queries.ListPetGuardiansByPetID(ctx, sqlc.ListPetGuardiansByPetIDParams{
		PetID:     petID,
		CompanyID: tenantA.companyID,
	})
	require.NoError(t, err)
	require.Len(t, guardians, 1)
	require.Equal(t, guardianB.String(), guardians[0].GuardianID.String())
}

func TestPetEndpoints_ListSupportsStructuredFilters(t *testing.T) {
	ctx := context.Background()
	pool := setupPetIntegrationPool(t)
	queries := sqlc.New(pool)

	tenant := mustCreateTenantFixture(t, ctx, pool, "pet-filters")
	moduleID := mustCreateClientModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenant.companyID, moduleID)

	clientID := mustCreateClientRecord(t, ctx, pool, queries, tenant.companyID, "Ana Lima", "12345678935")
	dogID := mustCreatePetRecordWithAttrs(t, ctx, queries, clientID, "Thor", sqlc.PetKindDog, sqlc.PetSizeMedium, sqlc.PetTemperamentPlayful, true)
	_ = mustCreatePetRecordWithAttrs(t, ctx, queries, clientID, "Mingau", sqlc.PetKindCat, sqlc.PetSizeSmall, sqlc.PetTemperamentCalm, false)

	router := setupPetRouterForTenant(queries, tenant)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pets?kind=dog&size=medium&temperament=playful&is_active=true", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), dogID.String())
	require.Contains(t, res.Body.String(), "Thor")
	require.NotContains(t, res.Body.String(), "Mingau")
}

func setupPetIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	setup := integration.SetupPostgresWithMigrations(t)
	return setup.Pool
}

func setupPetRouterForTenant(queries sqlc.Querier, tenant integrationTenantFixture) *gin.Engine {
	return setupPetRouterForTenantWithUploadResolver(queries, tenant, nil)
}

func setupPetRouterForTenantWithUploadResolver(queries sqlc.Querier, tenant integrationTenantFixture, uploadResolver handlerUploadResolver) *gin.Engine {
	gin.SetMode(gin.TestMode)

	petService := service.NewPetService(queries)
	petHandler := handler.NewPetHandler(petService, uploadResolver)

	router := gin.New()
	router.Use(middleware.RequestContext())
	router.Use(func(c *gin.Context) {
		c.Set("company_id", tenant.companyID)
		c.Next()
	})
	router.Use(middleware.Audit(queries, nil))

	pets := router.Group("/api/v1/pets")
	pets.Use(middleware.RequireModule(queries, "CRM"))
	pets.GET("", petHandler.List)
	pets.POST("", petHandler.Create)
	pets.GET("/:id", petHandler.GetByID)
	pets.PUT("/:id", petHandler.Update)
	pets.DELETE("/:id", petHandler.Delete)

	return router
}

type handlerUploadResolver interface {
	ResolveObjectKey(ctx context.Context, resource string, field string, objectKey string) (string, error)
}

func mustCreatePetRecord(t *testing.T, ctx context.Context, queries *sqlc.Queries, ownerID pgtype.UUID, name string) pgtype.UUID {
	t.Helper()
	return mustCreatePetRecordWithAttrs(t, ctx, queries, ownerID, name, sqlc.PetKindDog, sqlc.PetSizeMedium, sqlc.PetTemperamentPlayful, true)
}

func mustCreatePetRecordWithAttrs(t *testing.T, ctx context.Context, queries *sqlc.Queries, ownerID pgtype.UUID, name string, kind sqlc.PetKind, size sqlc.PetSize, temperament sqlc.PetTemperament, isActive bool) pgtype.UUID {
	t.Helper()
	pet, err := queries.CreatePet(ctx, sqlc.CreatePetParams{
		Name:           name,
		Race:           "SRD",
		Color:          "Preto",
		Sex:            "M",
		Size:           size,
		Kind:           kind,
		Temperament:    temperament,
		BirthDate:      pgtype.Date{Time: time.Date(2021, 8, 20, 0, 0, 0, 0, time.UTC), Valid: true},
		OwnerID:        ownerID,
		IsActive:       pgtype.Bool{Bool: isActive, Valid: true},
		IsDeceased:     pgtype.Bool{Bool: false, Valid: true},
		IsVaccinated:   pgtype.Bool{Bool: false, Valid: true},
		IsNeutered:     pgtype.Bool{Bool: false, Valid: true},
		IsMicrochipped: pgtype.Bool{Bool: false, Valid: true},
		Notes:          pgtype.Text{String: "Criado no teste", Valid: true},
	})
	require.NoError(t, err)
	return pet.ID
}

func mustCreateGuardianPersonRecord(t *testing.T, ctx context.Context, pool *pgxpool.Pool, companyID pgtype.UUID, fullName string, cpf string) pgtype.UUID {
	t.Helper()

	var personID pgtype.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO people (kind, is_active, has_system_user)
		VALUES ('guardian', TRUE, FALSE)
		RETURNING id
	`).Scan(&personID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO people_identifications (person_id, full_name, short_name, gender_identity, marital_status, birth_date, cpf)
		VALUES ($1, $2, $3, 'not_to_expose', 'single', DATE '1990-01-01', $4)
	`, personID, fullName, fullName, cpf)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO people_contacts (person_id, email, cellphone, has_whatsapp, is_primary)
		VALUES ($1, $2, '+5511999990000', TRUE, TRUE)
	`, personID, "guardian-"+cpf+"@petcontrol.local")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO company_people (company_id, person_id)
		VALUES ($1, $2)
	`, companyID, personID)
	require.NoError(t, err)

	return personID
}
