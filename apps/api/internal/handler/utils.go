package handler

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
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
