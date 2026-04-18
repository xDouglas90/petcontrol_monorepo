package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	Size        string `json:"size"`
	Kind        string `json:"kind"`
	Temperament string `json:"temperament"`
	ImageURL    string `json:"image_url"`
	UploadKey   string `json:"upload_object_key"`
	BirthDate   string `json:"birth_date"`
	Notes       string `json:"notes"`
}

type updatePetRequest struct {
	OwnerID     *string `json:"owner_id"`
	Name        *string `json:"name"`
	Size        *string `json:"size"`
	Kind        *string `json:"kind"`
	Temperament *string `json:"temperament"`
	ImageURL    *string `json:"image_url"`
	UploadKey   *string `json:"upload_object_key"`
	BirthDate   *string `json:"birth_date"`
	Notes       *string `json:"notes"`
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
	items, err := h.service.ListPetsByCompanyID(c.Request.Context(), companyID, params)
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

	item, err := h.service.GetPetByID(c.Request.Context(), companyID, petID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_pet_failed", "failed to get pet")
		return
	}

	middleware.JSONData(c, http.StatusOK, item)
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

	item, err := h.service.CreatePet(c.Request.Context(), service.CreatePetInput{
		CompanyID:   companyID,
		OwnerID:     ownerID,
		Name:        strings.TrimSpace(req.Name),
		Size:        size,
		Kind:        kind,
		Temperament: temperament,
		ImageURL:    textValue(imageURL),
		BirthDate:   birthDate,
		Notes:       textValue(strings.TrimSpace(req.Notes)),
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_pet_failed", "failed to create pet")
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

	middleware.JSONData(c, http.StatusCreated, item)
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

	item, err := h.service.UpdatePet(c.Request.Context(), service.UpdatePetInput{
		CompanyID:   companyID,
		PetID:       petID,
		OwnerID:     ownerID,
		Name:        parseOptionalTrimmed(req.Name),
		Size:        size,
		Kind:        kind,
		Temperament: temperament,
		ImageURL:    imageURL,
		BirthDate:   birthDate,
		Notes:       parseOptionalTrimmed(req.Notes),
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_pet_failed", "failed to update pet")
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

	middleware.JSONData(c, http.StatusOK, item)
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
	value, err := parsePetSize(*raw)
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

func parseOptionalPetKind(raw *string) (*sqlc.PetKind, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := parsePetKind(*raw)
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
	value, err := parsePetTemperament(*raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func hasPetUpdatePayload(req updatePetRequest) bool {
	return req.OwnerID != nil ||
		req.Name != nil ||
		req.Size != nil ||
		req.Kind != nil ||
		req.Temperament != nil ||
		req.ImageURL != nil ||
		req.UploadKey != nil ||
		req.BirthDate != nil ||
		req.Notes != nil
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
