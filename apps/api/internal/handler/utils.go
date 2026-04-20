package handler

import (
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

func parseOptionalTrimmed(raw *string) *string {
	if raw == nil {
		return nil
	}
	value := strings.TrimSpace(*raw)
	return &value
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
