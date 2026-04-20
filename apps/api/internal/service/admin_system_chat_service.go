package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type AdminSystemChatService struct {
	queries sqlc.Querier
}

type AdminSystemChatMessage struct {
	ID             pgtype.UUID
	ConversationID pgtype.UUID
	CompanyID      pgtype.UUID
	SenderUserID   pgtype.UUID
	SenderName     string
	SenderRole     sqlc.UserRoleType
	SenderImageURL *string
	Body           string
	CreatedAt      pgtype.Timestamptz
}

func NewAdminSystemChatService(queries sqlc.Querier) *AdminSystemChatService {
	return &AdminSystemChatService{queries: queries}
}

func (s *AdminSystemChatService) ListMessages(
	ctx context.Context,
	companyID pgtype.UUID,
	currentUserID pgtype.UUID,
	contactUserID pgtype.UUID,
) ([]AdminSystemChatMessage, error) {
	adminUserID, systemUserID, err := s.resolveParticipants(ctx, companyID, currentUserID, contactUserID)
	if err != nil {
		return nil, err
	}

	_, err = s.queries.UpsertAdminSystemConversation(ctx, sqlc.UpsertAdminSystemConversationParams{
		CompanyID:    companyID,
		AdminUserID:  adminUserID,
		SystemUserID: systemUserID,
	})
	if err != nil {
		return nil, err
	}

	rows, err := s.queries.ListAdminSystemMessages(ctx, sqlc.ListAdminSystemMessagesParams{
		CompanyID:    companyID,
		AdminUserID:  adminUserID,
		SystemUserID: systemUserID,
	})
	if err != nil {
		return nil, err
	}

	messages := make([]AdminSystemChatMessage, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, AdminSystemChatMessage{
			ID:             row.ID,
			ConversationID: row.ConversationID,
			CompanyID:      row.CompanyID,
			SenderUserID:   row.SenderUserID,
			SenderName:     row.SenderName,
			SenderRole:     row.SenderRole,
			SenderImageURL: textValuePointer(row.SenderImageUrl),
			Body:           row.Body,
			CreatedAt:      row.CreatedAt,
		})
	}

	return messages, nil
}

func (s *AdminSystemChatService) SendMessage(
	ctx context.Context,
	companyID pgtype.UUID,
	currentUserID pgtype.UUID,
	contactUserID pgtype.UUID,
	message string,
) (AdminSystemChatMessage, error) {
	body := strings.TrimSpace(message)
	if body == "" {
		return AdminSystemChatMessage{}, apperror.ErrBadRequest
	}

	adminUserID, systemUserID, err := s.resolveParticipants(ctx, companyID, currentUserID, contactUserID)
	if err != nil {
		return AdminSystemChatMessage{}, err
	}

	conversation, err := s.queries.UpsertAdminSystemConversation(ctx, sqlc.UpsertAdminSystemConversationParams{
		CompanyID:    companyID,
		AdminUserID:  adminUserID,
		SystemUserID: systemUserID,
	})
	if err != nil {
		return AdminSystemChatMessage{}, err
	}

	inserted, err := s.queries.InsertAdminSystemMessage(ctx, sqlc.InsertAdminSystemMessageParams{
		ConversationID: conversation.ID,
		CompanyID:      companyID,
		SenderUserID:   currentUserID,
		Body:           body,
	})
	if err != nil {
		return AdminSystemChatMessage{}, err
	}

	sender, err := s.queries.GetUserByID(ctx, currentUserID)
	if err != nil {
		return AdminSystemChatMessage{}, err
	}

	profile, err := s.GetSenderProfile(ctx, currentUserID)
	if err != nil {
		return AdminSystemChatMessage{}, err
	}

	senderName := profile.ShortName
	if senderName == nil {
		senderName = profile.FullName
	}
	name := sender.Email
	if senderName != nil && strings.TrimSpace(*senderName) != "" {
		name = strings.TrimSpace(*senderName)
	}

	return AdminSystemChatMessage{
		ID:             inserted.ID,
		ConversationID: inserted.ConversationID,
		CompanyID:      inserted.CompanyID,
		SenderUserID:   inserted.SenderUserID,
		SenderName:     name,
		SenderRole:     sender.Role,
		SenderImageURL: profile.ImageURL,
		Body:           inserted.Body,
		CreatedAt:      inserted.CreatedAt,
	}, nil
}

func (s *AdminSystemChatService) GetSenderProfile(ctx context.Context, userID pgtype.UUID) (CurrentUserProfile, error) {
	userService := NewUserService(s.queries)
	return userService.GetCurrentUserProfile(ctx, userID)
}

func (s *AdminSystemChatService) resolveParticipants(
	ctx context.Context,
	companyID pgtype.UUID,
	currentUserID pgtype.UUID,
	contactUserID pgtype.UUID,
) (pgtype.UUID, pgtype.UUID, error) {
	if !currentUserID.Valid || !contactUserID.Valid {
		return pgtype.UUID{}, pgtype.UUID{}, apperror.ErrBadRequest
	}

	if currentUserID == contactUserID {
		return pgtype.UUID{}, pgtype.UUID{}, apperror.ErrBadRequest
	}

	currentMembership, err := s.queries.GetCompanyUser(ctx, sqlc.GetCompanyUserParams{
		CompanyID: companyID,
		UserID:    currentUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, pgtype.UUID{}, apperror.ErrForbidden
	}
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, err
	}
	if !currentMembership.IsActive {
		return pgtype.UUID{}, pgtype.UUID{}, apperror.ErrForbidden
	}

	contactMembership, err := s.queries.GetCompanyUser(ctx, sqlc.GetCompanyUserParams{
		CompanyID: companyID,
		UserID:    contactUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, pgtype.UUID{}, apperror.ErrForbidden
	}
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, err
	}
	if !contactMembership.IsActive {
		return pgtype.UUID{}, pgtype.UUID{}, apperror.ErrForbidden
	}

	currentUser, err := s.queries.GetUserByID(ctx, currentUserID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, err
	}

	contactUser, err := s.queries.GetUserByID(ctx, contactUserID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, err
	}

	switch currentUser.Role {
	case sqlc.UserRoleTypeAdmin:
		if contactUser.Role != sqlc.UserRoleTypeSystem {
			return pgtype.UUID{}, pgtype.UUID{}, apperror.ErrForbidden
		}
		return currentUserID, contactUserID, nil
	case sqlc.UserRoleTypeSystem:
		if contactUser.Role != sqlc.UserRoleTypeAdmin {
			return pgtype.UUID{}, pgtype.UUID{}, apperror.ErrForbidden
		}
		return contactUserID, currentUserID, nil
	default:
		return pgtype.UUID{}, pgtype.UUID{}, apperror.ErrForbidden
	}
}
