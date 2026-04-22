package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/pagination"
	"github.com/xdouglas90/petcontrol_monorepo/internal/queue"
	"golang.org/x/crypto/bcrypt"
)

type PeopleService struct {
	db        peopleTxStarter
	queries   *sqlc.Queries
	publisher queue.Publisher
}

type PeopleListItem struct {
	ID              pgtype.UUID
	CompanyID       pgtype.UUID
	CompanyPersonID pgtype.UUID
	Kind            sqlc.PersonKind
	FullName        *string
	ShortName       *string
	ImageURL        *string
	CPF             *string
	HasSystemUser   bool
	IsActive        bool
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
}

type PeopleDetail struct {
	PeopleListItem
	Identification    *sqlc.GetPersonIdentificationsRow
	Contact           *sqlc.GetPersonContactsRow
	Address           *AddressDetail
	ClientDetails     *ClientDetails
	EmployeeDetails   *EmployeeDetails
	EmployeeDocuments *sqlc.GetEmployeeDocumentsRow
	EmployeeBenefits  *sqlc.EmployeeBenefit
	LinkedUser        *LinkedUserSummary
	GuardianPets      []GuardianPetSummary
}

type AddressDetail struct {
	Link    sqlc.PeopleAddress
	Address sqlc.GetAddressRow
}

type ClientDetails struct {
	ClientSince pgtype.Date
	Notes       pgtype.Text
}

type EmployeeDetails struct {
	CompanyEmployeeID pgtype.UUID
	Role              *string
	AdmissionDate     pgtype.Date
	ResignationDate   pgtype.Date
	Salary            *string
}

type LinkedUserSummary struct {
	UserID   pgtype.UUID
	Email    string
	Role     sqlc.UserRoleType
	Kind     sqlc.UserKind
	IsActive bool
	IsOwner  bool
	JoinedAt pgtype.Timestamptz
}

type GuardianPetSummary struct {
	PetID     pgtype.UUID
	Name      string
	Kind      sqlc.PetKind
	Size      sqlc.PetSize
	OwnerName string
}

type PersonEmploymentInput struct {
	Role            string
	AdmissionDate   pgtype.Date
	ResignationDate *pgtype.Date
	Salary          pgtype.Numeric
}

type PersonEmployeeDocumentsInput struct {
	RG          string
	IssuingBody string
	IssuingDate pgtype.Date
	CTPS        string
	CTPSSeries  string
	CTPSState   string
	PIS         string
	Graduation  sqlc.GraduationLevel
}

type PersonEmployeeBenefitsInput struct {
	MealTicket            bool
	MealTicketValue       pgtype.Numeric
	TransportVoucher      bool
	TransportVoucherQty   int16
	TransportVoucherValue pgtype.Numeric
	ValidFrom             pgtype.Date
	ValidUntil            *pgtype.Date
}

type PersonAddressInput struct {
	ZipCode    string
	Street     string
	Number     string
	Complement *string
	District   string
	City       string
	State      string
	Country    string
	Label      *string
}

type CreatePersonInput struct {
	CompanyID        pgtype.UUID
	ActorUserID      pgtype.UUID
	ActorRole        sqlc.UserRoleType
	Kind             sqlc.PersonKind
	FullName         string
	ShortName        string
	GenderIdentity   sqlc.GenderIdentity
	MaritalStatus    sqlc.MaritalStatus
	BirthDate        pgtype.Date
	CPF              string
	Email            string
	Phone            *string
	Cellphone        string
	HasWhatsapp      bool
	HasSystemUser    bool
	IsActive         bool
	Address          *PersonAddressInput
	ClientSince      *pgtype.Date
	Notes            *string
	Employment       *PersonEmploymentInput
	EmployeeDocs     *PersonEmployeeDocumentsInput
	EmployeeBenefits *PersonEmployeeBenefitsInput
	PetIDs           []pgtype.UUID
}

type UpdatePersonInput struct {
	CompanyID        pgtype.UUID
	ActorUserID      pgtype.UUID
	ActorRole        sqlc.UserRoleType
	PersonID         pgtype.UUID
	FullName         *string
	ShortName        *string
	GenderIdentity   *sqlc.GenderIdentity
	MaritalStatus    *sqlc.MaritalStatus
	BirthDate        *pgtype.Date
	CPF              *string
	Email            *string
	Phone            *string
	Cellphone        *string
	HasWhatsapp      *bool
	HasSystemUser    *bool
	IsActive         *bool
	Address          *PersonAddressInput
	ClientSince      *pgtype.Date
	Notes            *string
	Employment       *PersonEmploymentInput
	EmployeeDocs     *PersonEmployeeDocumentsInput
	EmployeeBenefits *PersonEmployeeBenefitsInput
	PetIDs           []pgtype.UUID
	HasPetIDs        bool
}

type peopleTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewPeopleService(db peopleTxStarter, queries *sqlc.Queries, publisher queue.Publisher) *PeopleService {
	return &PeopleService{db: db, queries: queries, publisher: publisher}
}

func (s *PeopleService) ListPeopleByCompanyID(ctx context.Context, companyID pgtype.UUID, search string) ([]PeopleListItem, error) {
	itemsByPerson := make(map[string]PeopleListItem)

	companyPeople, err := s.queries.ListCompanyPeople(ctx, sqlc.ListCompanyPeopleParams{
		CompanyID: companyID,
		Offset:    0,
		Limit:     int32(^uint32(0) >> 1),
	})
	if err != nil {
		return nil, err
	}

	for _, item := range companyPeople {
		itemsByPerson[uuidKey(item.PersonID)] = PeopleListItem{
			ID:              item.PersonID,
			CompanyID:       item.CompanyID,
			CompanyPersonID: item.CompanyPersonID,
			Kind:            item.PersonKind,
			FullName:        textValuePointer(item.IdentificationsFullName),
			ShortName:       textValuePointer(item.IdentificationsShortName),
			ImageURL:        textValuePointer(item.IdentificationsImageUrl),
			CPF:             textValuePointer(item.IdentificationsCpf),
			HasSystemUser:   item.PersonHasSystemUser,
			IsActive:        item.PersonIsActive,
			CreatedAt:       item.PersonCreatedAt,
			UpdatedAt:       item.PersonUpdatedAt,
		}
	}

	companyClients, err := s.queries.ListCompanyClients(ctx, sqlc.ListCompanyClientsParams{
		CompanyID: companyID,
		Offset:    0,
		Limit:     int32(^uint32(0) >> 1),
	})
	if err != nil {
		return nil, err
	}

	for _, item := range companyClients {
		client, err := s.queries.GetClientByIDAndCompanyID(ctx, sqlc.GetClientByIDAndCompanyIDParams{
			CompanyID: companyID,
			ID:        item.ClientID,
		})
		if err != nil {
			return nil, err
		}

		key := uuidKey(client.PersonID)
		if _, exists := itemsByPerson[key]; exists {
			continue
		}

		itemsByPerson[key] = PeopleListItem{
			ID:              client.PersonID,
			CompanyID:       companyID,
			CompanyPersonID: item.ID,
			Kind:            sqlc.PersonKindClient,
			FullName:        stringPointer(client.FullName),
			ShortName:       stringPointer(client.ShortName),
			ImageURL:        textValuePointer(item.ClientImageUrl),
			CPF:             stringPointer(client.Cpf),
			HasSystemUser:   false,
			IsActive:        item.IsActive,
			CreatedAt:       item.JoinedAt,
			UpdatedAt:       item.LeftAt,
		}
	}

	companyUsers, err := s.queries.ListCompanyUsersByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	for _, item := range companyUsers {
		profile, err := s.queries.GetUserProfile(ctx, item.UserID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}

		person, err := s.queries.GetPerson(ctx, profile.PersonID)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}

		key := uuidKey(profile.PersonID)
		if _, exists := itemsByPerson[key]; exists {
			continue
		}

		itemsByPerson[key] = PeopleListItem{
			ID:              profile.PersonID,
			CompanyID:       companyID,
			CompanyPersonID: item.ID,
			Kind:            person.Kind,
			FullName:        textValuePointer(person.FullName),
			ShortName:       textValuePointer(person.ShortName),
			ImageURL:        textValuePointer(person.ImageUrl),
			CPF:             textValuePointer(person.Cpf),
			HasSystemUser:   person.HasSystemUser,
			IsActive:        person.IsActive,
			CreatedAt:       person.CreatedAt,
			UpdatedAt:       person.UpdatedAt,
		}
	}

	items := make([]PeopleListItem, 0, len(itemsByPerson))
	for _, item := range itemsByPerson {
		items = append(items, item)
	}

	items = filterPeopleListItemsBySearch(items, search)

	sort.SliceStable(items, func(i, j int) bool {
		leftName := sortablePeopleName(items[i])
		rightName := sortablePeopleName(items[j])
		if leftName != rightName {
			return leftName < rightName
		}

		leftCreatedAt := items[i].CreatedAt.Time
		rightCreatedAt := items[j].CreatedAt.Time
		if !leftCreatedAt.Equal(rightCreatedAt) {
			return leftCreatedAt.Before(rightCreatedAt)
		}

		return uuidKey(items[i].ID) < uuidKey(items[j].ID)
	})

	return items, nil
}

func filterPeopleListItemsBySearch(items []PeopleListItem, search string) []PeopleListItem {
	term := strings.TrimSpace(strings.ToLower(search))
	if term == "" {
		return items
	}

	filtered := make([]PeopleListItem, 0, len(items))
	for _, item := range items {
		haystack := strings.ToLower(strings.Join([]string{
			stringValue(item.FullName),
			stringValue(item.ShortName),
			stringValue(item.CPF),
			string(item.Kind),
		}, " "))
		if strings.Contains(haystack, term) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func sortablePeopleName(item PeopleListItem) string {
	fullName := strings.TrimSpace(stringValue(item.FullName))
	if fullName != "" {
		return strings.ToLower(fullName)
	}

	shortName := strings.TrimSpace(stringValue(item.ShortName))
	if shortName != "" {
		return strings.ToLower(shortName)
	}

	return "~" + uuidKey(item.ID)
}

func (s *PeopleService) GetPersonDetailByID(ctx context.Context, companyID pgtype.UUID, personID pgtype.UUID) (PeopleDetail, error) {
	items, err := s.ListPeopleByCompanyID(ctx, companyID, "")
	if err != nil {
		return PeopleDetail{}, err
	}

	var base *PeopleListItem
	for _, item := range items {
		if uuidKey(item.ID) == uuidKey(personID) {
			copy := item
			base = &copy
			break
		}
	}
	if base == nil {
		return PeopleDetail{}, apperror.ErrNotFound
	}

	detail := PeopleDetail{PeopleListItem: *base}

	identification, err := s.queries.GetPersonIdentifications(ctx, personID)
	if err == nil {
		detail.Identification = &identification
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return PeopleDetail{}, err
	}

	contact, err := s.queries.GetPersonContacts(ctx, personID)
	if err == nil {
		detail.Contact = &contact
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return PeopleDetail{}, err
	}

	addresses, err := s.queries.ListPersonAddresses(ctx, sqlc.ListPersonAddressesParams{
		PersonID: personID,
		Offset:   0,
		Limit:    10,
	})
	if err != nil {
		return PeopleDetail{}, err
	}
	for _, link := range addresses {
		if !link.IsMain {
			continue
		}
		address, addrErr := s.queries.GetAddress(ctx, link.AddressID)
		if addrErr != nil {
			return PeopleDetail{}, addrErr
		}
		detail.Address = &AddressDetail{Link: link, Address: address}
		break
	}

	clients, err := s.queries.ListCompanyClients(ctx, sqlc.ListCompanyClientsParams{
		CompanyID: companyID,
		Offset:    0,
		Limit:     pagination.MaxPageSize,
	})
	if err != nil {
		return PeopleDetail{}, err
	}
	for _, companyClient := range clients {
		client, clientErr := s.queries.GetClientByIDAndCompanyID(ctx, sqlc.GetClientByIDAndCompanyIDParams{
			CompanyID: companyID,
			ID:        companyClient.ClientID,
		})
		if clientErr != nil {
			return PeopleDetail{}, clientErr
		}
		if uuidKey(client.PersonID) != uuidKey(personID) {
			continue
		}
		detail.ClientDetails = &ClientDetails{
			ClientSince: client.ClientSince,
			Notes:       client.Notes,
		}
		break
	}

	employee, err := s.queries.GetCompanyEmployee(ctx, sqlc.GetCompanyEmployeeParams{
		CompanyID: companyID,
		PersonID:  personID,
	})
	if err == nil {
		var role *string
		if employee.Role.Valid {
			role = &employee.Role.String
		}
		detail.EmployeeDetails = &EmployeeDetails{
			CompanyEmployeeID: employee.CompanyEmployeeID,
			Role:              role,
			AdmissionDate:     employee.AdmissionDate,
			ResignationDate:   employee.ResignationDate,
			Salary:            numericToStringPointer(employee.Salary),
		}

		documents, docErr := s.queries.GetEmployeeDocuments(ctx, personID)
		if docErr == nil {
			detail.EmployeeDocuments = &documents
		} else if !errors.Is(docErr, pgx.ErrNoRows) {
			return PeopleDetail{}, docErr
		}

		benefits, benErr := s.queries.GetEmployeeBenefitsByCompanyEmployeeID(ctx, employee.CompanyEmployeeID)
		if benErr == nil {
			detail.EmployeeBenefits = &benefits
		} else if !errors.Is(benErr, pgx.ErrNoRows) {
			return PeopleDetail{}, benErr
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return PeopleDetail{}, err
	}

	companyUsers, err := s.queries.ListCompanyUsersByCompanyID(ctx, companyID)
	if err != nil {
		return PeopleDetail{}, err
	}
	for _, companyUser := range companyUsers {
		profile, profileErr := s.queries.GetUserProfile(ctx, companyUser.UserID)
		if errors.Is(profileErr, pgx.ErrNoRows) {
			continue
		}
		if profileErr != nil {
			return PeopleDetail{}, profileErr
		}
		if uuidKey(profile.PersonID) != uuidKey(personID) {
			continue
		}

		user, userErr := s.queries.GetUserByID(ctx, companyUser.UserID)
		if userErr != nil {
			return PeopleDetail{}, userErr
		}

		detail.LinkedUser = &LinkedUserSummary{
			UserID:   companyUser.UserID,
			Email:    user.Email,
			Role:     user.Role,
			Kind:     companyUser.Kind,
			IsActive: companyUser.IsActive,
			IsOwner:  companyUser.IsOwner,
			JoinedAt: companyUser.CreatedAt,
		}
		break
	}

	guardianPets, err := s.queries.ListGuardianPetsByCompanyID(ctx, sqlc.ListGuardianPetsByCompanyIDParams{
		CompanyID:  companyID,
		GuardianID: personID,
	})
	if err != nil && !isUndefinedTableError(err, "pet_guardians") {
		return PeopleDetail{}, err
	}
	if len(guardianPets) > 0 {
		detail.GuardianPets = make([]GuardianPetSummary, 0, len(guardianPets))
		for _, pet := range guardianPets {
			detail.GuardianPets = append(detail.GuardianPets, GuardianPetSummary{
				PetID:     pet.PetID,
				Name:      pet.Name,
				Kind:      pet.Kind,
				Size:      pet.Size,
				OwnerName: pet.OwnerName,
			})
		}
	}

	return detail, nil
}

func (s *PeopleService) CreatePerson(ctx context.Context, input CreatePersonInput) (PeopleDetail, error) {
	if err := validateHasSystemUserRequest(input.ActorRole, input.Kind, input.HasSystemUser); err != nil {
		return PeopleDetail{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return PeopleDetail{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	txQueries := s.queries.WithTx(tx)
	person, err := txQueries.InsertPerson(ctx, sqlc.InsertPersonParams{
		Kind:          input.Kind,
		IsActive:      pgtype.Bool{Bool: input.IsActive, Valid: true},
		HasSystemUser: pgtype.Bool{Bool: input.HasSystemUser, Valid: true},
	})
	if err != nil {
		return PeopleDetail{}, mapClientDBError(err)
	}

	_, err = txQueries.InsertPersonIdentifications(ctx, sqlc.InsertPersonIdentificationsParams{
		PersonID:       person.ID,
		FullName:       input.FullName,
		ShortName:      input.ShortName,
		GenderIdentity: input.GenderIdentity,
		MaritalStatus:  input.MaritalStatus,
		ImageURL:       pgtype.Text{},
		BirthDate:      input.BirthDate,
		CPF:            input.CPF,
	})
	if err != nil {
		return PeopleDetail{}, mapClientDBError(err)
	}

	_, err = txQueries.InsertPersonContacts(ctx, sqlc.InsertPersonContactsParams{
		PersonID:         person.ID,
		Email:            input.Email,
		Phone:            optionalText(input.Phone),
		Cellphone:        input.Cellphone,
		HasWhatsapp:      input.HasWhatsapp,
		InstagramUser:    pgtype.Text{},
		EmergencyContact: pgtype.Text{},
		EmergencyPhone:   pgtype.Text{},
		IsPrimary:        true,
	})
	if err != nil {
		return PeopleDetail{}, mapClientDBError(err)
	}

	if input.Address != nil {
		if err := insertPersonAddress(ctx, txQueries, person.ID, *input.Address); err != nil {
			return PeopleDetail{}, mapClientDBError(err)
		}
	}

	_, err = txQueries.InsertCompanyPerson(ctx, sqlc.InsertCompanyPersonParams{
		CompanyID: input.CompanyID,
		PersonID:  person.ID,
	})
	if err != nil {
		return PeopleDetail{}, mapClientDBError(err)
	}

	if input.Kind == sqlc.PersonKindClient {
		clientSince := pgtype.Date{}
		if input.ClientSince != nil {
			clientSince = *input.ClientSince
		}
		notes := pgtype.Text{}
		if input.Notes != nil {
			notes = optionalText(input.Notes)
		}

		client, clientErr := txQueries.InsertClientRecord(ctx, sqlc.InsertClientRecordParams{
			PersonID:    person.ID,
			ClientSince: clientSince,
			Notes:       notes,
		})
		if clientErr != nil {
			return PeopleDetail{}, mapClientDBError(clientErr)
		}

		_, clientErr = txQueries.CreateCompanyClient(ctx, sqlc.CreateCompanyClientParams{
			CompanyID: input.CompanyID,
			ClientID:  client.ID,
		})
		if clientErr != nil {
			return PeopleDetail{}, mapClientDBError(clientErr)
		}
	}

	if input.Kind == sqlc.PersonKindEmployee || input.Kind == sqlc.PersonKindOutsourcedEmployee {
		if err := s.createEmployeeData(ctx, txQueries, input.CompanyID, person.ID, input); err != nil {
			return PeopleDetail{}, err
		}
	}

	if input.Kind == sqlc.PersonKindGuardian {
		if err := s.syncGuardianPets(ctx, txQueries, input.CompanyID, person.ID, input.PetIDs); err != nil {
			return PeopleDetail{}, err
		}
	}

	var accessPayload *queue.PersonAccessCredentialsPayload
	if input.HasSystemUser {
		payload, provisionErr := s.provisionSystemUser(ctx, txQueries, provisionSystemUserInput{
			CompanyID:   input.CompanyID,
			ActorUserID: input.ActorUserID,
			ActorRole:   input.ActorRole,
			PersonID:    person.ID,
			Kind:        input.Kind,
			Email:       input.Email,
			FullName:    input.FullName,
		})
		if provisionErr != nil {
			return PeopleDetail{}, provisionErr
		}
		accessPayload = payload
	}

	if err := tx.Commit(ctx); err != nil {
		return PeopleDetail{}, err
	}
	committed = true

	if accessPayload != nil && s.publisher != nil {
		if err := s.publisher.EnqueuePersonAccessCredentials(ctx, *accessPayload); err != nil {
			return PeopleDetail{}, err
		}
	}

	return s.GetPersonDetailByID(ctx, input.CompanyID, person.ID)
}

func (s *PeopleService) UpdatePerson(ctx context.Context, input UpdatePersonInput) (PeopleDetail, error) {
	current, err := s.GetPersonDetailByID(ctx, input.CompanyID, input.PersonID)
	if err != nil {
		return PeopleDetail{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return PeopleDetail{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	txQueries := s.queries.WithTx(tx)

	if err := validateHasSystemUserUpdate(input.ActorRole, current, input.HasSystemUser); err != nil {
		return PeopleDetail{}, err
	}

	_, err = txQueries.UpdatePerson(ctx, sqlc.UpdatePersonParams{
		Kind:          sqlc.NullPersonKind{},
		IsActive:      optionalBool(input.IsActive),
		HasSystemUser: optionalBool(input.HasSystemUser),
		ID:            input.PersonID,
	})
	if err != nil {
		return PeopleDetail{}, mapClientDBError(err)
	}

	_, err = txQueries.UpdatePersonIdentifications(ctx, sqlc.UpdatePersonIdentificationsParams{
		FullName:       optionalText(input.FullName),
		ShortName:      optionalText(input.ShortName),
		GenderIdentity: optionalGenderIdentity(input.GenderIdentity),
		MaritalStatus:  optionalMaritalStatus(input.MaritalStatus),
		ImageURL:       pgtype.Text{},
		BirthDate:      optionalDate(input.BirthDate),
		CPF:            optionalText(input.CPF),
		PersonID:       input.PersonID,
	})
	if err != nil {
		return PeopleDetail{}, mapClientDBError(err)
	}

	_, err = txQueries.UpdatePersonContacts(ctx, sqlc.UpdatePersonContactsParams{
		Email:            optionalText(input.Email),
		Phone:            optionalText(input.Phone),
		Cellphone:        optionalText(input.Cellphone),
		HasWhatsapp:      optionalBool(input.HasWhatsapp),
		InstagramUser:    pgtype.Text{},
		EmergencyContact: pgtype.Text{},
		EmergencyPhone:   pgtype.Text{},
		IsPrimary:        pgtype.Bool{},
		PersonID:         input.PersonID,
	})
	if err != nil {
		return PeopleDetail{}, mapClientDBError(err)
	}

	if input.Address != nil {
		if err := upsertPersonAddress(ctx, txQueries, input.PersonID, current.Address, *input.Address); err != nil {
			return PeopleDetail{}, mapClientDBError(err)
		}
	}

	if current.Kind == sqlc.PersonKindClient {
		clientID, clientErr := findCompanyClientIDByPersonID(ctx, txQueries, input.CompanyID, input.PersonID)
		if clientErr != nil {
			return PeopleDetail{}, clientErr
		}

		if clientID.Valid {
			_, clientErr = txQueries.UpdateClientRecord(ctx, sqlc.UpdateClientRecordParams{
				ClientSince: optionalDate(input.ClientSince),
				Notes:       optionalText(input.Notes),
				ID:          clientID,
				CompanyID:   input.CompanyID,
			})
			if clientErr != nil {
				return PeopleDetail{}, mapClientDBError(clientErr)
			}
		}
	}

	if current.Kind == sqlc.PersonKindEmployee || current.Kind == sqlc.PersonKindOutsourcedEmployee {
		if err := s.updateEmployeeData(ctx, txQueries, input.CompanyID, input.PersonID, current, input); err != nil {
			return PeopleDetail{}, err
		}
	}

	if current.Kind == sqlc.PersonKindGuardian && input.HasPetIDs {
		if err := s.syncGuardianPets(ctx, txQueries, input.CompanyID, input.PersonID, input.PetIDs); err != nil {
			return PeopleDetail{}, err
		}
	}

	var accessPayload *queue.PersonAccessCredentialsPayload
	shouldProvisionAccess := input.HasSystemUser != nil && *input.HasSystemUser && !current.HasSystemUser
	if shouldProvisionAccess {
		contactEmail := current.ContactEmail()
		if input.Email != nil && strings.TrimSpace(*input.Email) != "" {
			contactEmail = strings.TrimSpace(*input.Email)
		}

		payload, provisionErr := s.provisionSystemUser(ctx, txQueries, provisionSystemUserInput{
			CompanyID:   input.CompanyID,
			ActorUserID: input.ActorUserID,
			ActorRole:   input.ActorRole,
			PersonID:    input.PersonID,
			Kind:        current.Kind,
			Email:       contactEmail,
			FullName:    current.DisplayName(input.FullName),
		})
		if provisionErr != nil {
			return PeopleDetail{}, provisionErr
		}
		accessPayload = payload
	}

	if err := tx.Commit(ctx); err != nil {
		return PeopleDetail{}, err
	}
	committed = true

	if accessPayload != nil && s.publisher != nil {
		if err := s.publisher.EnqueuePersonAccessCredentials(ctx, *accessPayload); err != nil {
			return PeopleDetail{}, err
		}
	}

	return s.GetPersonDetailByID(ctx, input.CompanyID, input.PersonID)
}

type provisionSystemUserInput struct {
	CompanyID   pgtype.UUID
	ActorUserID pgtype.UUID
	ActorRole   sqlc.UserRoleType
	PersonID    pgtype.UUID
	Kind        sqlc.PersonKind
	Email       string
	FullName    string
}

func (d PeopleDetail) ContactEmail() string {
	if d.Contact == nil {
		return ""
	}
	return strings.TrimSpace(d.Contact.Email)
}

func (d PeopleDetail) DisplayName(override *string) string {
	if override != nil && strings.TrimSpace(*override) != "" {
		return strings.TrimSpace(*override)
	}
	if d.FullName != nil && strings.TrimSpace(*d.FullName) != "" {
		return strings.TrimSpace(*d.FullName)
	}
	if d.ShortName != nil && strings.TrimSpace(*d.ShortName) != "" {
		return strings.TrimSpace(*d.ShortName)
	}
	return "Pessoa"
}

func validateHasSystemUserRequest(actorRole sqlc.UserRoleType, kind sqlc.PersonKind, hasSystemUser bool) error {
	if !hasSystemUser {
		return nil
	}

	if actorRole == sqlc.UserRoleTypeAdmin && (kind == sqlc.PersonKindEmployee || kind == sqlc.PersonKindOutsourcedEmployee) {
		return nil
	}

	if actorRole == sqlc.UserRoleTypeSystem && kind == sqlc.PersonKindClient {
		return nil
	}

	return apperror.ErrUnprocessableEntity
}

func validateHasSystemUserUpdate(actorRole sqlc.UserRoleType, current PeopleDetail, requested *bool) error {
	if requested == nil {
		return nil
	}

	if current.HasSystemUser && !*requested {
		return apperror.ErrUnprocessableEntity
	}

	return validateHasSystemUserRequest(actorRole, current.Kind, *requested)
}

func resolveProvisionedUserRole(actorRole sqlc.UserRoleType, kind sqlc.PersonKind) (sqlc.UserRoleType, sqlc.UserKind, error) {
	switch {
	case actorRole == sqlc.UserRoleTypeAdmin && kind == sqlc.PersonKindEmployee:
		return sqlc.UserRoleTypeSystem, sqlc.UserKindEmployee, nil
	case actorRole == sqlc.UserRoleTypeAdmin && kind == sqlc.PersonKindOutsourcedEmployee:
		return sqlc.UserRoleTypeSystem, sqlc.UserKindOutsourcedEmployee, nil
	case actorRole == sqlc.UserRoleTypeSystem && kind == sqlc.PersonKindClient:
		return sqlc.UserRoleTypeCommon, sqlc.UserKindClient, nil
	default:
		return "", "", apperror.ErrUnprocessableEntity
	}
}

func ensureUserRoleCreationAllowed(actorRole sqlc.UserRoleType, targetRole sqlc.UserRoleType) error {
	if targetRole != sqlc.UserRoleTypeRoot && targetRole != sqlc.UserRoleTypeInternal {
		return nil
	}

	if actorRole != sqlc.UserRoleTypeRoot && actorRole != sqlc.UserRoleTypeInternal {
		return apperror.ErrForbidden
	}

	return nil
}

func generateTemporaryPassword() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *PeopleService) provisionSystemUser(ctx context.Context, queries *sqlc.Queries, input provisionSystemUserInput) (*queue.PersonAccessCredentialsPayload, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		return nil, apperror.ErrUnprocessableEntity
	}

	if _, err := queries.GetUserByEmail(ctx, email); err == nil {
		return nil, apperror.ErrConflict
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, mapClientDBError(err)
	}

	userRole, userKind, err := resolveProvisionedUserRole(input.ActorRole, input.Kind)
	if err != nil {
		return nil, err
	}
	if err := ensureUserRoleCreationAllowed(input.ActorRole, userRole); err != nil {
		return nil, err
	}

	temporaryPassword, err := generateTemporaryPassword()
	if err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(temporaryPassword)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user, err := queries.InsertUser(ctx, sqlc.InsertUserParams{
		Email:           email,
		EmailVerified:   true,
		EmailVerifiedAt: pgtype.Timestamptz{Time: now, Valid: true},
		Role:            userRole,
		IsActive:        true,
	})
	if err != nil {
		return nil, mapClientDBError(err)
	}

	if err := queries.InsertUserAuth(ctx, sqlc.InsertUserAuthParams{
		UserID:             user.ID,
		PasswordHash:       passwordHash,
		MustChangePassword: pgtype.Bool{Bool: true, Valid: true},
	}); err != nil {
		return nil, mapClientDBError(err)
	}

	if _, err := queries.InsertUserProfile(ctx, sqlc.InsertUserProfileParams{
		UserID:   user.ID,
		PersonID: input.PersonID,
	}); err != nil {
		return nil, mapClientDBError(err)
	}

	if _, err := queries.CreateCompanyUser(ctx, sqlc.CreateCompanyUserParams{
		CompanyID: input.CompanyID,
		UserID:    user.ID,
		Kind:      userKind,
		IsOwner:   false,
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
	}); err != nil {
		return nil, mapClientDBError(err)
	}

	if _, err := queries.InsertUserSettings(ctx, sqlc.InsertUserSettingsParams{
		UserID:               user.ID,
		NotificationsEnabled: pgtype.Bool{Bool: true, Valid: true},
		Theme:                pgtype.Text{String: "light", Valid: true},
		Language:             pgtype.Text{String: "pt-BR", Valid: true},
		Timezone:             pgtype.Text{String: "America/Sao_Paulo", Valid: true},
	}); err != nil {
		return nil, mapClientDBError(err)
	}

	permissionIDs, err := s.defaultPermissionIDsForRole(ctx, queries, input.CompanyID, userRole)
	if err != nil {
		return nil, err
	}
	if len(permissionIDs) > 0 {
		if _, err := queries.BulkInsertUserPermissions(ctx, sqlc.BulkInsertUserPermissionsParams{
			UserID:        user.ID,
			PermissionIDs: permissionIDs,
			GrantedBy:     input.ActorUserID,
		}); err != nil {
			return nil, mapClientDBError(err)
		}
	}

	return &queue.PersonAccessCredentialsPayload{
		Version:           1,
		CompanyID:         uuidKey(input.CompanyID),
		PersonID:          uuidKey(input.PersonID),
		UserID:            uuidKey(user.ID),
		RecipientName:     strings.TrimSpace(input.FullName),
		RecipientEmail:    email,
		TemporaryPassword: temporaryPassword,
		SystemURL:         "",
		Role:              string(userRole),
		OccurredAt:        now,
	}, nil
}

func (s *PeopleService) defaultPermissionIDsForRole(ctx context.Context, queries *sqlc.Queries, companyID pgtype.UUID, role sqlc.UserRoleType) ([]pgtype.UUID, error) {
	modules, err := queries.ListActiveModulesByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	permissionByID := make(map[string]pgtype.UUID)
	for _, module := range modules {
		permissions, permissionErr := queries.ListPermissionsByModule(ctx, sqlc.ListPermissionsByModuleParams{
			ModuleID: module.ID,
			Limit:    1000,
			Offset:   0,
		})
		if permissionErr != nil {
			return nil, permissionErr
		}

		for _, permission := range permissions {
			if !roleInDefaultRoles(role, permission.DefaultRoles) {
				continue
			}
			key := uuidKey(permission.ID)
			if key == "" {
				continue
			}
			permissionByID[key] = permission.ID
		}
	}

	permissionIDs := make([]pgtype.UUID, 0, len(permissionByID))
	for _, permissionID := range permissionByID {
		permissionIDs = append(permissionIDs, permissionID)
	}
	sort.SliceStable(permissionIDs, func(i, j int) bool {
		return strings.Compare(uuidKey(permissionIDs[i]), uuidKey(permissionIDs[j])) < 0
	})

	return permissionIDs, nil
}

func (s *PeopleService) syncGuardianPets(ctx context.Context, queries *sqlc.Queries, companyID pgtype.UUID, guardianID pgtype.UUID, petIDs []pgtype.UUID) error {
	if _, err := queries.DeletePetGuardiansByGuardianID(ctx, guardianID); err != nil {
		return mapClientDBError(err)
	}

	seen := make(map[string]struct{}, len(petIDs))
	for _, petID := range petIDs {
		key := uuidKey(petID)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		if _, err := queries.GetPetByIDAndCompanyID(ctx, sqlc.GetPetByIDAndCompanyIDParams{
			CompanyID: companyID,
			ID:        petID,
		}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.ErrUnprocessableEntity
			}
			return mapClientDBError(err)
		}

		if _, err := queries.UpsertPetGuardian(ctx, sqlc.UpsertPetGuardianParams{
			PetID:      petID,
			GuardianID: guardianID,
		}); err != nil {
			return mapClientDBError(err)
		}
	}

	return nil
}

func uuidKey(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	parsed, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return ""
	}
	return parsed.String()
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	copy := value
	return &copy
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func numericToStringPointer(value pgtype.Numeric) *string {
	if !value.Valid {
		return nil
	}
	formatted, err := value.Float64Value()
	if err != nil || !formatted.Valid {
		return nil
	}
	text := formatFloat(formatted.Float64)
	return &text
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func isUndefinedTableError(err error, tableName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "42P01" {
		return false
	}
	if strings.TrimSpace(tableName) == "" {
		return true
	}
	return strings.Contains(strings.ToLower(pgErr.Message), strings.ToLower(tableName))
}

func (s *PeopleService) createEmployeeData(ctx context.Context, queries *sqlc.Queries, companyID pgtype.UUID, personID pgtype.UUID, input CreatePersonInput) error {
	if input.Employment == nil {
		return apperror.ErrUnprocessableEntity
	}

	employee, err := queries.InsertCompanyEmployee(ctx, sqlc.InsertCompanyEmployeeParams{
		CompanyID: companyID,
		PersonID:  personID,
	})
	if err != nil {
		return mapClientDBError(err)
	}

	_, err = queries.InsertEployment(ctx, sqlc.InsertEploymentParams{
		CompanyEmployeeID: employee.ID,
		Role:              input.Employment.Role,
		AdmissionDate:     input.Employment.AdmissionDate,
		ResignationDate:   optionalDate(input.Employment.ResignationDate),
		Salary:            input.Employment.Salary,
	})
	if err != nil {
		return mapClientDBError(err)
	}

	if shouldPersistEmployeeDocuments(input.EmployeeDocs) {
		_, err = queries.InsertEmployeeDocuments(ctx, sqlc.InsertEmployeeDocumentsParams{
			PersonID:              personID,
			RG:                    input.EmployeeDocs.RG,
			IssuingBody:           input.EmployeeDocs.IssuingBody,
			IssuingDate:           input.EmployeeDocs.IssuingDate,
			CTPS:                  input.EmployeeDocs.CTPS,
			CTPSSeries:            input.EmployeeDocs.CTPSSeries,
			CTPSState:             input.EmployeeDocs.CTPSState,
			PIS:                   input.EmployeeDocs.PIS,
			VoterRegistration:     pgtype.Text{},
			VoteZone:              pgtype.Text{},
			VoteSection:           pgtype.Text{},
			MilitaryCertificate:   pgtype.Text{},
			MilitarySeries:        pgtype.Text{},
			MilitaryCategory:      pgtype.Text{},
			HasSpecialNeeds:       pgtype.Bool{Bool: false, Valid: true},
			HasChildren:           pgtype.Bool{Bool: false, Valid: true},
			ChildrenQty:           pgtype.Int2{Int16: 0, Valid: true},
			HasChildrenUnder18:    pgtype.Bool{Bool: false, Valid: true},
			HasFamilySpecialNeeds: pgtype.Bool{Bool: false, Valid: true},
			Graduation:            input.EmployeeDocs.Graduation,
			HasCNH:                pgtype.Bool{Bool: false, Valid: true},
			CNHType:               pgtype.Text{},
			CNHNumber:             pgtype.Text{},
			CNHExpirationDate:     pgtype.Date{},
		})
		if err != nil {
			return mapClientDBError(err)
		}
	}

	if input.EmployeeBenefits != nil {
		_, err = queries.InsertEmployeeBenefits(ctx, sqlc.InsertEmployeeBenefitsParams{
			CompanyEmployeeID:     employee.ID,
			MealTicket:            pgtype.Bool{Bool: input.EmployeeBenefits.MealTicket, Valid: true},
			MealTicketValue:       input.EmployeeBenefits.MealTicketValue,
			TransportVoucher:      pgtype.Bool{Bool: input.EmployeeBenefits.TransportVoucher, Valid: true},
			TransportVoucherQty:   pgtype.Int2{Int16: input.EmployeeBenefits.TransportVoucherQty, Valid: true},
			TransportVoucherValue: input.EmployeeBenefits.TransportVoucherValue,
			ValidFrom:             input.EmployeeBenefits.ValidFrom,
			ValidUntil:            optionalDate(input.EmployeeBenefits.ValidUntil),
		})
		if err != nil {
			return mapClientDBError(err)
		}
	}

	return nil
}

func (s *PeopleService) updateEmployeeData(ctx context.Context, queries *sqlc.Queries, companyID pgtype.UUID, personID pgtype.UUID, current PeopleDetail, input UpdatePersonInput) error {
	employee, err := queries.GetCompanyEmployee(ctx, sqlc.GetCompanyEmployeeParams{
		CompanyID: companyID,
		PersonID:  personID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		inserted, insertErr := queries.InsertCompanyEmployee(ctx, sqlc.InsertCompanyEmployeeParams{
			CompanyID: companyID,
			PersonID:  personID,
		})
		if insertErr != nil {
			return mapClientDBError(insertErr)
		}

		employee.CompanyEmployeeID = inserted.ID
		employee.CompanyID = inserted.CompanyID
		employee.PersonID = inserted.PersonID
		employee.EmploymentID = pgtype.UUID{}
	} else if err != nil {
		return mapClientDBError(err)
	}

	if input.Employment != nil {
		if employee.EmploymentID.Valid {
			_, err = queries.UpdateEmployment(ctx, sqlc.UpdateEmploymentParams{
				Role:            toText(input.Employment.Role),
				AdmissionDate:   input.Employment.AdmissionDate,
				ResignationDate: optionalDate(input.Employment.ResignationDate),
				Salary:          input.Employment.Salary,
				ID:              employee.EmploymentID,
			})
		} else {
			_, err = queries.InsertEployment(ctx, sqlc.InsertEploymentParams{
				CompanyEmployeeID: employee.CompanyEmployeeID,
				Role:              input.Employment.Role,
				AdmissionDate:     input.Employment.AdmissionDate,
				ResignationDate:   optionalDate(input.Employment.ResignationDate),
				Salary:            input.Employment.Salary,
			})
		}
		if err != nil {
			return mapClientDBError(err)
		}
	}

	if input.EmployeeDocs != nil {
		if current.EmployeeDocuments != nil {
			_, err = queries.UpdateEmployeeDocuments(ctx, sqlc.UpdateEmployeeDocumentsParams{
				RG:                    toText(input.EmployeeDocs.RG),
				IssuingBody:           toText(input.EmployeeDocs.IssuingBody),
				IssuingDate:           input.EmployeeDocs.IssuingDate,
				CTPS:                  toText(input.EmployeeDocs.CTPS),
				CTPSSeries:            toText(input.EmployeeDocs.CTPSSeries),
				CTPSState:             toText(input.EmployeeDocs.CTPSState),
				PIS:                   toText(input.EmployeeDocs.PIS),
				VoterRegistration:     pgtype.Text{},
				VoteZone:              pgtype.Text{},
				VoteSection:           pgtype.Text{},
				MilitaryCertificate:   pgtype.Text{},
				MilitarySeries:        pgtype.Text{},
				MilitaryCategory:      pgtype.Text{},
				HasSpecialNeeds:       pgtype.Bool{},
				HasChildren:           pgtype.Bool{},
				ChildrenQty:           pgtype.Int2{},
				HasChildrenUnder18:    pgtype.Bool{},
				HasFamilySpecialNeeds: pgtype.Bool{},
				Graduation:            sqlc.NullGraduationLevel{GraduationLevel: input.EmployeeDocs.Graduation, Valid: true},
				HasCNH:                pgtype.Bool{},
				CNHType:               pgtype.Text{},
				CNHNumber:             pgtype.Text{},
				CNHExpirationDate:     pgtype.Date{},
				PersonID:              personID,
			})
		} else if shouldPersistEmployeeDocuments(input.EmployeeDocs) {
			_, err = queries.InsertEmployeeDocuments(ctx, sqlc.InsertEmployeeDocumentsParams{
				PersonID:              personID,
				RG:                    input.EmployeeDocs.RG,
				IssuingBody:           input.EmployeeDocs.IssuingBody,
				IssuingDate:           input.EmployeeDocs.IssuingDate,
				CTPS:                  input.EmployeeDocs.CTPS,
				CTPSSeries:            input.EmployeeDocs.CTPSSeries,
				CTPSState:             input.EmployeeDocs.CTPSState,
				PIS:                   input.EmployeeDocs.PIS,
				VoterRegistration:     pgtype.Text{},
				VoteZone:              pgtype.Text{},
				VoteSection:           pgtype.Text{},
				MilitaryCertificate:   pgtype.Text{},
				MilitarySeries:        pgtype.Text{},
				MilitaryCategory:      pgtype.Text{},
				HasSpecialNeeds:       pgtype.Bool{Bool: false, Valid: true},
				HasChildren:           pgtype.Bool{Bool: false, Valid: true},
				ChildrenQty:           pgtype.Int2{Int16: 0, Valid: true},
				HasChildrenUnder18:    pgtype.Bool{Bool: false, Valid: true},
				HasFamilySpecialNeeds: pgtype.Bool{Bool: false, Valid: true},
				Graduation:            input.EmployeeDocs.Graduation,
				HasCNH:                pgtype.Bool{Bool: false, Valid: true},
				CNHType:               pgtype.Text{},
				CNHNumber:             pgtype.Text{},
				CNHExpirationDate:     pgtype.Date{},
			})
		} else {
			err = nil
		}
		if err != nil {
			return mapClientDBError(err)
		}
	}

	if input.EmployeeBenefits != nil {
		if current.EmployeeBenefits != nil {
			_, err = queries.UpdateEmployeeBenefits(ctx, sqlc.UpdateEmployeeBenefitsParams{
				MealTicket:            pgtype.Bool{Bool: input.EmployeeBenefits.MealTicket, Valid: true},
				MealTicketValue:       input.EmployeeBenefits.MealTicketValue,
				TransportVoucher:      pgtype.Bool{Bool: input.EmployeeBenefits.TransportVoucher, Valid: true},
				TransportVoucherQty:   pgtype.Int2{Int16: input.EmployeeBenefits.TransportVoucherQty, Valid: true},
				TransportVoucherValue: input.EmployeeBenefits.TransportVoucherValue,
				ValidFrom:             input.EmployeeBenefits.ValidFrom,
				ValidUntil:            optionalDate(input.EmployeeBenefits.ValidUntil),
				CompanyEmployeeID:     employee.CompanyEmployeeID,
			})
		} else {
			_, err = queries.InsertEmployeeBenefits(ctx, sqlc.InsertEmployeeBenefitsParams{
				CompanyEmployeeID:     employee.CompanyEmployeeID,
				MealTicket:            pgtype.Bool{Bool: input.EmployeeBenefits.MealTicket, Valid: true},
				MealTicketValue:       input.EmployeeBenefits.MealTicketValue,
				TransportVoucher:      pgtype.Bool{Bool: input.EmployeeBenefits.TransportVoucher, Valid: true},
				TransportVoucherQty:   pgtype.Int2{Int16: input.EmployeeBenefits.TransportVoucherQty, Valid: true},
				TransportVoucherValue: input.EmployeeBenefits.TransportVoucherValue,
				ValidFrom:             input.EmployeeBenefits.ValidFrom,
				ValidUntil:            optionalDate(input.EmployeeBenefits.ValidUntil),
			})
		}
		if err != nil {
			return mapClientDBError(err)
		}
	}

	return nil
}

func insertPersonAddress(ctx context.Context, queries *sqlc.Queries, personID pgtype.UUID, input PersonAddressInput) error {
	address, err := queries.CreateAddress(ctx, sqlc.CreateAddressParams{
		ZipCode:    input.ZipCode,
		Street:     input.Street,
		Number:     input.Number,
		Complement: optionalText(input.Complement),
		District:   input.District,
		City:       input.City,
		State:      input.State,
		Country:    input.Country,
	})
	if err != nil {
		return err
	}

	_, err = queries.InsertPersonAddress(ctx, sqlc.InsertPersonAddressParams{
		PersonID:  personID,
		AddressID: address.ID,
		IsMain:    true,
		Label:     optionalText(input.Label),
	})
	return err
}

func upsertPersonAddress(ctx context.Context, queries *sqlc.Queries, personID pgtype.UUID, current *AddressDetail, input PersonAddressInput) error {
	if current == nil {
		return insertPersonAddress(ctx, queries, personID, input)
	}

	_, err := queries.UpdateAddress(ctx, sqlc.UpdateAddressParams{
		ZipCode:    toText(input.ZipCode),
		Street:     toText(input.Street),
		Number:     toText(input.Number),
		Complement: optionalText(input.Complement),
		District:   toText(input.District),
		City:       toText(input.City),
		State:      toText(input.State),
		Country:    toText(input.Country),
		ID:         current.Link.AddressID,
	})
	if err != nil {
		return err
	}

	_, err = queries.UpdatePersonAddress(ctx, sqlc.UpdatePersonAddressParams{
		IsMain: pgtype.Bool{Bool: true, Valid: true},
		Label:  optionalText(input.Label),
		ID:     current.Link.ID,
	})
	return err
}

func findCompanyClientIDByPersonID(ctx context.Context, queries sqlc.Querier, companyID pgtype.UUID, personID pgtype.UUID) (pgtype.UUID, error) {
	clients, err := queries.ListCompanyClients(ctx, sqlc.ListCompanyClientsParams{
		CompanyID: companyID,
		Offset:    0,
		Limit:     pagination.MaxPageSize,
	})
	if err != nil {
		return pgtype.UUID{}, err
	}

	for _, companyClient := range clients {
		client, clientErr := queries.GetClientByIDAndCompanyID(ctx, sqlc.GetClientByIDAndCompanyIDParams{
			CompanyID: companyID,
			ID:        companyClient.ClientID,
		})
		if clientErr != nil {
			return pgtype.UUID{}, clientErr
		}
		if uuidKey(client.PersonID) == uuidKey(personID) {
			return companyClient.ClientID, nil
		}
	}

	return pgtype.UUID{}, nil
}

func shouldPersistEmployeeDocuments(input *PersonEmployeeDocumentsInput) bool {
	if input == nil {
		return false
	}

	return input.IssuingDate.Valid &&
		strings.TrimSpace(input.RG) != "" &&
		strings.TrimSpace(input.IssuingBody) != "" &&
		strings.TrimSpace(input.CTPS) != "" &&
		strings.TrimSpace(input.CTPSSeries) != "" &&
		strings.TrimSpace(input.CTPSState) != "" &&
		strings.TrimSpace(input.PIS) != ""
}
