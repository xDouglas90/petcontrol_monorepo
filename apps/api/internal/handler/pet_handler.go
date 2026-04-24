package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/pagination"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type PetHandler struct {
	service        *service.PetService
	uploadResolver petUploadResolver
}

type createPetRequest struct {
	OwnerID                 string   `json:"owner_id"`
	GuardianIDs             []string `json:"guardian_ids"`
	Name                    string   `json:"name"`
	Race                    string   `json:"race"`
	Color                   string   `json:"color"`
	Sex                     string   `json:"sex"`
	Size                    string   `json:"size"`
	Kind                    string   `json:"kind"`
	Temperament             string   `json:"temperament"`
	ImageURL                string   `json:"image_url"`
	UploadKey               string   `json:"upload_object_key"`
	BirthDate               string   `json:"birth_date"`
	IsActive                *bool    `json:"is_active"`
	IsDeceased              *bool    `json:"is_deceased"`
	IsVaccinated            *bool    `json:"is_vaccinated"`
	IsNeutered              *bool    `json:"is_neutered"`
	IsMicrochipped          *bool    `json:"is_microchipped"`
	MicrochipNumber         string   `json:"microchip_number"`
	MicrochipExpirationDate string   `json:"microchip_expiration_date"`
	Notes                   string   `json:"notes"`
}

type updatePetRequest struct {
	OwnerID                 *string   `json:"owner_id"`
	GuardianIDs             *[]string `json:"guardian_ids"`
	Name                    *string   `json:"name"`
	Race                    *string   `json:"race"`
	Color                   *string   `json:"color"`
	Sex                     *string   `json:"sex"`
	Size                    *string   `json:"size"`
	Kind                    *string   `json:"kind"`
	Temperament             *string   `json:"temperament"`
	ImageURL                *string   `json:"image_url"`
	UploadKey               *string   `json:"upload_object_key"`
	BirthDate               *string   `json:"birth_date"`
	IsActive                *bool     `json:"is_active"`
	IsDeceased              *bool     `json:"is_deceased"`
	IsVaccinated            *bool     `json:"is_vaccinated"`
	IsNeutered              *bool     `json:"is_neutered"`
	IsMicrochipped          *bool     `json:"is_microchipped"`
	MicrochipNumber         *string   `json:"microchip_number"`
	MicrochipExpirationDate *string   `json:"microchip_expiration_date"`
	Notes                   *string   `json:"notes"`
}

type petUploadResolver interface {
	ResolveObjectKey(ctx context.Context, resource string, field string, objectKey string) (string, error)
}

func NewPetHandler(service *service.PetService, uploadResolver ...petUploadResolver) *PetHandler {
	handler := &PetHandler{service: service}
	if len(uploadResolver) > 0 {
		handler.uploadResolver = uploadResolver[0]
	}
	return handler
}

// List godoc
// @Summary List pets
// @Description Returns pets from the authenticated tenant.
// @Tags pets
// @Security BearerAuth
// @Produce json
// @Success 200 {object} PetListResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /pets [get]
func (h *PetHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	params := pagination.ParseParams(c)
	sizeRaw := c.Query("size")
	kindRaw := c.Query("kind")
	temperamentRaw := c.Query("temperament")
	raceRaw := c.Query("race")
	size, err := parseOptionalPetSize(&sizeRaw)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_size", "invalid pet size")
		return
	}
	kind, err := parseOptionalPetKind(&kindRaw)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_kind", "invalid pet kind")
		return
	}
	temperament, err := parseOptionalPetTemperament(&temperamentRaw)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_temperament", "invalid pet temperament")
		return
	}
	filters := service.PetFilters{
		Size:        size,
		Kind:        kind,
		Temperament: temperament,
		Race:        parseOptionalTrimmed(&raceRaw),
		IsActive:    parseOptionalBool(c.Query("is_active")),
	}
	items, err := h.service.ListPetsByCompanyID(c.Request.Context(), companyID, params, filters)
	if err != nil {
		middleware.JSONError(c, http.StatusInternalServerError, "list_pets_failed", "failed to list pets")
		return
	}

	total := 0
	if len(items) > 0 {
		total = int(items[0].TotalCount)
	}

	middleware.JSONPaginated(c, http.StatusOK, items, pagination.NewMeta(total, params.Page, params.Limit))
}

// GetByID godoc
// @Summary Get pet by ID
// @Description Returns a single pet from the authenticated tenant.
// @Tags pets
// @Security BearerAuth
// @Produce json
// @Param id path string true "Pet ID"
// @Success 200 {object} PetItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /pets/{id} [get]
func (h *PetHandler) GetByID(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	petID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_id", "invalid pet id")
		return
	}

	item, err := h.service.GetPetDetailByID(c.Request.Context(), companyID, petID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_pet_failed", "failed to get pet")
		return
	}

	guardians, err := h.service.GetPetGuardians(c.Request.Context(), companyID, petID)
	if err != nil {
		middleware.JSONError(c, http.StatusInternalServerError, "get_guardians_failed", "failed to get pet guardians")
		return
	}

	middleware.JSONData(c, http.StatusOK, gin.H{
		"data": buildPetResponseData(item, guardians),
	})
}

// Create godoc
// @Summary Create pet
// @Description Creates a pet for the authenticated tenant.
// @Tags pets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body PetCreateRequestDoc true "Pet payload"
// @Success 201 {object} PetItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 409 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /pets [post]
func (h *PetHandler) Create(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	var req createPetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	ownerID, err := parseUUID(req.OwnerID)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_owner_id", "invalid owner_id")
		return
	}
	race, err := parseRequiredTrimmed(req.Race)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_race", "invalid pet race")
		return
	}
	color, err := parseRequiredTrimmed(req.Color)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_color", "invalid pet color")
		return
	}
	sex, err := parsePetSex(req.Sex)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_sex", "invalid pet sex")
		return
	}
	guardianIDs, err := parseUUIDSlice(req.GuardianIDs)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_guardian_ids", "invalid guardian_ids")
		return
	}

	size, err := parsePetSize(req.Size)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_size", "invalid pet size")
		return
	}

	kind, err := parsePetKind(req.Kind)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_kind", "invalid pet kind")
		return
	}

	temperament, err := parsePetTemperament(req.Temperament)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_temperament", "invalid pet temperament")
		return
	}

	birthDate, err := parseOptionalDate(req.BirthDate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_birth_date", "invalid birth_date")
		return
	}

	imageURL, err := h.resolveUploadObjectKey(c.Request.Context(), req.UploadKey, strings.TrimSpace(req.ImageURL))
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "invalid_upload_object_key", "invalid upload_object_key")
		return
	}

	microchipDate, err := parseOptionalDate(req.MicrochipExpirationDate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_microchip_expiration_date", "invalid microchip expiration date")
		return
	}

	item, err := h.service.CreatePet(c.Request.Context(), service.CreatePetInput{
		CompanyID:               companyID,
		OwnerID:                 ownerID,
		GuardianIDs:             guardianIDs,
		Name:                    strings.TrimSpace(req.Name),
		Race:                    race,
		Color:                   color,
		Sex:                     sex,
		Size:                    size,
		Kind:                    kind,
		Temperament:             temperament,
		ImageURL:                textValue(imageURL),
		BirthDate:               birthDate,
		IsActive:                boolValueOrDefault(req.IsActive, true),
		IsDeceased:              boolValueOrDefault(req.IsDeceased, false),
		IsVaccinated:            boolValueOrDefault(req.IsVaccinated, false),
		IsNeutered:              boolValueOrDefault(req.IsNeutered, false),
		IsMicrochipped:          boolValueOrDefault(req.IsMicrochipped, false),
		MicrochipNumber:         textValue(strings.TrimSpace(req.MicrochipNumber)),
		MicrochipExpirationDate: microchipDate,
		Notes:                   textValue(strings.TrimSpace(req.Notes)),
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_pet_failed", "failed to create pet")
		return
	}

	guardians, err := h.service.GetPetGuardians(c.Request.Context(), companyID, item.ID)
	if err != nil {
		middleware.JSONError(c, http.StatusInternalServerError, "get_guardians_failed", "failed to get pet guardians")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionCreate,
		EntityTable: "pets",
		EntityID:    item.ID,
		CompanyID:   companyID,
		OldData:     nil,
		NewData:     item,
	})

	middleware.JSONData(c, http.StatusCreated, gin.H{
		"data": buildPetResponseData(item, guardians),
	})
}

// Update godoc
// @Summary Update pet
// @Description Updates a pet from the authenticated tenant.
// @Tags pets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Pet ID"
// @Param request body PetUpdateRequestDoc true "Pet payload"
// @Success 200 {object} PetItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 409 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /pets/{id} [put]
func (h *PetHandler) Update(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	petID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_id", "invalid pet id")
		return
	}

	before, err := h.service.GetPetByID(c.Request.Context(), companyID, petID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_pet_failed", "failed to get pet")
		return
	}

	var req updatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	if !hasPetUpdatePayload(req) {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "empty_update_payload", "at least one field must be provided")
		return
	}

	ownerID, err := parseOptionalUUID(req.OwnerID)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_owner_id", "invalid owner_id")
		return
	}
	guardianIDs, err := parseOptionalUUIDSlice(req.GuardianIDs)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_guardian_ids", "invalid guardian_ids")
		return
	}
	sex, err := parseOptionalPetSex(req.Sex)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_sex", "invalid pet sex")
		return
	}

	size, err := parseOptionalPetSize(req.Size)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_size", "invalid pet size")
		return
	}

	kind, err := parseOptionalPetKind(req.Kind)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_kind", "invalid pet kind")
		return
	}

	temperament, err := parseOptionalPetTemperament(req.Temperament)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_temperament", "invalid pet temperament")
		return
	}

	birthDate, err := parseOptionalDatePointer(req.BirthDate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_birth_date", "invalid birth_date")
		return
	}

	imageURL, err := h.resolveOptionalUploadObjectKey(c.Request.Context(), req.UploadKey, req.ImageURL)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "invalid_upload_object_key", "invalid upload_object_key")
		return
	}

	microchipDate, err := parseOptionalDatePointer(req.MicrochipExpirationDate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_microchip_expiration_date", "invalid microchip expiration date")
		return
	}

	item, err := h.service.UpdatePet(c.Request.Context(), service.UpdatePetInput{
		CompanyID:               companyID,
		PetID:                   petID,
		OwnerID:                 ownerID,
		GuardianIDs:             guardianIDs,
		Name:                    parseOptionalTrimmed(req.Name),
		Race:                    parseOptionalTrimmed(req.Race),
		Color:                   parseOptionalTrimmed(req.Color),
		Sex:                     sex,
		Size:                    size,
		Kind:                    kind,
		Temperament:             temperament,
		ImageURL:                imageURL,
		BirthDate:               birthDate,
		IsActive:                req.IsActive,
		IsDeceased:              req.IsDeceased,
		IsVaccinated:            req.IsVaccinated,
		IsNeutered:              req.IsNeutered,
		IsMicrochipped:          req.IsMicrochipped,
		MicrochipNumber:         parseOptionalTrimmed(req.MicrochipNumber),
		MicrochipExpirationDate: microchipDate,
		Notes:                   parseOptionalTrimmed(req.Notes),
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_pet_failed", "failed to update pet")
		return
	}

	guardians, err := h.service.GetPetGuardians(c.Request.Context(), companyID, item.ID)
	if err != nil {
		middleware.JSONError(c, http.StatusInternalServerError, "get_guardians_failed", "failed to get pet guardians")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionUpdate,
		EntityTable: "pets",
		EntityID:    item.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData:     item,
	})

	middleware.JSONData(c, http.StatusOK, gin.H{
		"data": buildPetResponseData(item, guardians),
	})
}

// Delete godoc
// @Summary Delete pet
// @Description Soft deletes a pet from the authenticated tenant.
// @Tags pets
// @Security BearerAuth
// @Param id path string true "Pet ID"
// @Success 204
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /pets/{id} [delete]
func (h *PetHandler) Delete(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	petID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_id", "invalid pet id")
		return
	}

	before, err := h.service.GetPetByID(c.Request.Context(), companyID, petID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_pet_failed", "failed to get pet")
		return
	}

	if err := h.service.DeletePet(c.Request.Context(), companyID, petID); err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "delete_pet_failed", "failed to delete pet")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionDelete,
		EntityTable: "pets",
		EntityID:    before.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData: gin.H{
			"id":        before.ID,
			"deleted":   true,
			"is_active": false,
		},
	})

	c.Status(http.StatusNoContent)
}

func parsePetSize(raw string) (sqlc.PetSize, error) {
	value := sqlc.PetSize(strings.TrimSpace(raw))
	switch value {
	case sqlc.PetSizeSmall, sqlc.PetSizeMedium, sqlc.PetSizeLarge, sqlc.PetSizeGiant:
		return value, nil
	default:
		return "", apperror.ErrUnprocessableEntity
	}
}

func parseOptionalPetSize(raw *string) (*sqlc.PetSize, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	value, err := parsePetSize(trimmed)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parsePetKind(raw string) (sqlc.PetKind, error) {
	value := sqlc.PetKind(strings.TrimSpace(raw))
	switch value {
	case sqlc.PetKindDog, sqlc.PetKindCat, sqlc.PetKindBird, sqlc.PetKindFish, sqlc.PetKindReptile, sqlc.PetKindRodent, sqlc.PetKindRabbit, sqlc.PetKindOther:
		return value, nil
	default:
		return "", apperror.ErrUnprocessableEntity
	}
}

func parsePetSex(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	switch trimmed {
	case "M", "F":
		return trimmed, nil
	default:
		return "", apperror.ErrUnprocessableEntity
	}
}

func parseOptionalPetSex(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	value, err := parsePetSex(trimmed)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseRequiredTrimmed(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", apperror.ErrUnprocessableEntity
	}
	return trimmed, nil
}

func parseOptionalPetKind(raw *string) (*sqlc.PetKind, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	value, err := parsePetKind(trimmed)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parsePetTemperament(raw string) (sqlc.PetTemperament, error) {
	value := sqlc.PetTemperament(strings.TrimSpace(raw))
	switch value {
	case sqlc.PetTemperamentCalm, sqlc.PetTemperamentNervous, sqlc.PetTemperamentAggressive, sqlc.PetTemperamentPlayful, sqlc.PetTemperamentLoving:
		return value, nil
	default:
		return "", apperror.ErrUnprocessableEntity
	}
}

func parseOptionalPetTemperament(raw *string) (*sqlc.PetTemperament, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	value, err := parsePetTemperament(trimmed)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func buildPetResponseData(item sqlc.GetPetDetailByIDAndCompanyIDRow, guardians []sqlc.ListPetGuardiansByPetIDRow) gin.H {
	return gin.H{
		"id":                        uuidToString(item.ID),
		"owner_id":                  uuidToString(item.OwnerID),
		"company_id":                uuidToString(item.CompanyID),
		"owner_person_id":           uuidToString(item.OwnerPersonID),
		"owner_name":                item.OwnerName,
		"owner_short_name":          item.OwnerShortName,
		"owner_image_url":           nullableText(item.OwnerImageUrl),
		"owner_email":               item.OwnerEmail,
		"owner_cellphone":           item.OwnerCellphone,
		"owner_has_whatsapp":        item.OwnerHasWhatsapp,
		"name":                      item.Name,
		"race":                      item.Race,
		"color":                     item.Color,
		"sex":                       item.Sex,
		"size":                      item.Size,
		"kind":                      item.Kind,
		"temperament":               item.Temperament,
		"image_url":                 nullableText(item.ImageUrl),
		"birth_date":                nullableDate(item.BirthDate),
		"is_active":                 item.IsActive,
		"is_deceased":               item.IsDeceased,
		"is_vaccinated":             item.IsVaccinated,
		"is_neutered":               item.IsNeutered,
		"is_microchipped":           item.IsMicrochipped,
		"microchip_number":          nullableText(item.MicrochipNumber),
		"microchip_expiration_date": nullableDate(item.MicrochipExpirationDate),
		"notes":                     nullableText(item.Notes),
		"created_at":                formatTimestamptz(item.CreatedAt),
		"updated_at":                nullableTimestamptz(item.UpdatedAt),
		"deleted_at":                nullableTimestamptz(item.DeletedAt),
		"guardians":                 buildPetGuardianDocs(guardians),
	}
}

func buildPetGuardianDocs(items []sqlc.ListPetGuardiansByPetIDRow) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"pet_id":       uuidToString(item.PetID),
			"guardian_id":  uuidToString(item.GuardianID),
			"full_name":    item.FullName,
			"short_name":   item.ShortName,
			"image_url":    nullableText(item.ImageUrl),
			"email":        item.Email,
			"cellphone":    item.Cellphone,
			"has_whatsapp": item.HasWhatsapp,
		})
	}
	return result
}

func hasPetUpdatePayload(req updatePetRequest) bool {
	return req.OwnerID != nil ||
		req.GuardianIDs != nil ||
		req.Name != nil ||
		req.Race != nil ||
		req.Color != nil ||
		req.Sex != nil ||
		req.Size != nil ||
		req.Kind != nil ||
		req.Temperament != nil ||
		req.ImageURL != nil ||
		req.UploadKey != nil ||
		req.BirthDate != nil ||
		req.IsActive != nil ||
		req.IsDeceased != nil ||
		req.IsVaccinated != nil ||
		req.IsNeutered != nil ||
		req.IsMicrochipped != nil ||
		req.MicrochipNumber != nil ||
		req.MicrochipExpirationDate != nil ||
		req.Notes != nil
}

func parseOptionalUUIDSlice(raw *[]string) (*[]pgtype.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	items, err := parseUUIDSlice(*raw)
	if err != nil {
		return nil, err
	}
	return &items, nil
}

func (h *PetHandler) resolveUploadObjectKey(ctx context.Context, objectKey string, fallback string) (string, error) {
	trimmedKey := strings.TrimSpace(objectKey)
	if trimmedKey == "" {
		return fallback, nil
	}
	if h.uploadResolver == nil {
		return "", apperror.ErrServiceUnavailable
	}
	return h.uploadResolver.ResolveObjectKey(ctx, "pets", "image_url", trimmedKey)
}

func (h *PetHandler) resolveOptionalUploadObjectKey(ctx context.Context, objectKey *string, fallback *string) (*string, error) {
	if objectKey == nil {
		return parseOptionalTrimmed(fallback), nil
	}

	resolved, err := h.resolveUploadObjectKey(ctx, *objectKey, "")
	if err != nil {
		return nil, err
	}
	return &resolved, nil
}
