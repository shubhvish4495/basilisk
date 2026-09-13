package handlers

import (
	"net/http"

	"basilisk/pkg/auth"
	"basilisk/pkg/helper"
)

func LogOutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)
	sessionID := helper.GetSessionIdFromContext(ctx)

	err := auth.JWTServiceInstance.AddSesssionToDenyList(ctx, logger, sessionID)
	if err != nil {
		// we just log do not return err as it is ok to miss on cache addition of denial of session
		logger.Error("error while revoking session access", "error", err)
	}

	helper.SendSuccessResponse(w, http.StatusOK, nil)
}
