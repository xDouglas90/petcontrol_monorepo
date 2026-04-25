package service

import (
	"context"
	"errors"
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
	subServiceID := newDomainUUID(t)
	averageTimeID := newDomainUUID(t)
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

	mock.ExpectQuery(`(?s)name: InsertSubService`).
		WithArgs(serviceID, typeID, "Banho médio", "Banho para pets médios", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "service_id", "type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(subServiceID.String(), serviceID.String(), typeID.String(), "Banho médio", "Banho para pets médios", nil, "89.90", "0.00", nil, true, now, nil, nil))

	mock.ExpectQuery(`(?s)name: InsertServiceAverageTime`).
		WithArgs(serviceID, subServiceID, sqlc.PetSizeMedium, sqlc.PetKindDog, sqlc.PetTemperamentPlayful, int16(60)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "service_id", "sub_service_id", "pet_size", "pet_kind", "pet_temperament", "average_time_minutes", "created_at", "updated_at"}).
			AddRow(averageTimeID.String(), serviceID.String(), subServiceID.String(), "medium", "dog", "playful", int16(60), now, nil))

	mock.ExpectCommit()

	mock.ExpectQuery(`(?s)name: GetServiceByIDAndCompanyID`).
		WithArgs(companyID, serviceID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "type_id", "type_name", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "sub_services_count", "average_times_count"}).
			AddRow(serviceID.String(), typeID.String(), "Banho", "Banho completo", "Banho com secagem", nil, "89.90", "0.00", nil, true, int64(1), int64(1)))

	mock.ExpectQuery(`(?s)name: ListSubServicesByServiceID`).
		WithArgs(serviceID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "service_id", "type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(subServiceID.String(), serviceID.String(), typeID.String(), "Banho médio", "Banho para pets médios", nil, "89.90", "0.00", nil, true, now, nil, nil))

	mock.ExpectQuery(`(?s)name: ListServiceAverageTimesByServiceID`).
		WithArgs(serviceID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "service_id", "sub_service_id", "pet_size", "pet_kind", "pet_temperament", "average_time_minutes", "created_at", "updated_at"}).
			AddRow(averageTimeID.String(), serviceID.String(), subServiceID.String(), "medium", "dog", "playful", int16(60), now, nil))

	item, err := serviceUnderTest.CreateService(context.Background(), CreateServiceInput{
		CompanyID:    companyID,
		TypeName:     "Banho",
		Title:        "Banho completo",
		Description:  "Banho com secagem",
		Price:        price,
		DiscountRate: discount,
		IsActive:     true,
		SubServices: []ServiceSubServiceInput{
			{
				Title:        "Banho médio",
				Description:  "Banho para pets médios",
				Price:        price,
				DiscountRate: discount,
				IsActive:     true,
				AverageTimes: []ServiceAverageTimeInput{
					{
						PetSize:            sqlc.PetSizeMedium,
						PetKind:            sqlc.PetKindDog,
						PetTemperament:     sqlc.PetTemperamentPlayful,
						AverageTimeMinutes: 60,
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, serviceID, item.Item.ID)
	require.Equal(t, "Banho", item.Item.TypeName)
	require.Len(t, item.SubServices, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceService_CreateServiceRollsBackWhenAverageTimeFails(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewServiceService(mock, queries)

	companyID := newDomainUUID(t)
	typeID := newDomainUUID(t)
	serviceID := newDomainUUID(t)
	subServiceID := newDomainUUID(t)
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

	mock.ExpectQuery(`(?s)name: InsertSubService`).
		WithArgs(serviceID, typeID, "Banho médio", "Banho para pets médios", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "service_id", "type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(subServiceID.String(), serviceID.String(), typeID.String(), "Banho médio", "Banho para pets médios", nil, "89.90", "0.00", nil, true, now, nil, nil))

	mock.ExpectQuery(`(?s)name: InsertServiceAverageTime`).
		WithArgs(serviceID, subServiceID, sqlc.PetSizeMedium, sqlc.PetKindDog, sqlc.PetTemperamentPlayful, int16(60)).
		WillReturnError(errors.New("average time insert failed"))

	mock.ExpectRollback()

	_, err = serviceUnderTest.CreateService(context.Background(), CreateServiceInput{
		CompanyID:    companyID,
		TypeName:     "Banho",
		Title:        "Banho completo",
		Description:  "Banho com secagem",
		Price:        price,
		DiscountRate: discount,
		IsActive:     true,
		SubServices: []ServiceSubServiceInput{
			{
				Title:        "Banho médio",
				Description:  "Banho para pets médios",
				Price:        price,
				DiscountRate: discount,
				IsActive:     true,
				AverageTimes: []ServiceAverageTimeInput{
					{
						PetSize:            sqlc.PetSizeMedium,
						PetKind:            sqlc.PetKindDog,
						PetTemperament:     sqlc.PetTemperamentPlayful,
						AverageTimeMinutes: 60,
					},
				},
			},
		},
	})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func parseDomainNumeric(raw string) (value pgtype.Numeric, err error) {
	err = value.Scan(raw)
	return value, err
}
