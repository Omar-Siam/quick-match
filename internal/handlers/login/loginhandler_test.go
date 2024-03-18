package login

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"quick-match/internal/models"
	"testing"
)

type MockLoginUserRepo struct {
	mock.Mock
}

func (m *MockLoginUserRepo) GetUserByEmail(email string) (*models.UserDetails, error) {
	args := m.Called(email)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*models.UserDetails), args.Error(1)
}

func (m *MockLoginUserRepo) InsertUser(user models.UserDetails) error {
	args := m.Called(user)
	return args.Error(0)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) CompareHashAndPassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

func (m *MockPasswordService) GenerateHashedPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name             string
		body             models.LoginCredentials
		setupMocks       func(*MockLoginUserRepo, *MockTokenService, *MockPasswordService)
		expectedStatus   int
		expectedResponse *models.LoginResponse
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "successful login",
			body: models.LoginCredentials{Email: "user@example.com", Password: "password"},
			setupMocks: func(mr *MockLoginUserRepo, mt *MockTokenService, mp *MockPasswordService) {
				mr.On("GetUserByEmail", "user@example.com").Return(&models.UserDetails{UserID: "123", Email: "user@example.com", PasswordHashed: "hashedpassword"}, nil)
				mp.On("CompareHashAndPassword", "hashedpassword", "password").Return(nil)
				mt.On("GenerateToken", "123").Return("token123", nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: &models.LoginResponse{Token: "token123"},
		},
		{
			name:             "validation failure",
			body:             models.LoginCredentials{Email: "invalidemail", Password: "password"},
			setupMocks:       func(mr *MockLoginUserRepo, mt *MockTokenService, mp *MockPasswordService) {},
			expectedStatus:   http.StatusBadRequest,
			expectError:      true,
			expectedErrorMsg: "Invalid email or password",
		},
		{
			name: "user not found",
			body: models.LoginCredentials{Email: "missing@example.com", Password: "password"},
			setupMocks: func(mr *MockLoginUserRepo, mt *MockTokenService, mp *MockPasswordService) {
				mr.On("GetUserByEmail", "missing@example.com").Return(nil, errors.New("user not found"))
			},
			expectedStatus:   http.StatusInternalServerError,
			expectError:      true,
			expectedErrorMsg: "Server error",
		},
		{
			name: "incorrect password",
			body: models.LoginCredentials{Email: "user@example.com", Password: "wrongpassword"},
			setupMocks: func(mr *MockLoginUserRepo, mt *MockTokenService, mp *MockPasswordService) {
				mr.On("GetUserByEmail", "user@example.com").Return(&models.UserDetails{UserID: "123", Email: "user@example.com", PasswordHashed: "hashedpassword"}, nil)
				mp.On("CompareHashAndPassword", "hashedpassword", "wrongpassword").Return(errors.New("incorrect password"))
			},
			expectedStatus:   http.StatusUnauthorized,
			expectError:      true,
			expectedErrorMsg: "Invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockLoginUserRepo)
			mockTokenService := new(MockTokenService)
			mockPasswordService := new(MockPasswordService)
			tt.setupMocks(mockRepo, mockTokenService, mockPasswordService)

			deps := LoginDeps{
				UserRepo:        mockRepo,
				TokenService:    mockTokenService,
				PasswordService: mockPasswordService,
			}

			handler := LoginHandler(&deps)

			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectError {
				responseBody := rr.Body.String()
				assert.Contains(t, responseBody, tt.expectedErrorMsg, "Error message does not match")
			} else {
				var response models.LoginResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse.Token, response.Token)
			}

			mockRepo.AssertExpectations(t)
			mockTokenService.AssertExpectations(t)
			mockPasswordService.AssertExpectations(t)
		})
	}
}
