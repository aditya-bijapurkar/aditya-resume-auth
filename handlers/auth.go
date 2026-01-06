package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/aditya-bijapurkar/aditya-resume-auth/middleware/auth"
	"github.com/aditya-bijapurkar/aditya-resume-auth/models"
	"github.com/aditya-bijapurkar/aditya-resume-auth/services"
	"github.com/aditya-bijapurkar/aditya-resume-auth/utils"
	"golang.org/x/crypto/bcrypt"
)

var userStore *models.UserStore
var emailService *services.EmailService

func SetUserStore(store *models.UserStore) {
	userStore = store
}

func SetEmailService(service *services.EmailService) {
	emailService = service
}

type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
}

type GetMeResponse struct {
	Status int                 `json:"status"`
	User   *models.UserDetails `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	if len(req.Password) < 6 {
		respondError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	if err := userStore.CreateUser(user); err != nil {
		if err == models.ErrUserExists {
			respondError(w, http.StatusConflict, "User already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	verificationToken, err := utils.GenerateVerificationToken(user.Email)
	if err == nil && emailService != nil {
		emailService.SendVerificationEmailAsync(user.Email, verificationToken)
	}

	respondJSON(w, http.StatusCreated, SignupResponse{
		Status:  http.StatusCreated,
		Message: "User created successfully",
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := userStore.GetUserByEmail(req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Email or password is incorrect, please try again")
		return
	}
	if !user.IsVerified {
		respondError(w, http.StatusUnauthorized, "Please verify your email first before logging in...")
		return
	}

	token, err := utils.GenerateToken(user.Username, user.ID, user.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, LoginResponse{
		Status: http.StatusOK,
		Token:  token,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func GetMe(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(auth.UserEmailKey).(string)
	if !ok || userEmail == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userDetails, err := userStore.GetUserDetailsByEmail(userEmail)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GetMeResponse{
		Status: http.StatusOK,
		User:   userDetails,
	})
}

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "Verification token is required")
		return
	}

	claims, err := utils.ValidateVerificationToken(token)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid or expired verification token")
		return
	}

	if err := userStore.UpdateUserVerification(claims.Email); err != nil {
		if err == models.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	frontendURL := os.Getenv("REDIRECT_URL")
	http.Redirect(w, r, frontendURL, http.StatusFound)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
