package login

import (
	"encoding/json"
	"log"
	"net/http"
	"quick-match/internal/middleware/authentication"
	"quick-match/internal/middleware/validation"
	"quick-match/internal/models"
	"quick-match/internal/repository"
	"quick-match/internal/services"
)

type LoginDeps struct {
	UserRepo        repository.LoginUserRepo
	TokenService    authentication.TokenService
	PasswordService services.PasswordService
}

/*
LoginHandler processes login requests.
Validates the provided login credentials.
Attempts to retrieve the user by email from the repository.
Compares the provided password with the user's stored hashed password using the PasswordService.
Generates an authentication token for the user using the TokenService.
*/
func LoginHandler(deps *LoginDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var lc models.LoginCredentials
		if err := json.NewDecoder(r.Body).Decode(&lc); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := validation.ValidateLogin(lc); err != nil {
			log.Printf("Validation Failure: %v", err)
			http.Error(w, "Invalid email or password", http.StatusBadRequest)
			return
		}

		user, err := deps.UserRepo.GetUserByEmail(lc.Email)
		if err != nil {
			log.Printf("Query Failure: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		if err = deps.PasswordService.CompareHashAndPassword(user.PasswordHashed, lc.Password); err != nil {
			log.Printf("Password Dycrption Failure: %v", err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := deps.TokenService.GenerateToken(user.UserID)
		if err != nil {
			log.Printf("Token Generation Failure: %v", err)
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		response := models.LoginResponse{Token: token}
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
