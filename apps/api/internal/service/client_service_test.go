package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestClientService_CreateClient(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewClientService(mock, queries)

	companyID := newDomainUUID(t)
	clientID := newDomainUUID(t)
	personID := newDomainUUID(t)
	birthDate := pgtype.Date{Time: time.Date(1992, 6, 15, 0, 0, 0, 0, time.UTC), Valid: true}
	clientSince := pgtype.Date{Time: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	createdAt := time.Now().UTC().Truncate(time.Second)

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)name: InsertClientPerson`).
		WillReturnRows(pgxmock.NewRows([]string{"id", "kind", "is_active", "has_system_user", "created_at", "updated_at"}).
			AddRow(personID.String(), sqlc.PersonKindClient, true, false, createdAt, nil))

	mock.ExpectQuery(`(?s)name: InsertClientIdentification`).
		WithArgs(
			personID,
			"Maria Silva",
			"Maria",
			sqlc.GenderIdentityWomanCisgender,
			sqlc.MaritalStatusSingle,
			birthDate,
			"12345678901",
		).
		WillReturnRows(pgxmock.NewRows([]string{"id", "person_id", "full_name", "short_name", "gender_identity", "marital_status", "image_url", "birth_date", "cpf", "created_at", "updated_at"}).
			AddRow(newDomainUUID(t).String(), personID.String(), "Maria Silva", "Maria", sqlc.GenderIdentityWomanCisgender, sqlc.MaritalStatusSingle, nil, birthDate, "12345678901", createdAt, nil))

	mock.ExpectQuery(`(?s)name: InsertClientPrimaryContact`).
		WithArgs(
			personID,
			"maria.silva@petcontrol.local",
			pgxmock.AnyArg(),
			"+5511999990001",
			true,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id", "person_id", "email", "phone", "cellphone", "has_whatsapp", "instagram_user", "emergency_contact", "emergency_phone", "is_primary", "created_at", "updated_at"}).
			AddRow(newDomainUUID(t).String(), personID.String(), "maria.silva@petcontrol.local", "+551130000000", "+5511999990001", true, nil, nil, nil, true, createdAt, nil))

	mock.ExpectQuery(`(?s)name: InsertClientRecord`).
		WithArgs(personID, clientSince, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "person_id", "client_since", "recommended_by", "notes", "created_at", "updated_at", "deleted_at"}).
			AddRow(clientID.String(), personID.String(), clientSince, nil, "Cliente recorrente", createdAt, nil, nil))

	mock.ExpectQuery(`(?s)name: CreateCompanyClient`).
		WithArgs(companyID, clientID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "is_active", "joined_at", "left_at"}).
			AddRow(newDomainUUID(t).String(), companyID.String(), clientID.String(), true, createdAt, nil))

	mock.ExpectCommit()
	mock.ExpectQuery(`(?s)name: GetClientByIDAndCompanyID`).
		WithArgs(companyID, clientID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "person_id", "company_id", "full_name", "short_name", "gender_identity",
			"marital_status", "birth_date", "cpf", "email", "phone", "cellphone",
			"has_whatsapp", "client_since", "notes", "is_active", "created_at",
			"updated_at", "joined_at", "left_at",
		}).AddRow(
			clientID.String(),
			personID.String(),
			companyID.String(),
			"Maria Silva",
			"Maria",
			sqlc.GenderIdentityWomanCisgender,
			sqlc.MaritalStatusSingle,
			birthDate,
			"12345678901",
			"maria.silva@petcontrol.local",
			"+551130000000",
			"+5511999990001",
			true,
			clientSince,
			"Cliente recorrente",
			true,
			createdAt,
			nil,
			createdAt,
			nil,
		))

	result, err := serviceUnderTest.CreateClient(context.Background(), CreateClientInput{
		CompanyID:      companyID,
		FullName:       "Maria Silva",
		ShortName:      "Maria",
		GenderIdentity: sqlc.GenderIdentityWomanCisgender,
		MaritalStatus:  sqlc.MaritalStatusSingle,
		BirthDate:      birthDate,
		CPF:            "12345678901",
		Email:          "maria.silva@petcontrol.local",
		Phone:          pgtype.Text{String: "+551130000000", Valid: true},
		Cellphone:      "+5511999990001",
		HasWhatsapp:    true,
		ClientSince:    clientSince,
		Notes:          pgtype.Text{String: "Cliente recorrente", Valid: true},
	})
	require.NoError(t, err)
	require.Equal(t, clientID, result.ID)
	require.Equal(t, "Maria Silva", result.FullName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestClientService_UpdateClient(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewClientService(mock, queries)

	companyID := newDomainUUID(t)
	clientID := newDomainUUID(t)
	personID := newDomainUUID(t)
	createdAt := time.Now().UTC().Truncate(time.Second)
	updatedAt := createdAt.Add(10 * time.Minute)

	mock.ExpectQuery(`(?s)name: GetClientByIDAndCompanyID`).
		WithArgs(companyID, clientID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "person_id", "company_id", "full_name", "short_name", "gender_identity",
			"marital_status", "birth_date", "cpf", "email", "phone", "cellphone",
			"has_whatsapp", "client_since", "notes", "is_active", "created_at",
			"updated_at", "joined_at", "left_at",
		}).AddRow(
			clientID.String(),
			personID.String(),
			companyID.String(),
			"Maria Silva",
			"Maria",
			sqlc.GenderIdentityWomanCisgender,
			sqlc.MaritalStatusSingle,
			time.Date(1992, 6, 15, 0, 0, 0, 0, time.UTC),
			"12345678901",
			"maria.silva@petcontrol.local",
			"+551130000000",
			"+5511999990001",
			true,
			time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			"Cliente recorrente",
			true,
			createdAt,
			nil,
			createdAt,
			nil,
		))

	mock.ExpectExec(`(?s)name: UpdateClientIdentification`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), clientID, companyID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`(?s)name: UpdateClientPrimaryContact`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), clientID, companyID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`(?s)name: UpdateClientRecord`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), clientID, companyID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectQuery(`(?s)name: GetClientByIDAndCompanyID`).
		WithArgs(companyID, clientID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "person_id", "company_id", "full_name", "short_name", "gender_identity",
			"marital_status", "birth_date", "cpf", "email", "phone", "cellphone",
			"has_whatsapp", "client_since", "notes", "is_active", "created_at",
			"updated_at", "joined_at", "left_at",
		}).AddRow(
			clientID.String(),
			personID.String(),
			companyID.String(),
			"Maria Souza",
			"Mari",
			sqlc.GenderIdentityWomanCisgender,
			sqlc.MaritalStatusMarried,
			time.Date(1992, 6, 15, 0, 0, 0, 0, time.UTC),
			"12345678901",
			"maria.souza@petcontrol.local",
			"+551130000000",
			"+5511999990002",
			true,
			time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			"Atualizada",
			true,
			createdAt,
			updatedAt,
			createdAt,
			nil,
		))

	fullName := "Maria Souza"
	shortName := "Mari"
	email := "maria.souza@petcontrol.local"
	cellphone := "+5511999990002"
	notes := "Atualizada"
	hasWhatsapp := true
	maritalStatus := sqlc.MaritalStatusMarried

	result, err := serviceUnderTest.UpdateClient(context.Background(), UpdateClientInput{
		CompanyID:     companyID,
		ClientID:      clientID,
		FullName:      &fullName,
		ShortName:     &shortName,
		MaritalStatus: &maritalStatus,
		Email:         &email,
		Cellphone:     &cellphone,
		HasWhatsapp:   &hasWhatsapp,
		Notes:         &notes,
	})
	require.NoError(t, err)
	require.Equal(t, "Maria Souza", result.FullName)
	require.Equal(t, sqlc.MaritalStatusMarried, result.MaritalStatus)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestClientService_DeactivateClientNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewClientService(mock, queries)

	companyID := newDomainUUID(t)
	clientID := newDomainUUID(t)

	mock.ExpectExec(`(?s)name: DeactivateClient`).
		WithArgs(companyID, clientID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err = serviceUnderTest.DeactivateClient(context.Background(), companyID, clientID)
	require.ErrorIs(t, err, apperror.ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
