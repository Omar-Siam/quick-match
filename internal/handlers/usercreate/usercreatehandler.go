package usercreate

import (
	"encoding/json"
	"log"
	"net/http"
	"quick-match/internal/repository"
	"quick-match/internal/services"
)

type CreateUserDeps struct {
	UserRepoES repository.InsertUserESRepo
	UserRepo   repository.InsertUserRepo
}

/*
CreateUserHandler generates new users. The process involves the following steps:
Generates a new user entity using the GenerateNewUser function from the services package. This entity includes all necessary details for a new user.
Inserts the new generated user into DynamoDB & ElasticSearch with sensitive data stripped.
*/
func CreateUserHandler(deps *CreateUserDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newUser := services.GenerateNewUser()

		err := deps.UserRepo.InsertUser(newUser)
		if err != nil {
			log.Printf("Query Failure: %v", err)
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			return
		}

		userES := repository.CreateElasticSearchUser(newUser)
		err = deps.UserRepoES.InsertUserES(userES)
		if err != nil {
			log.Printf("Query Failure: %v", err)
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(newUser); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
