package handlers

import (
	"encoding/json"
	"net/http"

	"basilisk/pkg/auth"
	"basilisk/pkg/helper"

	"github.com/google/uuid"
)

type GoogleLoginReqBody struct {
	IDToken string `json:"id_token"`
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)

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

	//generate sessionID here
	sessionID := uuid.New().String()

	token, exp, err := auth.JWTServiceInstance.GenerateToken(ctx, logger, gUserDet.ID, sessionID)
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	refreshTkn, err := auth.JWTServiceInstance.GenerateRefreshToken(ctx, logger, gUserDet.ID, sessionID)
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	helper.SendSuccessResponse(w, http.StatusOK, map[string]any{"token": token, "expiry": exp, "refresh_token": refreshTkn})
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginWithPwd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)

	var lreq LoginReq
	defer func() {
		_ = r.Body.Close()
	}()

	err := json.NewDecoder(r.Body).Decode(&lreq)
	if err != nil {
		logger.Error("error while decoding request body", "error", err)
		helper.SendError(w, helper.BadRequestError)
		return
	}

	//dummy user id being set here
	userID := uuid.New().String()
	sessionID := uuid.New().String()

	token, exp, err := auth.JWTServiceInstance.GenerateToken(ctx, logger, userID, sessionID)
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	refreshTkn, err := auth.JWTServiceInstance.GenerateRefreshToken(ctx, logger, userID, sessionID)
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	helper.SendSuccessResponse(w, http.StatusOK, map[string]any{"token": token, "expiry": exp, "refresh_token": refreshTkn})
}
