package repository

import (
	"quick-match/internal/models"
)

type InsertUserRepo interface {
	InsertUser(user models.UserDetails) error
}

type LoginUserRepo interface {
	GetUserByEmail(email string) (*models.UserDetails, error)
	InsertUserRepo
}

type InsertUserESRepo interface {
	InsertUserES(user models.UserDetailsES) error
}

type SwipeRepo interface {
	InsertSwipeRecord(swipe models.Swipe) error
	CheckSwipeMatch(swipedUserID, currentUserID string) (bool, error)
}

type GetSwipedUserRepo interface {
	GetSwipedUserIDs(userID string) ([]string, error)
}

type DiscoverRepo interface {
	GetUserByID(userID string) (models.UserDetailsES, error)
	SearchUsers(currentUserLocation models.UserLocationES, swipedUserIDs []string, discover models.DiscoverFilters) ([]models.UserDetailsES, error)
}
