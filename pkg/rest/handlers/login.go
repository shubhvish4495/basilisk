package handlers

import (
	"encoding/json"
	"net/http"

	"basilisk/pkg/auth"
	"basilisk/pkg/db"
	"basilisk/pkg/helper"
)

type GoogleLoginReqBody struct {
	IDToken string `json:"id_token"`
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)

	defer func() {
		_ = r.Body.Close()
	}()

	var reqBdy GoogleLoginReqBody
	err := json.NewDecoder(r.Body).Decode(&reqBdy)
	if err != nil {
		logger.Error("error while decoding request body", "error", err)
		helper.SendError(w, helper.InternalServerError)
		return
	}

	gUserDet, err := auth.GoogleAuthInstance.ValidateIDToken(ctx, logger, reqBdy.IDToken)
	if err != nil {
		helper.SendError(w, helper.UnauthorizedError)
		return
	}

	dbUser, httpErr := db.GetInstance().GetUserByEmail(ctx, logger, gUserDet.Email)
	if httpErr != nil {
		helper.SendError(w, httpErr)
		return
	}

	// check if user is not deleted
	if dbUser.IsDeleted {
		logger.Error("user is soft deleted, can't proceed with login")
		helper.SendError(w, helper.UnauthorizedError)
		return
	}

	// check if user is of the same google login sign up type
	if dbUser.SignUpType != db.GoogleAuthType {
		logger.Error("user is not signed up using Google Auth, can't proceed to login")
		helper.SendError(w, helper.UnauthorizedError)
		return
	}

	token, exp, err := auth.JWTServiceInstance.GenerateToken(dbUser.ID.String())
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	refreshTkn, err := auth.JWTServiceInstance.GenerateRefreshToken(dbUser.ID.String())
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	helper.SendSuccessResponse(w, http.StatusOK, map[string]any{"token": token, "expiry": exp, "refresh_token": refreshTkn})
}

type UserLoginUsingPwdRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserLoginUsingPassword handles users login using password
func UserLoginUsingPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)

	var rBdy UserLoginUsingPwdRequest
	defer func() {
		_ = r.Body.Close()
	}()

	if err := json.NewDecoder(r.Body).Decode(&rBdy); err != nil {
		logger.Error("error while decoding request body", "error", err)
		helper.SendError(w, helper.InternalServerError)
		return
	}

	dbUser, httpErr := db.GetInstance().GetUserByEmail(ctx, logger, rBdy.Email)
	if httpErr != nil {
		helper.SendError(w, httpErr)
		return
	}

	// hash passed password and match it with db
	// if not correct we send back Unauthorized Error
	hshPwd := helper.HashString(rBdy.Password)
	if hshPwd != dbUser.Password {
		logger.Error("password does not match with what we have in db")
		helper.SendError(w, helper.NewHttpError(http.StatusUnauthorized, "password incorrect"))
		return
	}

	token, exp, err := auth.JWTServiceInstance.GenerateToken(dbUser.ID.String())
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	refreshTkn, err := auth.JWTServiceInstance.GenerateRefreshToken(dbUser.ID.String())
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	helper.SendSuccessResponse(w, http.StatusOK, map[string]any{"token": token, "expiry": exp, "refresh_token": refreshTkn})
}
