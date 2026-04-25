package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/pagination"
)

type ServiceService struct {
	db      clientTxStarter
	queries *sqlc.Queries
}

type CreateServiceInput struct {
	CompanyID    pgtype.UUID
	TypeName     string
	Title        string
	Description  string
	Notes        pgtype.Text
	Price        pgtype.Numeric
	DiscountRate pgtype.Numeric
	ImageURL     pgtype.Text
	IsActive     bool
	SubServices  []ServiceSubServiceInput
}

type UpdateServiceInput struct {
	CompanyID    pgtype.UUID
	ServiceID    pgtype.UUID
	TypeName     *string
	Title        *string
	Description  *string
	Notes        *string
	Price        *pgtype.Numeric
	DiscountRate *pgtype.Numeric
	ImageURL     *string
	IsActive     *bool
	SubServices  *[]ServiceSubServiceInput
}

type ServiceSubServiceInput struct {
	TypeName     string
	Title        string
	Description  string
	Notes        pgtype.Text
	Price        pgtype.Numeric
	DiscountRate pgtype.Numeric
	ImageURL     pgtype.Text
	IsActive     bool
	AverageTimes []ServiceAverageTimeInput
}

type ServiceAverageTimeInput struct {
	PetSize            sqlc.PetSize
	PetKind            sqlc.PetKind
	PetTemperament     sqlc.PetTemperament
	AverageTimeMinutes int16
}

type ServiceDetail struct {
	Item        sqlc.GetServiceByIDAndCompanyIDRow
	SubServices []ServiceSubServiceDetail
}

type ServiceSubServiceDetail struct {
	Item         sqlc.SubService
	AverageTimes []sqlc.ServicesAverageTime
}

func NewServiceService(db clientTxStarter, queries *sqlc.Queries) *ServiceService {
	return &ServiceService{db: db, queries: queries}
}

func (s *ServiceService) ListServicesByCompanyID(ctx context.Context, companyID pgtype.UUID, p pagination.Params) ([]sqlc.ListServicesByCompanyIDRow, error) {
	return s.queries.ListServicesByCompanyID(ctx, sqlc.ListServicesByCompanyIDParams{
		CompanyID: companyID,
		Search:    p.Search,
		Offset:    int32(p.Offset),
		Limit:     int32(p.Limit),
	})
}

func (s *ServiceService) GetServiceByID(ctx context.Context, companyID pgtype.UUID, serviceID pgtype.UUID) (sqlc.GetServiceByIDAndCompanyIDRow, error) {
	item, err := s.queries.GetServiceByIDAndCompanyID(ctx, sqlc.GetServiceByIDAndCompanyIDParams{
		CompanyID: companyID,
		ID:        serviceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}
	return item, err
}

func (s *ServiceService) GetServiceDetailByID(ctx context.Context, companyID pgtype.UUID, serviceID pgtype.UUID) (ServiceDetail, error) {
	item, err := s.GetServiceByID(ctx, companyID, serviceID)
	if err != nil {
		return ServiceDetail{}, err
	}

	subServices, err := s.queries.ListSubServicesByServiceID(ctx, sqlc.ListSubServicesByServiceIDParams{
		ServiceID: serviceID,
		Limit:     1000,
		Offset:    0,
	})
	if err != nil {
		return ServiceDetail{}, err
	}

	averageTimes, err := s.queries.ListServiceAverageTimesByServiceID(ctx, serviceID)
	if err != nil {
		return ServiceDetail{}, err
	}

	return ServiceDetail{
		Item:        item,
		SubServices: groupServiceAverageTimes(subServices, averageTimes),
	}, nil
}

func (s *ServiceService) CreateService(ctx context.Context, input CreateServiceInput) (ServiceDetail, error) {
	if err := validateServiceSubServices(input.SubServices); err != nil {
		return ServiceDetail{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ServiceDetail{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	txQueries := s.queries.WithTx(tx)
	typeID, err := resolveServiceTypeID(ctx, txQueries, input.TypeName)
	if err != nil {
		return ServiceDetail{}, err
	}

	created, err := txQueries.CreateService(ctx, sqlc.CreateServiceParams{
		TypeID:       typeID,
		Title:        input.Title,
		Description:  input.Description,
		Notes:        input.Notes,
		Price:        input.Price,
		DiscountRate: input.DiscountRate,
		ImageURL:     input.ImageURL,
		IsActive:     input.IsActive,
	})
	if err != nil {
		return ServiceDetail{}, mapServiceDBError(err)
	}

	_, err = txQueries.CreateCompanyService(ctx, sqlc.CreateCompanyServiceParams{
		CompanyID: input.CompanyID,
		ServiceID: created.ID,
	})
	if err != nil {
		return ServiceDetail{}, mapServiceDBError(err)
	}

	if err := replaceServiceSubServices(ctx, txQueries, created.ID, typeID, input.SubServices); err != nil {
		return ServiceDetail{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ServiceDetail{}, err
	}
	committed = true

	return s.GetServiceDetailByID(ctx, input.CompanyID, created.ID)
}

func (s *ServiceService) UpdateService(ctx context.Context, input UpdateServiceInput) (ServiceDetail, error) {
	if _, err := s.GetServiceByID(ctx, input.CompanyID, input.ServiceID); err != nil {
		return ServiceDetail{}, err
	}

	if input.SubServices != nil {
		if err := validateServiceSubServices(*input.SubServices); err != nil {
			return ServiceDetail{}, err
		}
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ServiceDetail{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	txQueries := s.queries.WithTx(tx)
	typeID := pgtype.UUID{}
	if input.TypeName != nil {
		resolvedTypeID, err := resolveServiceTypeID(ctx, txQueries, *input.TypeName)
		if err != nil {
			return ServiceDetail{}, err
		}
		typeID = resolvedTypeID
	}

	updated, err := txQueries.UpdateServiceByIDAndCompanyID(ctx, sqlc.UpdateServiceByIDAndCompanyIDParams{
		TypeID:       typeID,
		Title:        optionalText(input.Title),
		Description:  optionalText(input.Description),
		Notes:        optionalText(input.Notes),
		Price:        optionalNumeric(input.Price),
		DiscountRate: optionalNumeric(input.DiscountRate),
		ImageURL:     optionalText(input.ImageURL),
		IsActive:     optionalBool(input.IsActive),
		ID:           input.ServiceID,
		CompanyID:    input.CompanyID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ServiceDetail{}, apperror.ErrNotFound
		}
		return ServiceDetail{}, mapServiceDBError(err)
	}

	if input.SubServices != nil {
		if _, err := txQueries.DeleteServiceAverageTimesByServiceID(ctx, input.ServiceID); err != nil {
			return ServiceDetail{}, err
		}
		if _, err := txQueries.DeleteSubServicesByServiceID(ctx, input.ServiceID); err != nil {
			return ServiceDetail{}, err
		}
		if err := replaceServiceSubServices(ctx, txQueries, input.ServiceID, updated.TypeID, *input.SubServices); err != nil {
			return ServiceDetail{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return ServiceDetail{}, err
	}
	committed = true

	return s.GetServiceDetailByID(ctx, input.CompanyID, input.ServiceID)
}

func (s *ServiceService) DeactivateService(ctx context.Context, companyID pgtype.UUID, serviceID pgtype.UUID) error {
	_, err := s.queries.DeactivateCompanyService(ctx, sqlc.DeactivateCompanyServiceParams{
		CompanyID: companyID,
		ServiceID: serviceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.ErrNotFound
		}
		return err
	}
	return nil
}

func replaceServiceSubServices(ctx context.Context, queries sqlc.Querier, serviceID pgtype.UUID, defaultTypeID pgtype.UUID, subServices []ServiceSubServiceInput) error {
	for _, item := range subServices {
		typeID := defaultTypeID
		if strings.TrimSpace(item.TypeName) != "" {
			resolvedTypeID, err := resolveServiceTypeID(ctx, queries, item.TypeName)
			if err != nil {
				return err
			}
			typeID = resolvedTypeID
		}

		created, err := queries.InsertSubService(ctx, sqlc.InsertSubServiceParams{
			ServiceID:    serviceID,
			TypeID:       typeID,
			Title:        item.Title,
			Description:  item.Description,
			Notes:        item.Notes,
			Price:        item.Price,
			DiscountRate: item.DiscountRate,
			ImageURL:     item.ImageURL,
			IsActive:     pgtype.Bool{Bool: item.IsActive, Valid: true},
		})
		if err != nil {
			return mapServiceDBError(err)
		}

		for _, averageTime := range item.AverageTimes {
			_, err := queries.InsertServiceAverageTime(ctx, sqlc.InsertServiceAverageTimeParams{
				ServiceID:          serviceID,
				SubServiceID:       created.ID,
				PetSize:            averageTime.PetSize,
				PetKind:            averageTime.PetKind,
				PetTemperament:     averageTime.PetTemperament,
				AverageTimeMinutes: averageTime.AverageTimeMinutes,
			})
			if err != nil {
				return mapServiceDBError(err)
			}
		}
	}
	return nil
}

func validateServiceSubServices(subServices []ServiceSubServiceInput) error {
	if len(subServices) == 0 {
		return apperror.ErrUnprocessableEntity
	}
	for _, item := range subServices {
		if strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.Description) == "" || len(item.AverageTimes) == 0 {
			return apperror.ErrUnprocessableEntity
		}
		for _, averageTime := range item.AverageTimes {
			if averageTime.PetSize == "" || averageTime.PetKind == "" || averageTime.PetTemperament == "" || averageTime.AverageTimeMinutes <= 0 {
				return apperror.ErrUnprocessableEntity
			}
		}
	}
	return nil
}

func groupServiceAverageTimes(subServices []sqlc.SubService, averageTimes []sqlc.ServicesAverageTime) []ServiceSubServiceDetail {
	bySubServiceID := make(map[string][]sqlc.ServicesAverageTime, len(subServices))
	for _, averageTime := range averageTimes {
		if averageTime.SubServiceID.Valid {
			bySubServiceID[averageTime.SubServiceID.String()] = append(bySubServiceID[averageTime.SubServiceID.String()], averageTime)
		}
	}

	details := make([]ServiceSubServiceDetail, 0, len(subServices))
	for _, subService := range subServices {
		details = append(details, ServiceSubServiceDetail{
			Item:         subService,
			AverageTimes: bySubServiceID[subService.ID.String()],
		})
	}
	return details
}

func resolveServiceTypeID(ctx context.Context, queries sqlc.Querier, rawName string) (pgtype.UUID, error) {
	name := strings.TrimSpace(rawName)
	if name == "" {
		return pgtype.UUID{}, apperror.ErrUnprocessableEntity
	}

	serviceType, err := queries.FindServiceTypeByName(ctx, name)
	if err == nil {
		return serviceType.ID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, err
	}

	created, err := queries.CreateServiceType(ctx, sqlc.CreateServiceTypeParams{
		Name:        name,
		Description: pgtype.Text{},
	})
	if err != nil {
		return pgtype.UUID{}, mapServiceDBError(err)
	}

	return created.ID, nil
}

func optionalNumeric(value *pgtype.Numeric) pgtype.Numeric {
	if value == nil {
		return pgtype.Numeric{}
	}
	return *value
}

func mapServiceDBError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return apperror.ErrUnprocessableEntity
		case "23505":
			return apperror.ErrConflict
		}
	}
	return err
}
