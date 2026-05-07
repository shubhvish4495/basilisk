package db

import (
	"basilisk/pkg/helper"
	"context"
	"database/sql"
	"errors"
	"log/slog"
)

type AuthType string

var (
	GoogleAuthType AuthType = "google-auth"
	EmailAuthType  AuthType = "email-auth"
)

type User struct {
	ID         string
	Name       string
	Email      string
	ProfilePic string
	Password   string
	SingUpType AuthType
	Roles      []string
	BaseStruct
}

func (db *DBStruct) GetUser(ctx context.Context, logger *slog.Logger, userID string) (*User, helper.HttpError) {
	//mock transactional method call for building
	_ = db.withTransaction(ctx, logger, func(logger *slog.Logger, tx *sql.Tx) error {
		//dummy function
		return nil
	})

	return nil, helper.NewHttpError(helper.InternalServerError.Code, errors.New("missing implementation").Error())
}
