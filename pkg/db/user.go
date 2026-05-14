package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"basilisk/pkg/helper"

	"github.com/google/uuid"
)

var (
	GoogleAuthType AuthType = "google-auth"
	EmailAuthType  AuthType = "email-auth"
)

type AuthType string

type User struct {
	Name         string
	Email        string
	ProfilePic   string
	Password     string
	MobileNumber string
	SignUpType   AuthType
	IsVerified   *bool
	BaseStruct
}

func getCreateUserQuery(user User) (string, []any) {
	args := make([]any, 0)
	sql := "INSERT INTO USERS (name, email, profile_pic, password, mobile_number, sign_up_type, is_verified) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"

	args = append(args, user.Name)
	args = append(args, user.Email)
	args = append(args, user.ProfilePic)
	args = append(args, user.Password)
	args = append(args, user.MobileNumber)
	args = append(args, user.SignUpType)
	args = append(args, *user.IsVerified)

	return sql, args
}

func getGetUserByEmailIDQuery(emailID string) (string, []any) {
	args := make([]any, 0)
	sql := `SELECT id, name, email,
			profile_pic, password, mobile_number,
			sign_up_type, is_verified, is_deleted,
			created_at, updated_at
			FROM USERS where email = $1
			AND is_deleted = false`

	args = append(args, emailID)
	return sql, args
}

func getGetUserByUserIDQuery(userID string) (string, []any) {
	args := make([]any, 0)
	sql := `SELECT id, name, email,
			profile_pic, password, mobile_number,
			sign_up_type, is_verified, is_deleted,
			created_at, updated_at
			FROM USERS where id = $1
			AND is_deleted = false`

	args = append(args, userID)
	return sql, args
}

func (db *DBStruct) CreateUser(ctx context.Context, logger *slog.Logger, user User) (uuid.UUID, helper.HttpError) {
	logger.Info("creating user in database")

	q, args := getCreateUserQuery(user)
	id, httpErr := db.insertReturningID(ctx, logger, q, args...)
	if httpErr != nil {
		logger.Error("user creation failed")
		return uuid.Nil, httpErr
	}

	logger.Info("user successfully created in database")
	return id, nil
}

// GetUserByEmail fetches the user from database, be sure to not return this exact struct
// in response as it contains user's password as well
func (db *DBStruct) GetUserByEmail(ctx context.Context, logger *slog.Logger, emailID string) (*User, helper.HttpError) {
	logger.Info("getting user from database with emailID", "emailID", emailID)

	q, args := getGetUserByEmailIDQuery(emailID)
	row, httpErr := db.getSingleRow(ctx, logger, q, args...)
	if httpErr != nil {
		logger.Error("get user operation failed")
		return nil, httpErr
	}

	var user User
	var profilePic, mobileNum, password sql.NullString
	scanErr := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&profilePic,
		&password,
		&mobileNum,
		&user.SignUpType,
		&user.IsVerified,
		&user.IsDeleted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if scanErr != nil {
		logger.Error("error while scanning user row", "error", scanErr.Error())
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil, helper.NotFoundError
		}
		return nil, helper.InternalServerError
	}

	if password.Valid {
		user.Password = password.String
	}

	if mobileNum.Valid {
		user.MobileNumber = mobileNum.String
	}

	if profilePic.Valid {
		user.ProfilePic = profilePic.String
	}

	logger.Info("successfully got back user from database")
	return &user, nil
}

// GetUserByID fetches the user from database, be sure to not return this exact struct
// in response as it contains user's password as well
func (db *DBStruct) GetUserByID(ctx context.Context, logger *slog.Logger, userID string) (*User, helper.HttpError) {
	logger.Info("getting user from database with userID", "userID", userID)

	q, args := getGetUserByUserIDQuery(userID)
	row, httpErr := db.getSingleRow(ctx, logger, q, args...)
	if httpErr != nil {
		logger.Error("get user operation failed")
		return nil, httpErr
	}

	var user User
	var profilePic, mobileNum, password sql.NullString
	scanErr := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&profilePic,
		&password,
		&mobileNum,
		&user.SignUpType,
		&user.IsVerified,
		&user.IsDeleted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil, helper.NotFoundError
		}
		logger.Error("error while scanning user row", "error", scanErr.Error())
		return nil, helper.InternalServerError
	}

	if password.Valid {
		user.Password = password.String
	}

	if mobileNum.Valid {
		user.MobileNumber = mobileNum.String
	}

	if profilePic.Valid {
		user.ProfilePic = profilePic.String
	}

	logger.Info("successfully got back user from database")
	return &user, nil
}
