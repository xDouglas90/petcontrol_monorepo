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

// CurrentUserResponseDoc documents the current authenticated user payload.
type CurrentUserResponseDoc struct {
	Data CurrentUserDoc `json:"data"`
}

// CurrentUserDoc describes the authenticated user profile returned by `/users/me`.
type CurrentUserDoc struct {
	UserID    string  `json:"user_id" example:"11111111-1111-1111-1111-111111111111"`
	CompanyID string  `json:"company_id" example:"22222222-2222-2222-2222-222222222222"`
	PersonID  string  `json:"person_id" example:"33333333-3333-3333-3333-333333333333"`
	Role      string  `json:"role" example:"admin"`
	Kind      string  `json:"kind" example:"owner"`
	FullName  *string `json:"full_name,omitempty" example:"Maria da Silva"`
	ShortName *string `json:"short_name,omitempty" example:"Maria"`
	ImageURL  *string `json:"image_url,omitempty" example:"https://cdn.example.com/users/maria.png"`
}

// CompanySystemConfigResponseDoc documents the current tenant system config payload.
type CompanySystemConfigResponseDoc struct {
	Data CompanySystemConfigDoc `json:"data"`
}

// CompanySystemConfigDoc describes the current tenant system configuration.
type CompanySystemConfigDoc struct {
	CompanyID             string   `json:"company_id" example:"22222222-2222-2222-2222-222222222222"`
	ScheduleInitTime      string   `json:"schedule_init_time" example:"08:00"`
	SchedulePauseInitTime string   `json:"schedule_pause_init_time" example:"12:00"`
	SchedulePauseEndTime  string   `json:"schedule_pause_end_time" example:"13:00"`
	ScheduleEndTime       string   `json:"schedule_end_time" example:"18:00"`
	MinSchedulesPerDay    int16    `json:"min_schedules_per_day" example:"4"`
	MaxSchedulesPerDay    int16    `json:"max_schedules_per_day" example:"18"`
	ScheduleDays          []string `json:"schedule_days" example:"monday,tuesday,wednesday,thursday,friday,saturday"`
	DynamicCages          bool     `json:"dynamic_cages" example:"false"`
	TotalSmallCages       int16    `json:"total_small_cages" example:"8"`
	TotalMediumCages      int16    `json:"total_medium_cages" example:"6"`
	TotalLargeCages       int16    `json:"total_large_cages" example:"4"`
	TotalGiantCages       int16    `json:"total_giant_cages" example:"2"`
	WhatsappNotifications bool     `json:"whatsapp_notifications" example:"true"`
	WhatsappConversation  bool     `json:"whatsapp_conversation" example:"true"`
	WhatsappBusinessPhone *string  `json:"whatsapp_business_phone,omitempty" example:"+5511999990001"`
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
	UploadKey   string `json:"upload_object_key,omitempty" example:"uploads/pets/image_url/2026/04/uuid-thor.png"`
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
	UploadKey   *string `json:"upload_object_key,omitempty" example:"uploads/pets/image_url/2026/04/uuid-thor.png"`
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

// CompanyItemResponseDoc documents the current company response envelope for Swagger.
type CompanyItemResponseDoc struct {
	Data CompanyDoc `json:"data"`
}

// CompanyDoc describes the public company shape returned by API responses.
type CompanyDoc struct {
	ID             string  `json:"id" example:"11111111-1111-1111-1111-111111111111"`
	Slug           string  `json:"slug" example:"petcontrol-dev"`
	Name           string  `json:"name" example:"PetControl Dev"`
	FantasyName    string  `json:"fantasy_name" example:"PetControl"`
	CNPJ           string  `json:"cnpj" example:"12345678000195"`
	FoundationDate *string `json:"foundation_date,omitempty" example:"2026-01-01"`
	LogoURL        *string `json:"logo_url,omitempty" example:"https://example.com/logo.png"`
	ResponsibleID  string  `json:"responsible_id" example:"22222222-2222-2222-2222-222222222222"`
	ActivePackage  string  `json:"active_package" example:"starter"`
	IsActive       bool    `json:"is_active" example:"true"`
	CreatedAt      string  `json:"created_at" example:"2026-04-10T10:00:00Z"`
	UpdatedAt      *string `json:"updated_at,omitempty" example:"2026-04-10T11:00:00Z"`
	DeletedAt      *string `json:"deleted_at,omitempty" example:"2026-04-11T11:00:00Z"`
}

// ModuleListResponseDoc documents the active modules response envelope for Swagger.
type ModuleListResponseDoc struct {
	Data []ModuleDoc `json:"data"`
}

// ModuleAccessResponseDoc documents a module access check response envelope for Swagger.
type ModuleAccessResponseDoc struct {
	Data ModuleAccessDoc `json:"data"`
}

// ModuleAccessDoc describes a module access check result.
type ModuleAccessDoc struct {
	Allowed bool   `json:"allowed" example:"true"`
	Module  string `json:"module" example:"SCH"`
}

// ModuleDoc describes the public module shape returned by API responses.
type ModuleDoc struct {
	ID          string  `json:"id" example:"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"`
	Code        string  `json:"code" example:"SCH"`
	Name        string  `json:"name" example:"Agendamentos"`
	Description string  `json:"description" example:"Gestao de agenda e servicos"`
	MinPackage  string  `json:"min_package" example:"starter"`
	IsActive    bool    `json:"is_active" example:"true"`
	CreatedAt   string  `json:"created_at" example:"2026-04-10T10:00:00Z"`
	UpdatedAt   *string `json:"updated_at,omitempty" example:"2026-04-10T11:00:00Z"`
	DeletedAt   *string `json:"deleted_at,omitempty" example:"2026-04-11T11:00:00Z"`
}

// PlanItemResponseDoc documents the current plan response envelope for Swagger.
type PlanItemResponseDoc struct {
	Data PlanDoc `json:"data"`
}

// PlanDoc describes the public plan shape returned by API responses.
type PlanDoc struct {
	ID               string  `json:"id" example:"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"`
	PlanTypeID       string  `json:"plan_type_id" example:"cccccccc-cccc-cccc-cccc-cccccccccccc"`
	Name             string  `json:"name" example:"Starter"`
	Description      string  `json:"description" example:"Plano inicial para operacao enxuta"`
	Package          string  `json:"package" example:"starter"`
	Price            string  `json:"price" example:"99.90"`
	BillingCycleDays int     `json:"billing_cycle_days" example:"30"`
	MaxUsers         *int    `json:"max_users,omitempty" example:"5"`
	IsActive         bool    `json:"is_active" example:"true"`
	ImageURL         *string `json:"image_url,omitempty" example:"https://example.com/plans/starter.png"`
	CreatedAt        string  `json:"created_at" example:"2026-04-10T10:00:00Z"`
	UpdatedAt        *string `json:"updated_at,omitempty" example:"2026-04-10T11:00:00Z"`
	DeletedAt        *string `json:"deleted_at,omitempty" example:"2026-04-11T11:00:00Z"`
}

// CompanyUserCreateRequestDoc documents company user creation payload for Swagger.
type CompanyUserCreateRequestDoc struct {
	UserID  string `json:"user_id" example:"22222222-2222-2222-2222-222222222222"`
	IsOwner bool   `json:"is_owner" example:"false"`
}

// CompanyUserListResponseDoc documents the company users response envelope for Swagger.
type CompanyUserListResponseDoc struct {
	Data []CompanyUserDoc `json:"data"`
}

// CompanyUserItemResponseDoc documents a single company user response envelope for Swagger.
type CompanyUserItemResponseDoc struct {
	Data CompanyUserDoc `json:"data"`
}

// CompanyUserDoc describes the public company user shape returned by API responses.
type CompanyUserDoc struct {
	ID        string  `json:"id" example:"dddddddd-dddd-dddd-dddd-dddddddddddd"`
	CompanyID string  `json:"company_id" example:"11111111-1111-1111-1111-111111111111"`
	UserID    string  `json:"user_id" example:"22222222-2222-2222-2222-222222222222"`
	Kind      string  `json:"kind" example:"employee"`
	Role      string  `json:"role" example:"system"`
	IsOwner   bool    `json:"is_owner" example:"false"`
	IsActive  bool    `json:"is_active" example:"true"`
	FullName  *string `json:"full_name,omitempty" example:"System PetControl"`
	ShortName *string `json:"short_name,omitempty" example:"System"`
	ImageURL  *string `json:"image_url,omitempty" example:"https://cdn.example.com/users/system.png"`
	JoinedAt  string  `json:"joined_at" example:"2026-04-10T10:00:00Z"`
	LeftAt    *string `json:"left_at,omitempty" example:"2026-04-11T11:00:00Z"`
}

type AdminSystemChatMessageCreateRequestDoc struct {
	Message string `json:"message" example:"Tudo certo por aí? Precisamos revisar os agendamentos de hoje."`
}

type AdminSystemChatMessageDoc struct {
	ID             string  `json:"id" example:"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"`
	ConversationID string  `json:"conversation_id" example:"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"`
	CompanyID      string  `json:"company_id" example:"11111111-1111-1111-1111-111111111111"`
	SenderUserID   string  `json:"sender_user_id" example:"22222222-2222-2222-2222-222222222222"`
	SenderName     string  `json:"sender_name" example:"System"`
	SenderRole     string  `json:"sender_role" example:"system"`
	SenderImageURL *string `json:"sender_image_url,omitempty" example:"https://cdn.example.com/users/system.png"`
	Body           string  `json:"body" example:"Tudo certo por aí? Precisamos revisar os agendamentos de hoje."`
	CreatedAt      string  `json:"created_at" example:"2026-04-20T09:30:00Z"`
}

type AdminSystemChatMessageListResponseDoc struct {
	Data []AdminSystemChatMessageDoc `json:"data"`
}

type AdminSystemChatMessageItemResponseDoc struct {
	Data AdminSystemChatMessageDoc `json:"data"`
}

// APIErrorResponseDoc documents unified API error envelope for Swagger.
type APIErrorResponseDoc = middleware.ErrorResponse
