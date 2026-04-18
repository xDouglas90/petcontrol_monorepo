package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	if err := ConfigurePool(ctx, poolConfig); err != nil {
		return nil, fmt.Errorf("configure db pool: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("open db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}

func ConfigurePool(ctx context.Context, poolConfig *pgxpool.Config) error {
	previousAfterConnect := poolConfig.AfterConnect
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if previousAfterConnect != nil {
			if err := previousAfterConnect(ctx, conn); err != nil {
				return err
			}
		}

		if err := registerDerivedTypes(ctx, conn); err != nil {
			return fmt.Errorf("register derived postgres types: %w", err)
		}

		return nil
	}

	return nil
}

func registerDerivedTypes(ctx context.Context, conn *pgx.Conn) error {
	for _, typeName := range derivedTypeNames {
		loadedType, err := conn.LoadType(ctx, typeName)
		if err != nil {
			return fmt.Errorf("load type %s: %w", typeName, err)
		}
		conn.TypeMap().RegisterType(loadedType)
	}

	return nil
}

var derivedTypeNames = []string{
	"user_role_type",
	"_user_role_type",
	"user_kind",
	"_user_kind",
	"module_package",
	"_module_package",
	"log_action",
	"_log_action",
	"gender_identity",
	"_gender_identity",
	"marital_status",
	"_marital_status",
	"pet_size",
	"_pet_size",
	"pet_temperament",
	"_pet_temperament",
	"pet_kind",
	"_pet_kind",
	"employee_kind",
	"_employee_kind",
	"pix_key_kind",
	"_pix_key_kind",
	"week_day",
	"_week_day",
	"graduation_level",
	"_graduation_level",
	"person_kind",
	"_person_kind",
	"bank_account_kind",
	"_bank_account_kind",
	"payment_method",
	"_payment_method",
	"schedule_status",
	"_schedule_status",
	"product_kind",
	"_product_kind",
	"login_result",
	"_login_result",
	"logout_reason",
	"_logout_reason",
	"notification_level",
	"_notification_level",
}
