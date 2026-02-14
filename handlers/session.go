package handlers

import (
	"net/http"
	"time"

	"github.com/aditya-bijapurkar/aditya-resume-auth/middleware/auth"
	"github.com/aditya-bijapurkar/aditya-resume-auth/utils"
)

func CreateSessionUser(w http.ResponseWriter, r *http.Request) {
	userId, _ := userStore.CreateSessionUser()

	token, err := utils.GenerateToken("session_user", userId, "session_user@adityabijapurkar.in")
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.AuthTokenCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   secureToken,
		SameSite: http.SameSiteStrictMode,
	})
}
