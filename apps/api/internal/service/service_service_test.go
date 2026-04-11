package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestServiceService_CreateService(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewServiceService(mock, queries)

	companyID := newDomainUUID(t)
	typeID := newDomainUUID(t)
	serviceID := newDomainUUID(t)
	now := time.Now().UTC().Truncate(time.Second)
	price, err := parseDomainNumeric("89.90")
	require.NoError(t, err)
	discount, err := parseDomainNumeric("0.00")
	require.NoError(t, err)

	mock.ExpectBegin()

	mock.ExpectQuery(`(?s)name: FindServiceTypeByName`).
		WithArgs("Banho").
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at", "deleted_at"}).
			AddRow(typeID.String(), "Banho", nil, now, nil, nil))

	mock.ExpectQuery(`(?s)name: CreateService`).
		WithArgs(typeID, "Banho completo", "Banho com secagem", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), true).
		WillReturnRows(pgxmock.NewRows([]string{"id", "type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(serviceID.String(), typeID.String(), "Banho completo", "Banho com secagem", nil, "89.90", "0.00", nil, true, now, nil, nil))

	mock.ExpectQuery(`(?s)name: CreateCompanyService`).
		WithArgs(companyID, serviceID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "service_id", "is_active", "created_at", "updated_at"}).
			AddRow(newDomainUUID(t).String(), companyID.String(), serviceID.String(), true, now, nil))

	mock.ExpectCommit()

	mock.ExpectQuery(`(?s)name: GetServiceByIDAndCompanyID`).
		WithArgs(companyID, serviceID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "type_id", "type_name", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active"}).
			AddRow(serviceID.String(), typeID.String(), "Banho", "Banho completo", "Banho com secagem", nil, "89.90", "0.00", nil, true))

	item, err := serviceUnderTest.CreateService(context.Background(), CreateServiceInput{
		CompanyID:    companyID,
		TypeName:     "Banho",
		Title:        "Banho completo",
		Description:  "Banho com secagem",
		Price:        price,
		DiscountRate: discount,
		IsActive:     true,
	})
	require.NoError(t, err)
	require.Equal(t, serviceID, item.ID)
	require.Equal(t, "Banho", item.TypeName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func parseDomainNumeric(raw string) (value pgtype.Numeric, err error) {
	err = value.Scan(raw)
	return value, err
}
