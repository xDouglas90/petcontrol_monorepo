package handler

import (
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

// LoginRequestDoc documents the login payload for Swagger.
type LoginRequestDoc struct {
	Email    string `json:"email" example:"admin@petcontrol.local"`
	Password string `json:"password" example:"password123"`
}

// LoginResponseDoc documents the login response envelope for Swagger.
type LoginResponseDoc struct {
	Data service.LoginResult `json:"data"`
}

// ScheduleCreateRequestDoc documents schedule creation payload for Swagger.
type ScheduleCreateRequestDoc struct {
	ClientID     string  `json:"client_id" example:"11111111-1111-1111-1111-111111111111"`
	PetID        string  `json:"pet_id" example:"22222222-2222-2222-2222-222222222222"`
	ScheduledAt  string  `json:"scheduled_at" example:"2026-04-10T15:00:00Z"`
	EstimatedEnd *string `json:"estimated_end,omitempty" example:"2026-04-10T16:00:00Z"`
	Notes        string  `json:"notes" example:"Banho e tosa"`
	Status       string  `json:"status,omitempty" example:"waiting"`
	StatusNotes  string  `json:"status_notes,omitempty" example:"Aguardando confirmação"`
}

// ScheduleUpdateRequestDoc documents schedule update payload for Swagger.
type ScheduleUpdateRequestDoc struct {
	ClientID     *string `json:"client_id,omitempty" example:"11111111-1111-1111-1111-111111111111"`
	PetID        *string `json:"pet_id,omitempty" example:"22222222-2222-2222-2222-222222222222"`
	ScheduledAt  *string `json:"scheduled_at,omitempty" example:"2026-04-10T15:00:00Z"`
	EstimatedEnd *string `json:"estimated_end,omitempty" example:"2026-04-10T16:00:00Z"`
	Notes        *string `json:"notes,omitempty" example:"Confirmado com o cliente"`
	Status       *string `json:"status,omitempty" example:"confirmed"`
	StatusNotes  *string `json:"status_notes,omitempty" example:"Confirmado por telefone"`
}

// ScheduleListResponseDoc documents the list schedules response envelope for Swagger.
type ScheduleListResponseDoc struct {
	Data []ScheduleDoc `json:"data"`
}

// ScheduleItemResponseDoc documents single schedule response envelope for Swagger.
type ScheduleItemResponseDoc struct {
	Data ScheduleDoc `json:"data"`
}

// ScheduleHistoryResponseDoc documents schedule status history response envelope for Swagger.
type ScheduleHistoryResponseDoc struct {
	Data []ScheduleHistoryItemDoc `json:"data"`
}

// ScheduleDoc describes the public schedule shape returned by API responses.
type ScheduleDoc struct {
	ID            string  `json:"id" example:"33333333-3333-3333-3333-333333333333"`
	CompanyID     string  `json:"company_id" example:"11111111-1111-1111-1111-111111111111"`
	ClientID      string  `json:"client_id" example:"44444444-4444-4444-4444-444444444444"`
	PetID         string  `json:"pet_id" example:"55555555-5555-5555-5555-555555555555"`
	ScheduledAt   string  `json:"scheduled_at" example:"2026-04-10T15:00:00Z"`
	EstimatedEnd  *string `json:"estimated_end,omitempty" example:"2026-04-10T16:00:00Z"`
	Notes         *string `json:"notes,omitempty" example:"Banho e tosa"`
	CurrentStatus string  `json:"current_status" example:"confirmed"`
	CreatedAt     string  `json:"created_at" example:"2026-04-10T10:00:00Z"`
	UpdatedAt     *string `json:"updated_at,omitempty" example:"2026-04-10T11:00:00Z"`
	DeletedAt     *string `json:"deleted_at,omitempty" example:"2026-04-11T11:00:00Z"`
}

// ScheduleHistoryItemDoc describes one schedule status transition.
type ScheduleHistoryItemDoc struct {
	ID         string  `json:"id" example:"66666666-6666-6666-6666-666666666666"`
	ScheduleID string  `json:"schedule_id" example:"33333333-3333-3333-3333-333333333333"`
	Status     string  `json:"status" example:"confirmed"`
	ChangedAt  string  `json:"changed_at" example:"2026-04-10T11:00:00Z"`
	ChangedBy  string  `json:"changed_by" example:"22222222-2222-2222-2222-222222222222"`
	Notes      *string `json:"notes,omitempty" example:"Confirmado por telefone"`
}

// APIErrorResponseDoc documents unified API error envelope for Swagger.
type APIErrorResponseDoc = middleware.ErrorResponse
