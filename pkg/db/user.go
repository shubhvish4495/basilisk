package db

import (
	"basilisk/pkg/helper"
	"context"
	"errors"
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

func (db *DBStruct) GetUser(ctx context.Context, userID string) (*User, helper.HttpError) {
	return nil, helper.NewHttpError(helper.InternalServerError.Code, errors.New("missing implementation").Error())
}
