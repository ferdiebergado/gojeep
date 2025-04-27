package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/model"
	"github.com/ferdiebergado/gojeep/internal/pkg/logging"
	"github.com/ferdiebergado/gojeep/internal/pkg/message"
	secMock "github.com/ferdiebergado/gojeep/internal/pkg/security/mock"

	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/ferdiebergado/gojeep/internal/service/mock"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	regURL         = "/auth/register"
	testEmail      = "abc@example.com"
	testPass       = "test"
	testPassHashed = "hashed"
)

type testCase struct {
	name           string
	request        handler.RegisterUserRequest
	setupMocks     func(mockService *mock.MockUserService)
	expectedStatus int
	expectedMsg    string
	verifyResponse func(t *testing.T, res handler.Response[handler.RegisterUserResponse])
}

var validate *validator.Validate

func TestMain(m *testing.M) {
	logging.SetLogger(os.Stderr, "testing", "error")
	validate = validator.New()

	os.Exit(m.Run())
}

func TestUserHandler_HandleUserRegister(t *testing.T) {
	user := model.User{
		Model: model.Model{ID: "1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Email: testEmail,
	}

	tests := []testCase{
		{
			name: "Success",
			request: handler.RegisterUserRequest{
				Email:           testEmail,
				Password:        testPass,
				PasswordConfirm: testPass,
			},
			setupMocks: func(mockService *mock.MockUserService) {
				mockService.EXPECT().
					RegisterUser(gomock.Any(), service.RegisterUserParams{
						Email:    testEmail,
						Password: testPass,
					}).
					Return(user, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedMsg:    message.UserRegSuccess,
			verifyResponse: func(t *testing.T, res handler.Response[handler.RegisterUserResponse]) {
				t.Helper()
				assert.Equal(t, user.ID, res.Data.ID)
				assert.Equal(t, user.Email, res.Data.Email)
				assert.NotZero(t, res.Data.CreatedAt)
				assert.NotZero(t, res.Data.UpdatedAt)
			},
		},
		{
			name: "Invalid input - empty email",
			request: handler.RegisterUserRequest{
				Password:        testPass,
				PasswordConfirm: testPass,
			},
			setupMocks: func(mockService *mock.MockUserService) {
				mockService.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    message.UserInputInvalid,
		},
		{
			name: "Invalid input - invalid email",
			request: handler.RegisterUserRequest{
				Email:           "not@nemail",
				Password:        testPass,
				PasswordConfirm: testPass,
			},
			setupMocks: func(mockService *mock.MockUserService) {
				mockService.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    message.UserInputInvalid,
		},
		{
			name: "Duplicate user",
			request: handler.RegisterUserRequest{
				Email:           testEmail,
				Password:        testPass,
				PasswordConfirm: testPass,
			},
			setupMocks: func(mockService *mock.MockUserService) {
				mockService.EXPECT().
					RegisterUser(gomock.Any(), service.RegisterUserParams{
						Email:    testEmail,
						Password: testPass,
					}).
					Return(model.User{}, service.ErrUserExists)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedMsg:    message.UserExists,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runRegisterUserTest(t, tc)
		})
	}
}

func runRegisterUserTest(t *testing.T, tc testCase) {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockUserService(ctrl)
	mockSigner := secMock.NewMockSigner(ctrl)

	cfg := &config.Config{
		JWT: &config.JWTOptions{
			Issuer:   "localhost:8888",
			Duration: 15,
		},
	}
	if tc.setupMocks != nil {
		tc.setupMocks(mockService)
	}

	userHandler := handler.NewUserHandler(mockService, mockSigner, cfg)
	registerHandler := handler.ValidateInput[handler.RegisterUserRequest](validate)(
		http.HandlerFunc(userHandler.HandleUserRegister))
	registerHandler = handler.DecodeJSON[handler.RegisterUserRequest]()(registerHandler)

	reqBody, err := json.Marshal(tc.request)
	require.NoError(t, err, "failed to marshal request")

	req := httptest.NewRequest(http.MethodPost, regURL, bytes.NewReader(reqBody))
	req.Header.Set(handler.HeaderContentType, handler.MimeJSON)
	rec := httptest.NewRecorder()

	registerHandler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, tc.expectedStatus, res.StatusCode)
	assert.Equal(t, handler.MimeJSON, res.Header.Get(handler.HeaderContentType))

	var apiRes handler.Response[handler.RegisterUserResponse]
	require.NoError(t, json.NewDecoder(res.Body).Decode(&apiRes), "failed to decode response")

	assert.Equal(t, tc.expectedMsg, apiRes.Message)

	if tc.verifyResponse != nil {
		tc.verifyResponse(t, apiRes)
	}
}

func TestUserHandler_HandleUserLogin(t *testing.T) {
	t.Parallel()
	const url = "/auth/login"
	ctrl := gomock.NewController(t)

	tests := []struct {
		name            string
		email           string
		password        string
		expectedStatus  int
		expectedMessage string
		mockServiceCall func(mockService *mock.MockUserService)
	}{
		{
			name:            "Valid login credentials",
			email:           testEmail,
			password:        testPass,
			expectedStatus:  http.StatusOK,
			expectedMessage: message.UserLoginSuccess,
			mockServiceCall: func(mockService *mock.MockUserService) {
				mockService.EXPECT().
					LoginUser(gomock.Any(), service.LoginUserParams{
						Email:    testEmail,
						Password: testPass,
					}).
					Return("mock_access_token", "mock_refresh_token", nil)
			},
		},
		{
			name:            "Invalid email",
			email:           "notanemail",
			password:        testPass,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: message.UserInputInvalid,
			mockServiceCall: func(mockService *mock.MockUserService) {
			},
		},
		{
			name:            "Blank email",
			password:        testPass,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: message.UserInputInvalid,
			mockServiceCall: func(mockService *mock.MockUserService) {
			},
		},
		{
			name:            "Blank password",
			email:           testEmail,
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: message.UserInputInvalid,
			mockServiceCall: func(mockService *mock.MockUserService) {
			},
		},
		{
			name:            "Login failure",
			email:           testEmail,
			password:        "wrongpass",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: message.UserNotFound,
			mockServiceCall: func(mockService *mock.MockUserService) {
				mockService.EXPECT().
					LoginUser(gomock.Any(), service.LoginUserParams{
						Email:    testEmail,
						Password: "wrongpass",
					}).
					Return("", "", service.ErrUserNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockService := mock.NewMockUserService(ctrl)
			mockSigner := secMock.NewMockSigner(ctrl)

			cfg := &config.Config{
				JWT: &config.JWTOptions{
					Issuer:   "localhost:8888",
					Duration: 15,
				},
			}
			tt.mockServiceCall(mockService)

			userHandler := handler.NewUserHandler(mockService, mockSigner, cfg)
			userLoginHandler := handler.ValidateInput[handler.UserLoginRequest](validate)(
				http.HandlerFunc(userHandler.HandleUserLogin))
			userLoginHandler = handler.DecodeJSON[handler.UserLoginRequest]()(userLoginHandler)

			loginRequest := handler.UserLoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}
			reqBody, err := json.Marshal(loginRequest)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
			req.Header.Set(handler.HeaderContentType, handler.MimeJSON)
			rec := httptest.NewRecorder()

			userLoginHandler.ServeHTTP(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			if tt.expectedMessage != "" {
				var apiRes handler.Response[any]
				require.NoError(t, json.NewDecoder(res.Body).Decode(&apiRes), "failed to decode response")
				assert.Equal(t, tt.expectedMessage, apiRes.Message)
			}
		})
	}
}
