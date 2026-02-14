package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aditya-bijapurkar/aditya-resume-auth/middleware/auth"
	"github.com/aditya-bijapurkar/aditya-resume-auth/models"
	"github.com/aditya-bijapurkar/aditya-resume-auth/utils"
	"golang.org/x/crypto/bcrypt"
)

type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type SignupOperation string

const (
	NewUser             SignupOperation = "new_user"
	SyncSessionUser     SignupOperation = "sync_user"
	UpdateAnonymousUser SignupOperation = "update_anonymous_user"
)

func getSignupOperation(r *http.Request, user *models.User) (operation SignupOperation, userId string) {
	operation = NewUser
	userId = ""

	// Check if an anonymous user was created with the same email while scheduling meetings
	anonymousUserId, _ := userStore.GetAnonymousUserByEmail(user.Email)
	if anonymousUserId != "" {
		operation = UpdateAnonymousUser
		userId = anonymousUserId
	}

	// Check if the user has to be synced with the session user
	sessionUserId, _ := r.Context().Value(auth.UserIDKey).(string)
	if sessionUserId != "" && userStore.AnonymousUserIdExists(sessionUserId) {
		operation = SyncSessionUser
		userId = sessionUserId

		// Do nothing if anonymous user is not deleted
		if err := userStore.DeleteUserById(anonymousUserId); err != nil {
			log.Printf("Failed to delete anonymous user %s: %v", anonymousUserId, err)
		}
	}

	return operation, userId
}

func Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.RespondError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	if len(req.Password) < 6 {
		utils.RespondError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	operation, userId := getSignupOperation(r, user)

	switch operation {
	case NewUser:
		userId, err = userStore.CreateUser(user)
	case SyncSessionUser, UpdateAnonymousUser:
		err = userStore.UpdateUserSignupDetails(user, userId)
	default:
		log.Fatalf("Invalid signup operation: %s", operation)
	}

	if err != nil {
		if err == models.ErrUserExists {
			utils.RespondError(w, http.StatusConflict, "User already exists")
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	verificationToken, err := utils.GenerateVerificationToken(user.Email)
	if err == nil && emailService != nil {
		emailService.SendVerificationEmailAsync(user.Email, verificationToken)
	}

	utils.RespondJSON(w, http.StatusCreated, SignupResponse{
		Status:  http.StatusCreated,
		Message: "User created successfully",
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.RespondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := userStore.GetUserByEmail(req.Email)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Email or password is incorrect, please try again")
		return
	}
	if !user.IsVerified {
		utils.RespondError(w, http.StatusUnauthorized, "Please verify your email first before logging in...")
		return
	}

	token, err := utils.GenerateToken(user.Username, user.ID, user.Email)
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

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Logged in successfully"})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.AuthTokenCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   secureToken,
		SameSite: http.SameSiteLaxMode,
	})

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully!"})
}

type GetMeResponse struct {
	Status int                 `json:"status"`
	User   *models.UserDetails `json:"user"`
}

func GetMe(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userId == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userDetails, err := userStore.GetUserDetailsById(userId)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, GetMeResponse{
		Status: http.StatusOK,
		User:   userDetails,
	})
}
