package handlers

import (
	"net/http"

	"basilisk/pkg/db"
	"basilisk/pkg/helper"
)

type UserGetResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	MobileNumber string `json:"mobile_number"`
	IsVerified   bool   `json:"is_verified"`
	ProfilePic   string `json:"profile_pic"`
}

// SelfUserGetHandler returns users details for which the request has been made
// userID is extracted from JWT token and is present in context as injected by
// middleware
func SelfUserGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := helper.GetLogger(ctx)

	//extract userId from context
	userID := helper.GetUserIDFromContext(ctx)
	user, err := db.GetInstance().GetUserByID(ctx, logger, userID)
	if err != nil {
		helper.SendError(w, err)
		return
	}

	// convert from db user to response user
	resp := UserGetResponse{
		ID:           user.ID.String(),
		Name:         user.Name,
		Email:        user.Email,
		MobileNumber: user.MobileNumber,
		IsVerified:   *user.IsVerified,
		ProfilePic:   user.ProfilePic,
	}

	helper.SendSuccessResponse(w, http.StatusOK, resp)
}
