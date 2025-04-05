package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ferdiebergado/goexpress"
	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/model"
	"github.com/ferdiebergado/gojeep/internal/pkg/message"
	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/ferdiebergado/gojeep/internal/service/mock"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	regURL         = "/api/auth/register"
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

func TestUserHandler_HandleUserRegister(t *testing.T) {
	user := &model.User{
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
			expectedMsg:    message.Get("regSuccess"),
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
			expectedMsg:    message.Get("invalidInput"),
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
			expectedMsg:    message.Get("invalidInput"),
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
					Return(nil, service.ErrDuplicateUser)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedMsg:    service.ErrDuplicateUser.Error(),
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

	validate := validator.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock.NewMockUserService(ctrl)

	if tc.setupMocks != nil {
		tc.setupMocks(mockService)
	}

	userHandler := handler.NewUserHandler(mockService)

	r := goexpress.New()
	r.Post(regURL, userHandler.HandleUserRegister,
		handler.DecodeJSON[handler.RegisterUserRequest](),
		handler.ValidateInput[handler.RegisterUserRequest](validate),
	)

	reqBody, err := json.Marshal(tc.request)
	require.NoError(t, err, "failed to marshal request")

	req := httptest.NewRequest(http.MethodPost, regURL, bytes.NewReader(reqBody))
	req.Header.Set(handler.HeaderContentType, handler.MimeJSON)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

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
