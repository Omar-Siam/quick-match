package swipe

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/http"
	"quick-match/internal/models"
	"quick-match/internal/repository"
)

type SwipeDeps struct {
	SwipeRepo repository.SwipeRepo
}

/*
SwipeHandler processes swipe actions (like or dislike) between users.
Extracts the UserID from the request context
Updates the Swipe model with the UserID to associate the swipe action with the correct user.
If the swipe preference is false (dislike), it simply inserts the swipe record into the repository and sets the SwipeResponse's matched field to false.
If the swipe preference is true (like), it checks if the swiped user has also swiped right (liked) on the current user, indicating a potential match.
If a match is found, it generates a unique MatchID, updates the swipe action to indicate a match, and inserts the record into the repository.
The SwipeResponse includes the MatchID and indicates a successful match.
If no match is found, it inserts the swipe action as a non-matching action into the repository, and the SwipeResponse indicates no match.
*/
func SwipeHandler(deps *SwipeDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s models.Swipe
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
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
		s.UserID = UserID

		var sp models.SwipeResponse
		if s.Preference == false {
			sp.Matched = false
			// Insert the swipe record
			if err := deps.SwipeRepo.InsertSwipeRecord(s); err != nil {
				log.Printf("Query Failure: %v", err)
				http.Error(w, "Failed to insert swipe record", http.StatusInternalServerError)
				return
			}
		}

		if s.Preference == true {
			// Check if the swiped user has swiped "yes" on the current user
			isMatch, err := deps.SwipeRepo.CheckSwipeMatch(s.SwipedUserID, UserID)
			if err != nil {
				log.Printf("Query Failure: %v", err)
				http.Error(w, "Failed to check for swipe match", http.StatusInternalServerError)
				return
			}
			if isMatch == true {
				// It's a match! Generate a unique MatchID
				s.MatchID = uuid.New().String()
				s.Matched = true
				if err = deps.SwipeRepo.InsertSwipeRecord(s); err != nil {
					log.Printf("Query Failure: %v", err)
					http.Error(w, "Failed to insert swipe record", http.StatusInternalServerError)
					return
				}
				sp.MatchID = s.MatchID
				sp.Matched = true
			} else {
				sp.Matched = false
				s.Matched = false
				// Insert the swipe record
				if err = deps.SwipeRepo.InsertSwipeRecord(s); err != nil {
					log.Printf("Query Failure: %v", err)
					http.Error(w, "Failed to insert swipe record", http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(sp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
