package handlers

import (
	"basilisk/pkg/auth"
	"basilisk/pkg/helper"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RefreshTokenRequest struct {
	Token string `json:"refresh_token"`
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)
	currSessionID := helper.GetSessionIdFromContext(ctx)

	var reqBdy RefreshTokenRequest

	// Extract the request body
	defer func() { _ = r.Body.Close() }()
	err := json.NewDecoder(r.Body).Decode(&reqBdy)
	if err != nil {
		logger.Error("error decoding request body", "error", err)
		helper.SendError(w, helper.NewHttpError(http.StatusBadRequest, "Invalid request body"))
		return
	}

	// Validate the refresh token
	userID, err := auth.JWTServiceInstance.ValidateRefreshToken(ctx, logger, reqBdy.Token)
	if err != nil {
		logger.Error("error validating refresh token", "error", err)
		if err == jwt.ErrTokenExpired {
			helper.SendError(w, helper.NewHttpError(http.StatusUnauthorized, "Refresh token expired"))
		} else {
			helper.SendError(w, helper.NewHttpError(http.StatusUnauthorized, "Invalid refresh token"))
		}
		return
	}

	//generate new session id
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

	// just log the error here
	if err := auth.JWTServiceInstance.AddSesssionToDenyList(ctx, logger, currSessionID); err != nil {
		logger.Error("error while adding session to deny list", "error", err)
	}

	helper.SendSuccessResponse(w, http.StatusOK, map[string]any{"token": token, "expiry": exp, "refresh_token": refreshTkn})
}
