package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

func parseUUID(raw string) (pgtype.UUID, error) {
	var res pgtype.UUID
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return res, err
	}
	copy(res.Bytes[:], parsed[:])
	res.Valid = true
	return res, nil
}

func parseOptionalUUID(raw *string) (*pgtype.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	value, err := parseUUID(*raw)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseUUIDSlice(values []string) ([]pgtype.UUID, error) {
	if values == nil {
		return nil, nil
	}

	result := make([]pgtype.UUID, 0, len(values))
	for _, value := range values {
		parsed, err := parseUUID(value)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}

func parseOptionalTrimmed(raw *string) *string {
	if raw == nil {
		return nil
	}
	value := strings.TrimSpace(*raw)
	if value == "" {
		return nil
	}
	return &value
}

func parseOptionalBool(raw string) *bool {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil
	}
	return &parsed
}

func boolValueOrDefault(value *bool, defaultValue bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{Bool: defaultValue, Valid: true}
	}
	return pgtype.Bool{Bool: *value, Valid: true}
}

func uuidToString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	parsed, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return ""
	}
	return parsed.String()
}

func textValue(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func nullableText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func textPointer(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: strings.TrimSpace(*s), Valid: true}
}

func formatTime(value pgtype.Time) string {
	if !value.Valid {
		return ""
	}

	totalMicroseconds := value.Microseconds
	hours := totalMicroseconds / int64(time.Hour/time.Microsecond)
	minutes := (totalMicroseconds / int64(time.Minute/time.Microsecond)) % 60

	return time.Date(0, time.January, 1, int(hours), int(minutes), 0, 0, time.UTC).Format("15:04")
}

func weekDaysToStrings(values []sqlc.WeekDay) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, string(value))
	}
	return items
}

func mapCompanyUsers(items []service.CompanyUserWithProfile) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"id":         uuidToString(item.ID),
			"company_id": uuidToString(item.CompanyID),
			"user_id":    uuidToString(item.UserID),
			"kind":       string(item.Kind),
			"role":       string(item.Role),
			"is_owner":   item.IsOwner,
			"is_active":  item.IsActive,
			"full_name":  item.FullName,
			"short_name": item.ShortName,
			"image_url":  item.ImageURL,
			"joined_at":  formatTimestamptz(item.JoinedAt),
			"left_at":    nullableTimestamptz(item.LeftAt),
		})
	}
	return result
}

func formatTimestamptz(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(time.RFC3339)
}

func nullableTimestamptz(value pgtype.Timestamptz) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.Format(time.RFC3339)
	return &formatted
}

func mapAdminSystemChatMessages(items []service.AdminSystemChatMessage) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, mapAdminSystemChatMessage(item))
	}
	return result
}

func mapAdminSystemChatMessage(item service.AdminSystemChatMessage) map[string]any {
	return map[string]any{
		"id":               uuidToString(item.ID),
		"conversation_id":  uuidToString(item.ConversationID),
		"company_id":       uuidToString(item.CompanyID),
		"sender_user_id":   uuidToString(item.SenderUserID),
		"sender_name":      item.SenderName,
		"sender_role":      string(item.SenderRole),
		"sender_image_url": item.SenderImageURL,
		"body":             item.Body,
		"created_at":       formatTimestamptz(item.CreatedAt),
	}
}

func mapPeople(items []service.PeopleListItem) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"id":                uuidToString(item.ID),
			"company_id":        uuidToString(item.CompanyID),
			"company_person_id": uuidToString(item.CompanyPersonID),
			"kind":              string(item.Kind),
			"full_name":         item.FullName,
			"short_name":        item.ShortName,
			"email":             item.Email,
			"image_url":         item.ImageURL,
			"cpf":               item.CPF,
			"has_system_user":   item.HasSystemUser,
			"is_active":         item.IsActive,
			"created_at":        formatTimestamptz(item.CreatedAt),
			"updated_at":        nullableTimestamptz(item.UpdatedAt),
		})
	}
	return result
}

func mapPersonDetail(item service.PeopleDetail) map[string]any {
	result := map[string]any{
		"id":                 uuidToString(item.ID),
		"company_id":         uuidToString(item.CompanyID),
		"company_person_id":  uuidToString(item.CompanyPersonID),
		"kind":               string(item.Kind),
		"full_name":          item.FullName,
		"short_name":         item.ShortName,
		"image_url":          item.ImageURL,
		"cpf":                item.CPF,
		"has_system_user":    item.HasSystemUser,
		"is_active":          item.IsActive,
		"created_at":         formatTimestamptz(item.CreatedAt),
		"updated_at":         nullableTimestamptz(item.UpdatedAt),
		"contact":            nil,
		"address":            nil,
		"finance":            nil,
		"client_details":     nil,
		"employee_details":   nil,
		"employee_documents": nil,
		"employee_benefits":  nil,
		"linked_user":        nil,
		"guardian_pets":      []map[string]any{},
		"gender_identity":    nil,
		"marital_status":     nil,
		"birth_date":         nil,
	}

	if item.Identification != nil {
		result["gender_identity"] = string(item.Identification.GenderIdentity)
		result["marital_status"] = string(item.Identification.MaritalStatus)
		result["birth_date"] = nullableDate(item.Identification.BirthDate)
	}

	if item.Contact != nil {
		result["contact"] = map[string]any{
			"email":             item.Contact.Email,
			"phone":             nullableText(item.Contact.Phone),
			"cellphone":         item.Contact.Cellphone,
			"has_whatsapp":      item.Contact.HasWhatsapp,
			"instagram_user":    nullableText(item.Contact.InstagramUser),
			"emergency_contact": nullableText(item.Contact.EmergencyContact),
			"emergency_phone":   nullableText(item.Contact.EmergencyPhone),
		}
	}

	if item.Address != nil {
		result["address"] = map[string]any{
			"zip_code":   item.Address.Address.ZipCode,
			"street":     item.Address.Address.Street,
			"number":     item.Address.Address.Number,
			"complement": nullableText(item.Address.Address.Complement),
			"district":   item.Address.Address.District,
			"city":       item.Address.Address.City,
			"state":      item.Address.Address.State,
			"country":    item.Address.Address.Country,
			"label":      nullableText(item.Address.Link.Label),
			"is_main":    item.Address.Link.IsMain,
		}
	}

	if item.ClientDetails != nil {
		result["client_details"] = map[string]any{
			"client_since": nullableDate(item.ClientDetails.ClientSince),
			"notes":        nullableText(item.ClientDetails.Notes),
		}
	}

	if item.Finance != nil {
		result["finance"] = map[string]any{
			"bank_name":          item.Finance.BankName,
			"bank_code":          nullableText(item.Finance.BankCode),
			"bank_branch":        item.Finance.BankBranch,
			"bank_account":       item.Finance.BankAccount,
			"bank_account_digit": item.Finance.BankAccountDigit,
			"bank_account_type":  string(item.Finance.BankAccountType),
			"has_pix":            item.Finance.HasPix,
			"pix_key":            nullableText(item.Finance.PixKey),
			"pix_key_type":       nullablePixKeyKind(item.Finance.PixKeyType),
			"is_primary":         item.Finance.IsPrimary,
		}
	}

	if item.EmployeeDetails != nil {
		result["employee_details"] = map[string]any{
			"company_employee_id": uuidToString(item.EmployeeDetails.CompanyEmployeeID),
			"role":                item.EmployeeDetails.Role,
			"admission_date":      nullableDate(item.EmployeeDetails.AdmissionDate),
			"resignation_date":    nullableDate(item.EmployeeDetails.ResignationDate),
			"salary":              item.EmployeeDetails.Salary,
		}
	}

	if item.EmployeeDocuments != nil {
		result["employee_documents"] = map[string]any{
			"rg":           item.EmployeeDocuments.Rg,
			"issuing_body": item.EmployeeDocuments.IssuingBody,
			"ctps":         item.EmployeeDocuments.Ctps,
			"pis":          item.EmployeeDocuments.Pis,
			"graduation":   string(item.EmployeeDocuments.Graduation),
		}
	}

	if item.EmployeeBenefits != nil {
		result["employee_benefits"] = map[string]any{
			"meal_ticket":             item.EmployeeBenefits.MealTicket,
			"meal_ticket_value":       numericToString(item.EmployeeBenefits.MealTicketValue),
			"transport_voucher":       item.EmployeeBenefits.TransportVoucher,
			"transport_voucher_qty":   item.EmployeeBenefits.TransportVoucherQty,
			"transport_voucher_value": numericToString(item.EmployeeBenefits.TransportVoucherValue),
			"valid_from":              formatDate(item.EmployeeBenefits.ValidFrom),
			"valid_until":             nullableDate(item.EmployeeBenefits.ValidUntil),
		}
	}

	if item.LinkedUser != nil {
		result["linked_user"] = map[string]any{
			"user_id":   uuidToString(item.LinkedUser.UserID),
			"email":     item.LinkedUser.Email,
			"role":      string(item.LinkedUser.Role),
			"kind":      string(item.LinkedUser.Kind),
			"is_active": item.LinkedUser.IsActive,
			"is_owner":  item.LinkedUser.IsOwner,
			"joined_at": formatTimestamptz(item.LinkedUser.JoinedAt),
		}
	}

	if len(item.GuardianPets) > 0 {
		guardianPets := make([]map[string]any, 0, len(item.GuardianPets))
		for _, pet := range item.GuardianPets {
			guardianPets = append(guardianPets, map[string]any{
				"pet_id":     uuidToString(pet.PetID),
				"name":       pet.Name,
				"kind":       string(pet.Kind),
				"size":       string(pet.Size),
				"owner_name": pet.OwnerName,
			})
		}
		result["guardian_pets"] = guardianPets
	}

	return result
}

func nullableDate(value pgtype.Date) *string {
	if !value.Valid {
		return nil
	}
	formatted := formatDate(value)
	return &formatted
}

func formatDate(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func nullablePixKeyKind(value sqlc.NullPixKeyKind) *string {
	if !value.Valid {
		return nil
	}
	formatted := string(value.PixKeyKind)
	return &formatted
}
