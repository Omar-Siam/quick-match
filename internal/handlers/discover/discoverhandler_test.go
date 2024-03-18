package discover

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"quick-match/internal/models"
	"testing"
)

type MockGetSwipedUserRepo struct {
	mock.Mock
}

func (m *MockGetSwipedUserRepo) GetSwipedUserIDs(userID string) ([]string, error) {
	args := m.Called(userID)
	return args.Get(0).([]string), args.Error(1)
}

type MockDiscoverRepo struct {
	mock.Mock
}

func (m *MockDiscoverRepo) GetUserByID(userID string) (models.UserDetailsES, error) {
	args := m.Called(userID)
	return args.Get(0).(models.UserDetailsES), args.Error(1)
}

func (m *MockDiscoverRepo) SearchUsers(currentUserLocation models.UserLocationES, swipedUserIDs []string, df models.DiscoverFilters) ([]models.UserDetailsES, error) {
	args := m.Called(currentUserLocation, swipedUserIDs, df)
	return args.Get(0).([]models.UserDetailsES), args.Error(1)
}

func TestDiscoverUserInsert(t *testing.T) {
	tests := []struct {
		name             string
		body             models.DiscoverFilters
		setupMocks       func(*MockGetSwipedUserRepo, *MockDiscoverRepo)
		expectedStatus   int
		expectError      bool
		expectedErrorMsg string
		userIDInContext  string
	}{
		{
			name: "successful discovery",
			body: models.DiscoverFilters{},
			setupMocks: func(mg *MockGetSwipedUserRepo, md *MockDiscoverRepo) {
				mg.On("GetSwipedUserIDs", "userID").Return([]string{}, nil)
				md.On("GetUserByID", "userID").Return(models.UserDetailsES{}, nil)
				md.On("SearchUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.UserDetailsES{}, nil)
			},
			expectedStatus:  http.StatusOK,
			userIDInContext: "userID",
		},
		{
			name: "query failure on fetching swiped user IDs",
			body: models.DiscoverFilters{},
			setupMocks: func(mg *MockGetSwipedUserRepo, md *MockDiscoverRepo) {
				mg.On("GetSwipedUserIDs", "userID").Return(([]string)(nil), errors.New("database query error"))
			},
			expectedStatus:   http.StatusInternalServerError,
			expectError:      true,
			expectedErrorMsg: "Failed to find Swiped IDs",
			userIDInContext:  "userID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetSwipedUserRepo := new(MockGetSwipedUserRepo)
			mockDiscoverRepo := new(MockDiscoverRepo)
			tt.setupMocks(mockGetSwipedUserRepo, mockDiscoverRepo)

			deps := DiscoverUserDeps{
				UserRepo:   mockGetSwipedUserRepo,
				UserRepoES: mockDiscoverRepo,
			}

			handler := DiscoverUserInsert(&deps)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/discover", bytes.NewBuffer(bodyBytes))
			ctx := context.WithValue(req.Context(), "UserID", tt.userIDInContext)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectError {
				responseBody := rr.Body.String()
				assert.Contains(t, responseBody, tt.expectedErrorMsg, "Error message does not match")
			}

			mockGetSwipedUserRepo.AssertExpectations(t)
			mockDiscoverRepo.AssertExpectations(t)
		})
	}
}
