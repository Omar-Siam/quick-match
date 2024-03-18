package swipe

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"quick-match/internal/models"
	"testing"

	"context"
	"github.com/stretchr/testify/assert"
)

type MockSwipeRepo struct {
	mock.Mock
}

func (m *MockSwipeRepo) InsertSwipeRecord(swipe models.Swipe) error {
	args := m.Called(swipe)
	return args.Error(0)
}

func (m *MockSwipeRepo) CheckSwipeMatch(swipedUserID, currentUserID string) (bool, error) {
	args := m.Called(swipedUserID, currentUserID)
	return args.Bool(0), args.Error(1)
}

func TestSwipeHandler(t *testing.T) {
	tests := []struct {
		name             string
		body             models.Swipe
		mockSetup        func(m *MockSwipeRepo)
		expectedStatus   int
		expectedResponse models.SwipeResponse
		userID           string
	}{
		{
			name: "dislike swipe",
			body: models.Swipe{SwipedUserID: "user2", Preference: false},
			mockSetup: func(m *MockSwipeRepo) {
				m.On("InsertSwipeRecord", mock.AnythingOfType("models.Swipe")).Return(nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: models.SwipeResponse{Matched: false},
			userID:           "user1",
		},
		{
			name: "like swipe with no match",
			body: models.Swipe{SwipedUserID: "user2", Preference: true},
			mockSetup: func(m *MockSwipeRepo) {
				m.On("CheckSwipeMatch", "user2", "user1").Return(false, nil)
				m.On("InsertSwipeRecord", mock.AnythingOfType("models.Swipe")).Return(nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: models.SwipeResponse{Matched: false},
			userID:           "user1",
		},
		{
			name: "like swipe with match",
			body: models.Swipe{SwipedUserID: "user2", Preference: true},
			mockSetup: func(m *MockSwipeRepo) {
				m.On("CheckSwipeMatch", "user2", "user1").Return(true, nil)
				m.On("InsertSwipeRecord", mock.AnythingOfType("models.Swipe")).Return(nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: models.SwipeResponse{Matched: true}, // MatchID not tested here due to randomness
			userID:           "user1",
		},
		{
			name: "error on swipe record insertion",
			body: models.Swipe{SwipedUserID: "user2", Preference: false},
			mockSetup: func(m *MockSwipeRepo) {
				m.On("InsertSwipeRecord", mock.AnythingOfType("models.Swipe")).Return(errors.New("db error"))
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: models.SwipeResponse{},
			userID:           "user1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockSwipeRepo)
			tt.mockSetup(mockRepo)

			deps := SwipeDeps{
				SwipeRepo: mockRepo,
			}

			handler := SwipeHandler(&deps)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/swipe", bytes.NewBuffer(bodyBytes))
			req = req.WithContext(context.WithValue(req.Context(), "UserID", tt.userID)) // Simulate JWTMiddleware setting UserID in context

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response models.SwipeResponse
			if err := json.NewDecoder(rr.Body).Decode(&response); err == nil {
				assert.Equal(t, tt.expectedResponse.Matched, response.Matched)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
