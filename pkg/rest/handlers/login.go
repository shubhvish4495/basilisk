package handlers

import (
	"basilisk/pkg/auth"
	"basilisk/pkg/helper"
	"encoding/json"
	"net/http"
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

	token, exp, err := auth.JWTServiceInstance.GenerateToken(gUserDet.ID)
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	refreshTkn, err := auth.JWTServiceInstance.GenerateRefreshToken(gUserDet.ID)
	if err != nil {
		helper.SendError(w, helper.InternalServerError)
		return
	}

	helper.SendSuccessResponse(w, http.StatusOK, map[string]any{"token": token, "expiry": exp, "refresh_token": refreshTkn})
}
