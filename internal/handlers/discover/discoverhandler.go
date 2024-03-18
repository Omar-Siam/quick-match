package discover

import (
	"encoding/json"
	"log"
	"net/http"
	"quick-match/internal/models"
	"quick-match/internal/repository"
)

type DiscoverUserDeps struct {
	UserRepo   repository.GetSwipedUserRepo
	UserRepoES repository.DiscoverRepo
}

/*
DiscoverUserInsert processes user discovery requests.
User IDs that the authenticated user has already swiped on to exclude them from the discovery results.
Authenticated user's details are fetched from Elasticsearch, including their location, to be used in filtering compatible users.
Searches for compatible users based on the discovery filters provided and the authenticated user's location, excluding previously swiped users.
Any combination of filters can be provided. Non are mandatory.
*/
func DiscoverUserInsert(deps *DiscoverUserDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var df models.DiscoverFilters
		if err := json.NewDecoder(r.Body).Decode(&df); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Extract UserID from context, set by JWTMiddleware
		UserID, ok := r.Context().Value("UserID").(string)
		if !ok {
			log.Println("Could not extract UserID from token")
			http.Error(w, "Failed to authenticate", http.StatusInternalServerError)
			return
		}
		df.UserID = UserID

		swipedIDs, err := deps.UserRepo.GetSwipedUserIDs(UserID)
		if err != nil {
			log.Printf("Query Failure: %v", err)
			http.Error(w, "Failed to find Swiped IDs", http.StatusInternalServerError)
			return
		}

		user, err := deps.UserRepoES.GetUserByID(UserID)
		if err != nil {
			log.Printf("Query Failure: %v", err)
			http.Error(w, "Failed to fetch user from Elasticsearch", http.StatusInternalServerError)
			return
		}

		currentUserLocation := user.Location
		filteredUsers, err := deps.UserRepoES.SearchUsers(currentUserLocation, swipedIDs, df)
		if err != nil {
			log.Printf("Query Failure: %v", err)
			http.Error(w, "Failed to search users with df", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(filteredUsers); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
