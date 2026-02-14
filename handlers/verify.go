package handlers

import (
	"net/http"
	"os"

	"github.com/aditya-bijapurkar/aditya-resume-auth/models"
	"github.com/aditya-bijapurkar/aditya-resume-auth/utils"
)

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		utils.RespondError(w, http.StatusBadRequest, "Verification token is required")
		return
	}

	claims, err := utils.ValidateVerificationToken(token)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid or expired verification token")
		return
	}

	if err := userStore.UpdateUserVerification(claims.Email); err != nil {
		if err == models.ErrUserNotFound {
			utils.RespondError(w, http.StatusNotFound, "User not found")
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	frontendURL := os.Getenv("REDIRECT_URL")
	http.Redirect(w, r, frontendURL, http.StatusFound)
}
