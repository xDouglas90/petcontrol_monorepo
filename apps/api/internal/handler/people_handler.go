package handler

import (
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

type PeopleHandler struct {
	service *service.PeopleService
}

type personAddressRequest struct {
	ZipCode    string `json:"zip_code"`
	Street     string `json:"street"`
	Number     string `json:"number"`
	Complement string `json:"complement"`
	District   string `json:"district"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	Label      string `json:"label"`
}

type personEmploymentRequest struct {
	Role            string `json:"role"`
	AdmissionDate   string `json:"admission_date"`
	ResignationDate string `json:"resignation_date"`
	Salary          string `json:"salary"`
}

type personEmployeeDocumentsRequest struct {
	RG          string `json:"rg"`
	IssuingBody string `json:"issuing_body"`
	IssuingDate string `json:"issuing_date"`
	CTPS        string `json:"ctps"`
	CTPSSeries  string `json:"ctps_series"`
	CTPSState   string `json:"ctps_state"`
	PIS         string `json:"pis"`
	Graduation  string `json:"graduation"`
}

type personEmployeeBenefitsRequest struct {
	MealTicket            bool   `json:"meal_ticket"`
	MealTicketValue       string `json:"meal_ticket_value"`
	TransportVoucher      bool   `json:"transport_voucher"`
	TransportVoucherQty   int16  `json:"transport_voucher_qty"`
	TransportVoucherValue string `json:"transport_voucher_value"`
	ValidFrom             string `json:"valid_from"`
	ValidUntil            string `json:"valid_until"`
}

type personFinanceRequest struct {
	BankName         string `json:"bank_name"`
	BankCode         string `json:"bank_code"`
	BankBranch       string `json:"bank_branch"`
	BankAccount      string `json:"bank_account"`
	BankAccountDigit string `json:"bank_account_digit"`
	BankAccountType  string `json:"bank_account_type"`
	HasPix           bool   `json:"has_pix"`
	PixKey           string `json:"pix_key"`
	PixKeyType       string `json:"pix_key_type"`
}

type createPersonRequest struct {
	Kind             string                          `json:"kind"`
	FullName         string                          `json:"full_name"`
	ShortName        string                          `json:"short_name"`
	GenderIdentity   string                          `json:"gender_identity"`
	MaritalStatus    string                          `json:"marital_status"`
	BirthDate        string                          `json:"birth_date"`
	CPF              string                          `json:"cpf"`
	Email            string                          `json:"email"`
	Phone            string                          `json:"phone"`
	Cellphone        string                          `json:"cellphone"`
	HasWhatsapp      bool                            `json:"has_whatsapp"`
	HasSystemUser    bool                            `json:"has_system_user"`
	IsActive         *bool                           `json:"is_active"`
	Address          *personAddressRequest           `json:"address"`
	ClientSince      string                          `json:"client_since"`
	Notes            string                          `json:"notes"`
	PetIDs           []string                        `json:"pet_ids"`
	Employment       *personEmploymentRequest        `json:"employment"`
	Finance          *personFinanceRequest           `json:"finance"`
	EmployeeDocs     *personEmployeeDocumentsRequest `json:"employee_documents"`
	EmployeeBenefits *personEmployeeBenefitsRequest  `json:"employee_benefits"`
}

type updatePersonRequest struct {
	FullName         *string                         `json:"full_name"`
	ShortName        *string                         `json:"short_name"`
	GenderIdentity   *string                         `json:"gender_identity"`
	MaritalStatus    *string                         `json:"marital_status"`
	BirthDate        *string                         `json:"birth_date"`
	CPF              *string                         `json:"cpf"`
	Email            *string                         `json:"email"`
	Phone            *string                         `json:"phone"`
	Cellphone        *string                         `json:"cellphone"`
	HasWhatsapp      *bool                           `json:"has_whatsapp"`
	HasSystemUser    *bool                           `json:"has_system_user"`
	IsActive         *bool                           `json:"is_active"`
	Address          *personAddressRequest           `json:"address"`
	ClientSince      *string                         `json:"client_since"`
	Notes            *string                         `json:"notes"`
	PetIDs           []string                        `json:"pet_ids"`
	Employment       *personEmploymentRequest        `json:"employment"`
	Finance          *personFinanceRequest           `json:"finance"`
	EmployeeDocs     *personEmployeeDocumentsRequest `json:"employee_documents"`
	EmployeeBenefits *personEmployeeBenefitsRequest  `json:"employee_benefits"`
}

func NewPeopleHandler(service *service.PeopleService) *PeopleHandler {
	return &PeopleHandler{service: service}
}

// List godoc
// @Summary List people
// @Description Returns tenant-scoped people linked to the authenticated company.
// @Tags people
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /people [get]
func (h *PeopleHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok {
		middleware.JSONError(c, http.StatusUnauthorized, "auth_claims_required", "auth claims required")
		return
	}
	if claims.Role != string(sqlc.UserRoleTypeAdmin) && claims.Role != string(sqlc.UserRoleTypeSystem) {
		middleware.JSONError(c, http.StatusForbidden, "people_access_required", "people access required")
		return
	}

	kind, err := parseOptionalPeopleKind(c.Query("kind"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_person_kind", "invalid person kind")
		return
	}

	params := pagination.ParseParams(c)
	items, err := h.service.ListPeopleByCompanyID(c.Request.Context(), companyID, params.Search)
	if err != nil {
		middleware.JSONError(c, http.StatusInternalServerError, "list_people_failed", "failed to list people")
		return
	}

	items = filterPeopleByRole(items, claims.Role)
	items = filterPeopleByKind(items, kind)
	total := len(items)
	items = paginatePeople(items, params)

	middleware.JSONPaginated(c, http.StatusOK, mapPeople(items), pagination.NewMeta(total, params.Page, params.Limit))
}

// GetByID godoc
// @Summary Get person by ID
// @Description Returns tenant-scoped detail for one person linked to the authenticated company.
// @Tags people
// @Security BearerAuth
// @Produce json
// @Param id path string true "Person ID"
// @Success 200 {object} map[string]any
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /people/{id} [get]
func (h *PeopleHandler) GetByID(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok {
		middleware.JSONError(c, http.StatusUnauthorized, "auth_claims_required", "auth claims required")
		return
	}
	if claims.Role != string(sqlc.UserRoleTypeAdmin) && claims.Role != string(sqlc.UserRoleTypeSystem) {
		middleware.JSONError(c, http.StatusForbidden, "people_access_required", "people access required")
		return
	}

	personID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_person_id", "invalid person id")
		return
	}

	item, err := h.service.GetPersonDetailByID(c.Request.Context(), companyID, personID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_person_failed", "failed to get person")
		return
	}

	if claims.Role == string(sqlc.UserRoleTypeSystem) &&
		item.Kind != sqlc.PersonKindClient &&
		item.Kind != sqlc.PersonKindSupplier {
		middleware.JSONError(c, http.StatusForbidden, "people_access_required", "people access required")
		return
	}

	middleware.JSONData(c, http.StatusOK, mapPersonDetail(item))
}

// Create godoc
// @Summary Create person
// @Description Creates a tenant-scoped person for supported kinds.
// @Tags people
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} map[string]any
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 409 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /people [post]
func (h *PeopleHandler) Create(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok {
		middleware.JSONError(c, http.StatusUnauthorized, "auth_claims_required", "auth claims required")
		return
	}
	if claims.Role != string(sqlc.UserRoleTypeAdmin) && claims.Role != string(sqlc.UserRoleTypeSystem) {
		middleware.JSONError(c, http.StatusForbidden, "people_access_required", "people access required")
		return
	}

	var req createPersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	kind, err := parsePeopleKind(req.Kind)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_person_kind", "invalid person kind")
		return
	}
	if err := ensurePeopleWriteAccess(claims.Role, kind); err != nil {
		middleware.JSONError(c, http.StatusForbidden, "people_access_required", "people access required")
		return
	}

	birthDate, err := parseRequiredDate(req.BirthDate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_birth_date", "invalid birth_date")
		return
	}

	genderIdentity, err := parseGenderIdentity(req.GenderIdentity)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_gender_identity", "invalid gender_identity")
		return
	}

	maritalStatus, err := parseMaritalStatus(req.MaritalStatus)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_marital_status", "invalid marital_status")
		return
	}

	clientSince, err := parseOptionalDatePointer(stringToOptionalTrimmed(req.ClientSince))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_client_since", "invalid client_since")
		return
	}

	employment, err := mapEmploymentRequest(req.Employment)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_employment", "invalid employment payload")
		return
	}

	finance, err := mapFinanceRequest(req.Finance)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_finance", "invalid finance payload")
		return
	}

	employeeDocs, err := mapEmployeeDocumentsRequest(req.EmployeeDocs)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_employee_documents", "invalid employee documents payload")
		return
	}

	employeeBenefits, err := mapEmployeeBenefitsRequest(req.EmployeeBenefits)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_employee_benefits", "invalid employee benefits payload")
		return
	}

	petIDs, err := parseUUIDSlice(req.PetIDs)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_ids", "invalid pet_ids payload")
		return
	}

	item, err := h.service.CreatePerson(c.Request.Context(), service.CreatePersonInput{
		CompanyID:        companyID,
		ActorUserID:      mustParseUUID(claims.UserID),
		ActorRole:        sqlc.UserRoleType(claims.Role),
		Kind:             kind,
		FullName:         strings.TrimSpace(req.FullName),
		ShortName:        strings.TrimSpace(req.ShortName),
		GenderIdentity:   genderIdentity,
		MaritalStatus:    maritalStatus,
		BirthDate:        birthDate,
		CPF:              onlyDigits(req.CPF),
		Email:            strings.TrimSpace(req.Email),
		Phone:            stringToOptionalTrimmed(req.Phone),
		Cellphone:        strings.TrimSpace(req.Cellphone),
		HasWhatsapp:      req.HasWhatsapp,
		HasSystemUser:    req.HasSystemUser,
		IsActive:         req.IsActive == nil || *req.IsActive,
		Address:          mapAddressRequest(req.Address),
		ClientSince:      clientSince,
		Notes:            stringToOptionalTrimmed(req.Notes),
		PetIDs:           petIDs,
		Employment:       employment,
		Finance:          finance,
		EmployeeDocs:     employeeDocs,
		EmployeeBenefits: employeeBenefits,
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_person_failed", "failed to create person")
		return
	}

	middleware.JSONData(c, http.StatusCreated, mapPersonDetail(item))
}

// Update godoc
// @Summary Update person
// @Description Updates a tenant-scoped person for supported kinds.
// @Tags people
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Person ID"
// @Success 200 {object} map[string]any
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /people/{id} [patch]
func (h *PeopleHandler) Update(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return
	}

	claims, ok := middleware.GetClaims(c)
	if !ok {
		middleware.JSONError(c, http.StatusUnauthorized, "auth_claims_required", "auth claims required")
		return
	}
	if claims.Role != string(sqlc.UserRoleTypeAdmin) && claims.Role != string(sqlc.UserRoleTypeSystem) {
		middleware.JSONError(c, http.StatusForbidden, "people_access_required", "people access required")
		return
	}

	personID, err := parseUUID(c.Param("id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_person_id", "invalid person id")
		return
	}

	current, err := h.service.GetPersonDetailByID(c.Request.Context(), companyID, personID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_person_failed", "failed to get person")
		return
	}
	if err := ensurePeopleWriteAccess(claims.Role, current.Kind); err != nil {
		middleware.JSONError(c, http.StatusForbidden, "people_access_required", "people access required")
		return
	}

	var req updatePersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	birthDate, err := parseOptionalDatePointer(req.BirthDate)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_birth_date", "invalid birth_date")
		return
	}

	genderIdentity, err := parseOptionalGenderIdentity(req.GenderIdentity)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_gender_identity", "invalid gender_identity")
		return
	}

	maritalStatus, err := parseOptionalMaritalStatus(req.MaritalStatus)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_marital_status", "invalid marital_status")
		return
	}

	clientSince, err := parseOptionalDatePointer(req.ClientSince)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_client_since", "invalid client_since")
		return
	}

	employment, err := mapEmploymentRequest(req.Employment)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_employment", "invalid employment payload")
		return
	}

	finance, err := mapFinanceRequest(req.Finance)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_finance", "invalid finance payload")
		return
	}

	employeeDocs, err := mapEmployeeDocumentsRequest(req.EmployeeDocs)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_employee_documents", "invalid employee documents payload")
		return
	}

	employeeBenefits, err := mapEmployeeBenefitsRequest(req.EmployeeBenefits)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_employee_benefits", "invalid employee benefits payload")
		return
	}

	petIDs, err := parseUUIDSlice(req.PetIDs)
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_pet_ids", "invalid pet_ids payload")
		return
	}

	item, err := h.service.UpdatePerson(c.Request.Context(), service.UpdatePersonInput{
		CompanyID:        companyID,
		ActorUserID:      mustParseUUID(claims.UserID),
		ActorRole:        sqlc.UserRoleType(claims.Role),
		PersonID:         personID,
		FullName:         parseOptionalTrimmed(req.FullName),
		ShortName:        parseOptionalTrimmed(req.ShortName),
		GenderIdentity:   genderIdentity,
		MaritalStatus:    maritalStatus,
		BirthDate:        birthDate,
		CPF:              sanitizeOptionalDigits(req.CPF),
		Email:            parseOptionalTrimmed(req.Email),
		Phone:            parseOptionalTrimmed(req.Phone),
		Cellphone:        parseOptionalTrimmed(req.Cellphone),
		HasWhatsapp:      req.HasWhatsapp,
		HasSystemUser:    req.HasSystemUser,
		IsActive:         req.IsActive,
		Address:          mapAddressRequest(req.Address),
		ClientSince:      clientSince,
		Notes:            parseOptionalTrimmed(req.Notes),
		PetIDs:           petIDs,
		HasPetIDs:        req.PetIDs != nil,
		Employment:       employment,
		Finance:          finance,
		EmployeeDocs:     employeeDocs,
		EmployeeBenefits: employeeBenefits,
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_person_failed", "failed to update person")
		return
	}

	middleware.JSONData(c, http.StatusOK, mapPersonDetail(item))
}

func filterPeopleByRole(items []service.PeopleListItem, role string) []service.PeopleListItem {
	if role != string(sqlc.UserRoleTypeSystem) {
		return items
	}

	filtered := make([]service.PeopleListItem, 0, len(items))
	for _, item := range items {
		if item.Kind == sqlc.PersonKindClient || item.Kind == sqlc.PersonKindSupplier {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func mapFinanceRequest(req *personFinanceRequest) (*service.PersonFinanceInput, error) {
	if req == nil {
		return nil, nil
	}

	bankAccountType, err := parseBankAccountKind(req.BankAccountType)
	if err != nil {
		return nil, err
	}

	pixKeyType, err := parseOptionalPixKeyKind(stringToOptionalTrimmed(req.PixKeyType))
	if err != nil {
		return nil, err
	}

	return &service.PersonFinanceInput{
		BankName:         strings.TrimSpace(req.BankName),
		BankCode:         stringToOptionalTrimmed(req.BankCode),
		BankBranch:       strings.TrimSpace(req.BankBranch),
		BankAccount:      strings.TrimSpace(req.BankAccount),
		BankAccountDigit: strings.TrimSpace(req.BankAccountDigit),
		BankAccountType:  bankAccountType,
		HasPix:           req.HasPix,
		PixKey:           stringToOptionalTrimmed(req.PixKey),
		PixKeyType:       pixKeyType,
	}, nil
}

func parseBankAccountKind(raw string) (sqlc.BankAccountKind, error) {
	value := sqlc.BankAccountKind(strings.TrimSpace(raw))
	switch value {
	case sqlc.BankAccountKindChecking,
		sqlc.BankAccountKindSavings,
		sqlc.BankAccountKindSalary:
		return value, nil
	default:
		return "", apperror.ErrUnprocessableEntity
	}
}

func parseOptionalPixKeyKind(raw *string) (*sqlc.PixKeyKind, error) {
	if raw == nil {
		return nil, nil
	}

	value := sqlc.PixKeyKind(strings.TrimSpace(*raw))
	switch value {
	case sqlc.PixKeyKindCpf,
		sqlc.PixKeyKindCnpj,
		sqlc.PixKeyKindEmail,
		sqlc.PixKeyKindPhone,
		sqlc.PixKeyKindRandom:
		return &value, nil
	default:
		return nil, apperror.ErrUnprocessableEntity
	}
}

func filterPeopleByKind(items []service.PeopleListItem, kind *sqlc.PersonKind) []service.PeopleListItem {
	if kind == nil {
		return items
	}

	filtered := make([]service.PeopleListItem, 0, len(items))
	for _, item := range items {
		if item.Kind == *kind {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func paginatePeople(items []service.PeopleListItem, params pagination.Params) []service.PeopleListItem {
	if params.Offset >= len(items) {
		return []service.PeopleListItem{}
	}

	end := params.Offset + params.Limit
	if end > len(items) {
		end = len(items)
	}

	return items[params.Offset:end]
}

func parseOptionalPeopleKind(raw string) (*sqlc.PersonKind, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	kind, err := parsePeopleKind(trimmed)
	if err != nil {
		return nil, err
	}

	return &kind, nil
}

func parsePeopleKind(raw string) (sqlc.PersonKind, error) {
	value := sqlc.PersonKind(strings.TrimSpace(raw))
	switch value {
	case sqlc.PersonKindClient,
		sqlc.PersonKindEmployee,
		sqlc.PersonKindOutsourcedEmployee,
		sqlc.PersonKindSupplier,
		sqlc.PersonKindGuardian,
		sqlc.PersonKindResponsible:
		return value, nil
	default:
		return "", apperror.ErrUnprocessableEntity
	}
}

func ensurePeopleWriteAccess(role string, kind sqlc.PersonKind) error {
	if role == string(sqlc.UserRoleTypeSystem) {
		if kind != sqlc.PersonKindClient && kind != sqlc.PersonKindSupplier {
			return apperror.ErrForbidden
		}
		return nil
	}

	if role == string(sqlc.UserRoleTypeAdmin) {
		if kind == sqlc.PersonKindClient ||
			kind == sqlc.PersonKindEmployee ||
			kind == sqlc.PersonKindOutsourcedEmployee ||
			kind == sqlc.PersonKindSupplier ||
			kind == sqlc.PersonKindGuardian ||
			kind == sqlc.PersonKindResponsible {
			return nil
		}
	}

	return apperror.ErrForbidden
}

func mapAddressRequest(req *personAddressRequest) *service.PersonAddressInput {
	if req == nil {
		return nil
	}

	return &service.PersonAddressInput{
		ZipCode:    strings.TrimSpace(req.ZipCode),
		Street:     strings.TrimSpace(req.Street),
		Number:     strings.TrimSpace(req.Number),
		Complement: stringToOptionalTrimmed(req.Complement),
		District:   strings.TrimSpace(req.District),
		City:       strings.TrimSpace(req.City),
		State:      strings.TrimSpace(req.State),
		Country:    defaultCountry(req.Country),
		Label:      stringToOptionalTrimmed(req.Label),
	}
}

func stringToOptionalTrimmed(value string) *string {
	return parseOptionalTrimmed(&value)
}

func defaultCountry(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "Brasil"
	}
	return trimmed
}

func mapEmploymentRequest(req *personEmploymentRequest) (*service.PersonEmploymentInput, error) {
	if req == nil {
		return nil, nil
	}

	admissionDate, err := parseRequiredDate(req.AdmissionDate)
	if err != nil {
		return nil, err
	}

	resignationDate, err := parseOptionalDatePointer(stringToOptionalTrimmed(req.ResignationDate))
	if err != nil {
		return nil, err
	}

	salary, err := parseRequiredNumeric(req.Salary)
	if err != nil {
		return nil, err
	}

	return &service.PersonEmploymentInput{
		Role:            strings.TrimSpace(req.Role),
		AdmissionDate:   admissionDate,
		ResignationDate: resignationDate,
		Salary:          salary,
	}, nil
}

func mapEmployeeDocumentsRequest(req *personEmployeeDocumentsRequest) (*service.PersonEmployeeDocumentsInput, error) {
	if req == nil {
		return nil, nil
	}

	issuingDate, err := parseRequiredDate(req.IssuingDate)
	if err != nil {
		return nil, err
	}

	graduation, err := parseGraduationLevel(req.Graduation)
	if err != nil {
		return nil, err
	}

	return &service.PersonEmployeeDocumentsInput{
		RG:          onlyDigits(req.RG),
		IssuingBody: strings.TrimSpace(req.IssuingBody),
		IssuingDate: issuingDate,
		CTPS:        onlyDigits(req.CTPS),
		CTPSSeries:  strings.TrimSpace(req.CTPSSeries),
		CTPSState:   strings.ToUpper(strings.TrimSpace(req.CTPSState)),
		PIS:         onlyDigits(req.PIS),
		Graduation:  graduation,
	}, nil
}

func mapEmployeeBenefitsRequest(req *personEmployeeBenefitsRequest) (*service.PersonEmployeeBenefitsInput, error) {
	if req == nil {
		return nil, nil
	}

	validFrom, err := parseRequiredDate(req.ValidFrom)
	if err != nil {
		return nil, err
	}

	validUntil, err := parseOptionalDatePointer(stringToOptionalTrimmed(req.ValidUntil))
	if err != nil {
		return nil, err
	}

	mealTicketValue, err := parseOptionalNumeric(req.MealTicketValue)
	if err != nil {
		return nil, err
	}

	transportVoucherValue, err := parseOptionalNumeric(req.TransportVoucherValue)
	if err != nil {
		return nil, err
	}

	return &service.PersonEmployeeBenefitsInput{
		MealTicket:            req.MealTicket,
		MealTicketValue:       mealTicketValue,
		TransportVoucher:      req.TransportVoucher,
		TransportVoucherQty:   req.TransportVoucherQty,
		TransportVoucherValue: transportVoucherValue,
		ValidFrom:             validFrom,
		ValidUntil:            validUntil,
	}, nil
}

func parseGraduationLevel(raw string) (sqlc.GraduationLevel, error) {
	value := sqlc.GraduationLevel(strings.TrimSpace(raw))
	switch value {
	case sqlc.GraduationLevelElementaryIncomplete,
		sqlc.GraduationLevelElementaryComplete,
		sqlc.GraduationLevelMiddleIncomplete,
		sqlc.GraduationLevelMiddleComplete,
		sqlc.GraduationLevelHighIncomplete,
		sqlc.GraduationLevelHighComplete,
		sqlc.GraduationLevelCollegeIncomplete,
		sqlc.GraduationLevelCollegeComplete,
		sqlc.GraduationLevelPostgraduateIncomplete,
		sqlc.GraduationLevelPostgraduateComplete,
		sqlc.GraduationLevelMasterIncomplete,
		sqlc.GraduationLevelMasterComplete,
		sqlc.GraduationLevelDoctorateIncomplete,
		sqlc.GraduationLevelDoctorateComplete:
		return value, nil
	default:
		return "", apperror.ErrUnprocessableEntity
	}
}

func mustParseUUID(raw string) pgtype.UUID {
	value, err := parseUUID(raw)
	if err != nil {
		return pgtype.UUID{}
	}
	return value
}
