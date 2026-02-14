package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aditya-bijapurkar/aditya-resume-auth/models"
	"github.com/aditya-bijapurkar/aditya-resume-auth/utils"
)

type CreateUsersRequest struct {
	Users []models.AnonymousUser `json:"users"`
}

func CreateAnonymousUsers(w http.ResponseWriter, r *http.Request) {
	var req CreateUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := userStore.CreateAnonymousUsers(req.Users); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create anonymous users")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Anonymous users created successfully"})
}
