package handlers

import (
	"encoding/json"
	"net/http"

	"basilisk/pkg/auth"
	"basilisk/pkg/db"
	"basilisk/pkg/helper"
)

// CreateUserRequest is the request body for email+password based sign up
type CreateUserRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	MobileNumber string `json:"mobile_number"`
}

// CreateGoogleSignUpUserRequest is the request body for Google OAuth based sign up
type CreateGoogleSignUpUserRequest struct {
	IDToken string `json:"id_token"`
}

// CreateUserFromSignUpFormHandler handles when user sign up with us using
// email and password
func CreateUserFromSignUpFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)

	defer func() {
		_ = r.Body.Close()
	}()

	var loginReq CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		logger.Error("error while decoding request body", "error", err)
		helper.SendError(w, helper.InternalServerError)
		return
	}

	// create db User object for request
	dbReq := db.User{
		Name:         loginReq.Name,
		Email:        loginReq.Email,
		Password:     helper.HashString(loginReq.Password),
		MobileNumber: loginReq.MobileNumber,
		SignUpType:   db.EmailAuthType,
		IsVerified:   &[]bool{false}[0],
	}

	// make db call to create user
	userID, httpErr := db.GetInstance().CreateUser(ctx, logger, dbReq)
	if httpErr != nil {
		if httpErr.StatusCode() == helper.ConflictError.Code {
			helper.SendError(w, helper.NewHttpError(helper.ConflictError.Code, "User with same email already exists in database. Try logging in with password or Google Auth"))
		} else {
			helper.SendError(w, httpErr)
		}
		return
	}

	helper.SendSuccessResponse(w, http.StatusCreated, map[string]string{"user_id": userID.String()})
}

// CreateUserFromGoogleSignUpHandler creates user in our database when user clicks on
// Sign up with Google option on our UI
func CreateUserFromGoogleSignUpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)

	var rBdy CreateGoogleSignUpUserRequest
	defer func() {
		_ = r.Body.Close()
	}()

	if err := json.NewDecoder(r.Body).Decode(&rBdy); err != nil {
		logger.Error("error while decoding request body", "error", err)
		helper.SendError(w, helper.InternalServerError)
		return
	}

	authUser, err := auth.GoogleAuthInstance.ValidateIDToken(ctx, logger, rBdy.IDToken)
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	uID, httpErr := db.GetInstance().CreateUser(ctx, logger, db.User{
		Name:       authUser.Name,
		Email:      authUser.Email,
		SignUpType: db.GoogleAuthType,
		ProfilePic: authUser.Picture,
		IsVerified: &[]bool{true}[0],
	})

	if httpErr != nil {
		if httpErr.StatusCode() == helper.ConflictError.Code {
			helper.SendError(w, helper.NewHttpError(helper.ConflictError.Code, "User with same email already exists in database. Try logging in with password or Google Auth"))
		} else {
			helper.SendError(w, httpErr)
		}
		return
	}

	helper.SendSuccessResponse(w, http.StatusCreated, map[string]string{"user_id": uID.String()})
}
