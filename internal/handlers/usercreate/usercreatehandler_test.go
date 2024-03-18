package usercreate

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"quick-match/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) InsertUser(user models.UserDetails) error {
	args := m.Called(user)
	return args.Error(0)
}

type MockUserRepoES struct {
	mock.Mock
}

func (m *MockUserRepoES) InsertUserES(user models.UserDetailsES) error {
	args := m.Called(user)
	return args.Error(0)
}

func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name                 string
		mockUserRepo         func() *MockUserRepo
		mockUserRepoES       func() *MockUserRepoES
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "Successful User Creation",
			mockUserRepo: func() *MockUserRepo {
				m := new(MockUserRepo)
				m.On("InsertUser", mock.AnythingOfType("models.UserDetails")).Return(nil)
				return m
			},
			mockUserRepoES: func() *MockUserRepoES {
				m := new(MockUserRepoES)
				m.On("InsertUserES", mock.AnythingOfType("models.UserDetailsES")).Return(nil)
				return m
			},
			expectedStatus:       http.StatusCreated,
			expectedBodyContains: "",
		},
		{
			name: "Failure Inserting User into Repo",
			mockUserRepo: func() *MockUserRepo {
				m := new(MockUserRepo)
				m.On("InsertUser", mock.AnythingOfType("models.UserDetails")).Return(errors.New("insert user error"))
				return m
			},
			mockUserRepoES: func() *MockUserRepoES {
				m := new(MockUserRepoES)
				return m
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedBodyContains: "Failed to insert user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := tt.mockUserRepo()
			mockUserRepoES := tt.mockUserRepoES()

			deps := CreateUserDeps{
				UserRepo:   mockUserRepo,
				UserRepoES: mockUserRepoES,
			}

			handler := CreateUserHandler(&deps)

			req, err := http.NewRequest("POST", "/create", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handlerFunc := http.HandlerFunc(handler)

			handlerFunc.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedBodyContains != "" {
				body := rr.Body.String()
				assert.Contains(t, body, tt.expectedBodyContains, "Response body does not contain expected text")
			}

			mockUserRepo.AssertExpectations(t)
			mockUserRepoES.AssertExpectations(t)
		})
	}
}
