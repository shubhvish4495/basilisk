package handlers

import (
	"net/http"

	"basilisk/pkg/helper"
)

func LoggedInRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)

	logger.Info("logged in request")

	helper.SendSuccessResponse(w, http.StatusOK, map[string]string{"test-field": "field"})

}
