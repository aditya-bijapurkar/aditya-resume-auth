package handlers

import (
	"os"

	"github.com/aditya-bijapurkar/aditya-resume-auth/models"
	"github.com/aditya-bijapurkar/aditya-resume-auth/services"
)

var userStore *models.UserStore
var emailService *services.EmailService

func SetUserStore(store *models.UserStore) {
	userStore = store
}

func SetEmailService(service *services.EmailService) {
	emailService = service
}

var secureToken bool = os.Getenv("ENV") == "prod"
