package db

import (
	"basilisk/pkg/helper"
	"context"
	"errors"
)

func (db *DBStruct) GetUser(ctx context.Context, userID string) (*User, helper.HttpError) {
	return nil, helper.NewHttpError(helper.InternalServerError.Code, errors.New("missing implementation").Error())
}
