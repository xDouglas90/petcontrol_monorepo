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

// ClientCreateRequestDoc documents client creation payload for Swagger.
type ClientCreateRequestDoc struct {
	FullName       string `json:"full_name" example:"Maria Silva"`
	ShortName      string `json:"short_name" example:"Maria"`
	GenderIdentity string `json:"gender_identity" example:"woman_cisgender"`
	MaritalStatus  string `json:"marital_status" example:"single"`
	BirthDate      string `json:"birth_date" example:"1992-06-15"`
	CPF            string `json:"cpf" example:"12345678901"`
	Email          string `json:"email" example:"maria.silva@petcontrol.local"`
	Phone          string `json:"phone,omitempty" example:"+551130000000"`
	Cellphone      string `json:"cellphone" example:"+5511999990001"`
	HasWhatsapp    bool   `json:"has_whatsapp" example:"true"`
	ClientSince    string `json:"client_since,omitempty" example:"2026-04-01"`
	Notes          string `json:"notes,omitempty" example:"Cliente recorrente"`
}

// ClientUpdateRequestDoc documents client update payload for Swagger.
type ClientUpdateRequestDoc struct {
	FullName       *string `json:"full_name,omitempty" example:"Maria Souza"`
	ShortName      *string `json:"short_name,omitempty" example:"Mari"`
	GenderIdentity *string `json:"gender_identity,omitempty" example:"woman_cisgender"`
	MaritalStatus  *string `json:"marital_status,omitempty" example:"married"`
	BirthDate      *string `json:"birth_date,omitempty" example:"1992-06-15"`
	CPF            *string `json:"cpf,omitempty" example:"12345678901"`
	Email          *string `json:"email,omitempty" example:"maria.souza@petcontrol.local"`
	Phone          *string `json:"phone,omitempty" example:"+551130000000"`
	Cellphone      *string `json:"cellphone,omitempty" example:"+5511999990002"`
	HasWhatsapp    *bool   `json:"has_whatsapp,omitempty" example:"true"`
	ClientSince    *string `json:"client_since,omitempty" example:"2026-04-01"`
	Notes          *string `json:"notes,omitempty" example:"Preferência por contato via WhatsApp"`
}

// ClientListResponseDoc documents the list clients response envelope for Swagger.
type ClientListResponseDoc struct {
	Data []ClientDoc `json:"data"`
}

// ClientItemResponseDoc documents a single client response envelope for Swagger.
type ClientItemResponseDoc struct {
	Data ClientDoc `json:"data"`
}

// ClientDoc describes the public client shape returned by API responses.
type ClientDoc struct {
	ID             string  `json:"id" example:"44444444-4444-4444-4444-444444444444"`
	PersonID       string  `json:"person_id" example:"55555555-5555-5555-5555-555555555555"`
	CompanyID      string  `json:"company_id" example:"11111111-1111-1111-1111-111111111111"`
	FullName       string  `json:"full_name" example:"Maria Silva"`
	ShortName      string  `json:"short_name" example:"Maria"`
	GenderIdentity string  `json:"gender_identity" example:"woman_cisgender"`
	MaritalStatus  string  `json:"marital_status" example:"single"`
	BirthDate      string  `json:"birth_date" example:"1992-06-15"`
	CPF            string  `json:"cpf" example:"12345678901"`
	Email          string  `json:"email" example:"maria.silva@petcontrol.local"`
	Phone          *string `json:"phone,omitempty" example:"+551130000000"`
	Cellphone      string  `json:"cellphone" example:"+5511999990001"`
	HasWhatsapp    bool    `json:"has_whatsapp" example:"true"`
	ClientSince    *string `json:"client_since,omitempty" example:"2026-04-01"`
	Notes          *string `json:"notes,omitempty" example:"Cliente recorrente"`
	IsActive       bool    `json:"is_active" example:"true"`
	CreatedAt      string  `json:"created_at" example:"2026-04-10T10:00:00Z"`
	UpdatedAt      *string `json:"updated_at,omitempty" example:"2026-04-10T11:00:00Z"`
	JoinedAt       string  `json:"joined_at" example:"2026-04-10T10:00:00Z"`
	LeftAt         *string `json:"left_at,omitempty" example:"2026-04-11T11:00:00Z"`
}

// PetCreateRequestDoc documents pet creation payload for Swagger.
type PetCreateRequestDoc struct {
	OwnerID     string `json:"owner_id" example:"44444444-4444-4444-4444-444444444444"`
	Name        string `json:"name" example:"Thor"`
	Size        string `json:"size" example:"medium"`
	Kind        string `json:"kind" example:"dog"`
	Temperament string `json:"temperament" example:"playful"`
	ImageURL    string `json:"image_url,omitempty" example:"https://example.com/thor.png"`
	BirthDate   string `json:"birth_date,omitempty" example:"2021-08-20"`
	Notes       string `json:"notes,omitempty" example:"Gosta de brincar"`
}

// PetUpdateRequestDoc documents pet update payload for Swagger.
type PetUpdateRequestDoc struct {
	OwnerID     *string `json:"owner_id,omitempty" example:"44444444-4444-4444-4444-444444444444"`
	Name        *string `json:"name,omitempty" example:"Thorzinho"`
	Size        *string `json:"size,omitempty" example:"large"`
	Kind        *string `json:"kind,omitempty" example:"dog"`
	Temperament *string `json:"temperament,omitempty" example:"loving"`
	ImageURL    *string `json:"image_url,omitempty" example:"https://example.com/thor.png"`
	BirthDate   *string `json:"birth_date,omitempty" example:"2021-08-20"`
	Notes       *string `json:"notes,omitempty" example:"Atualizado após consulta"`
}

// PetListResponseDoc documents the list pets response envelope for Swagger.
type PetListResponseDoc struct {
	Data []PetDoc `json:"data"`
}

// PetItemResponseDoc documents a single pet response envelope for Swagger.
type PetItemResponseDoc struct {
	Data PetDoc `json:"data"`
}

// PetDoc describes the public pet shape returned by API responses.
type PetDoc struct {
	ID          string  `json:"id" example:"77777777-7777-7777-7777-777777777777"`
	OwnerID     string  `json:"owner_id" example:"44444444-4444-4444-4444-444444444444"`
	CompanyID   string  `json:"company_id" example:"11111111-1111-1111-1111-111111111111"`
	OwnerName   string  `json:"owner_name" example:"Maria Silva"`
	Name        string  `json:"name" example:"Thor"`
	Size        string  `json:"size" example:"medium"`
	Kind        string  `json:"kind" example:"dog"`
	Temperament string  `json:"temperament" example:"playful"`
	ImageURL    *string `json:"image_url,omitempty" example:"https://example.com/thor.png"`
	BirthDate   *string `json:"birth_date,omitempty" example:"2021-08-20"`
	IsActive    bool    `json:"is_active" example:"true"`
	Notes       *string `json:"notes,omitempty" example:"Gosta de brincar"`
	CreatedAt   string  `json:"created_at" example:"2026-04-10T10:00:00Z"`
	UpdatedAt   *string `json:"updated_at,omitempty" example:"2026-04-10T11:00:00Z"`
	DeletedAt   *string `json:"deleted_at,omitempty" example:"2026-04-11T11:00:00Z"`
}

// ServiceCreateRequestDoc documents service creation payload for Swagger.
type ServiceCreateRequestDoc struct {
	TypeName     string `json:"type_name" example:"Banho"`
	Title        string `json:"title" example:"Banho completo"`
	Description  string `json:"description" example:"Banho com secagem e perfume"`
	Notes        string `json:"notes,omitempty" example:"Inclui perfume hipoalergênico"`
	Price        string `json:"price" example:"89.90"`
	DiscountRate string `json:"discount_rate,omitempty" example:"0.00"`
	ImageURL     string `json:"image_url,omitempty" example:"https://example.com/services/bath.png"`
	IsActive     *bool  `json:"is_active,omitempty" example:"true"`
}

// ServiceUpdateRequestDoc documents service update payload for Swagger.
type ServiceUpdateRequestDoc struct {
	TypeName     *string `json:"type_name,omitempty" example:"Tosa"`
	Title        *string `json:"title,omitempty" example:"Tosa higiênica"`
	Description  *string `json:"description,omitempty" example:"Tosa leve para manutenção"`
	Notes        *string `json:"notes,omitempty" example:"Adicionar avaliação do pelo"`
	Price        *string `json:"price,omitempty" example:"59.90"`
	DiscountRate *string `json:"discount_rate,omitempty" example:"10.00"`
	ImageURL     *string `json:"image_url,omitempty" example:"https://example.com/services/tosa.png"`
	IsActive     *bool   `json:"is_active,omitempty" example:"true"`
}

// ServiceListResponseDoc documents the list services response envelope for Swagger.
type ServiceListResponseDoc struct {
	Data []ServiceDoc `json:"data"`
}

// ServiceItemResponseDoc documents a single service response envelope for Swagger.
type ServiceItemResponseDoc struct {
	Data ServiceDoc `json:"data"`
}

// ServiceDoc describes the public service shape returned by API responses.
type ServiceDoc struct {
	ID           string  `json:"id" example:"88888888-8888-8888-8888-888888888888"`
	TypeID       string  `json:"type_id" example:"99999999-9999-9999-9999-999999999999"`
	TypeName     string  `json:"type_name" example:"Banho"`
	Title        string  `json:"title" example:"Banho completo"`
	Description  string  `json:"description" example:"Banho com secagem e perfume"`
	Notes        *string `json:"notes,omitempty" example:"Inclui perfume hipoalergênico"`
	Price        string  `json:"price" example:"89.90"`
	DiscountRate string  `json:"discount_rate" example:"0.00"`
	ImageURL     *string `json:"image_url,omitempty" example:"https://example.com/services/bath.png"`
	IsActive     bool    `json:"is_active" example:"true"`
}

// ScheduleCreateRequestDoc documents schedule creation payload for Swagger.
type ScheduleCreateRequestDoc struct {
	ClientID     string   `json:"client_id" example:"11111111-1111-1111-1111-111111111111"`
	PetID        string   `json:"pet_id" example:"22222222-2222-2222-2222-222222222222"`
	ServiceIDs   []string `json:"service_ids,omitempty"`
	ScheduledAt  string   `json:"scheduled_at" example:"2026-04-10T15:00:00Z"`
	EstimatedEnd *string  `json:"estimated_end,omitempty" example:"2026-04-10T16:00:00Z"`
	Notes        string   `json:"notes" example:"Banho e tosa"`
	Status       string   `json:"status,omitempty" example:"waiting"`
	StatusNotes  string   `json:"status_notes,omitempty" example:"Aguardando confirmação"`
}

// ScheduleUpdateRequestDoc documents schedule update payload for Swagger.
type ScheduleUpdateRequestDoc struct {
	ClientID     *string  `json:"client_id,omitempty" example:"11111111-1111-1111-1111-111111111111"`
	PetID        *string  `json:"pet_id,omitempty" example:"22222222-2222-2222-2222-222222222222"`
	ServiceIDs   []string `json:"service_ids,omitempty"`
	ScheduledAt  *string  `json:"scheduled_at,omitempty" example:"2026-04-10T15:00:00Z"`
	EstimatedEnd *string  `json:"estimated_end,omitempty" example:"2026-04-10T16:00:00Z"`
	Notes        *string  `json:"notes,omitempty" example:"Confirmado com o cliente"`
	Status       *string  `json:"status,omitempty" example:"confirmed"`
	StatusNotes  *string  `json:"status_notes,omitempty" example:"Confirmado por telefone"`
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
	ID            string   `json:"id" example:"33333333-3333-3333-3333-333333333333"`
	CompanyID     string   `json:"company_id" example:"11111111-1111-1111-1111-111111111111"`
	ClientID      string   `json:"client_id" example:"44444444-4444-4444-4444-444444444444"`
	PetID         string   `json:"pet_id" example:"55555555-5555-5555-5555-555555555555"`
	ClientName    string   `json:"client_name" example:"Maria Silva"`
	PetName       string   `json:"pet_name" example:"Thor"`
	ServiceIDs    []string `json:"service_ids,omitempty"`
	ServiceTitles []string `json:"service_titles,omitempty"`
	ScheduledAt   string   `json:"scheduled_at" example:"2026-04-10T15:00:00Z"`
	EstimatedEnd  *string  `json:"estimated_end,omitempty" example:"2026-04-10T16:00:00Z"`
	Notes         *string  `json:"notes,omitempty" example:"Banho e tosa"`
	CurrentStatus string   `json:"current_status" example:"confirmed"`
	CreatedAt     string   `json:"created_at" example:"2026-04-10T10:00:00Z"`
	UpdatedAt     *string  `json:"updated_at,omitempty" example:"2026-04-10T11:00:00Z"`
	DeletedAt     *string  `json:"deleted_at,omitempty" example:"2026-04-11T11:00:00Z"`
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
