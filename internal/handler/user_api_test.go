package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ferdiebergado/goexpress"
	"github.com/ferdiebergado/goweb/internal/handler"
	"github.com/ferdiebergado/goweb/internal/model"
	"github.com/ferdiebergado/goweb/internal/pkg/message"
	"github.com/ferdiebergado/goweb/internal/service"
	"github.com/ferdiebergado/goweb/internal/service/mock"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	regUrl         = "/api/auth/register"
	testEmail      = "abc@example.com"
	testPass       = "test"
	testPassHashed = "hashed"
)

var validate *validator.Validate

func TestMain(t *testing.M) {
	validate = validator.New()
	t.Run()
}

func TestUserHandlerHandleUserRegisterSuccess(t *testing.T) {
	msg := message.Get("regSuccess")

	ctrl := gomock.NewController(t)
	mockService := mock.NewMockUserService(ctrl)
	regRequest := handler.RegisterUserRequest{
		Email:           testEmail,
		Password:        testPass,
		PasswordConfirm: testPass,
	}
	regParams := service.RegisterUserParams{
		Email:    testEmail,
		Password: testPass,
	}

	user := &model.User{
		Model: model.Model{ID: "1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Email: testEmail,
	}

	mockService.EXPECT().RegisterUser(handler.NewParamsContext(context.Background(), regRequest), regParams).Return(user, nil)
	userHandler := handler.NewUserAPIHandler(mockService)
	r := goexpress.New()
	r.Post(regUrl, userHandler.HandleUserRegister,
		handler.DecodeJSON[handler.RegisterUserRequest](), handler.ValidateInput[handler.RegisterUserRequest](validate))

	reqJSON, err := json.Marshal(regRequest)

	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", regUrl, bytes.NewBuffer(reqJSON))
	req.Header.Set(handler.HeaderContentType, handler.MimeJSONUTF8)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, handler.MimeJSONUTF8, res.Header[handler.HeaderContentType][0])

	var apiRes handler.APIResponse[handler.RegisterUserResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &apiRes); err != nil {
		t.Fatal(message.Get("jsonFailed"), err)
	}

	assert.Equal(t, msg, apiRes.Message)
	assert.Equal(t, user.ID, apiRes.Data.ID)
	assert.Equal(t, user.Email, apiRes.Data.Email)
	assert.NotZero(t, apiRes.Data.CreatedAt)
	assert.NotZero(t, apiRes.Data.UpdatedAt)
}

func TestUserHandlerHandleUserRegisterInvalidInput(t *testing.T) {
	const msg = "Invalid input."

	ctrl := gomock.NewController(t)
	mockService := mock.NewMockUserService(ctrl)
	userHandler := handler.NewUserAPIHandler(mockService)
	r := goexpress.New()
	r.Post(regUrl, userHandler.HandleUserRegister,
		handler.DecodeJSON[handler.RegisterUserRequest](), handler.ValidateInput[handler.RegisterUserRequest](validate))

	var tests = []struct {
		name       string
		regRequest handler.RegisterUserRequest
	}{
		{"Empty email", handler.RegisterUserRequest{Password: testPass, PasswordConfirm: testPass}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Times(0)
			reqJSON, err := json.Marshal(tt.regRequest)

			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest("POST", regUrl, bytes.NewBuffer(reqJSON))
			req.Header.Set(handler.HeaderContentType, handler.MimeJSONUTF8)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, handler.MimeJSONUTF8, res.Header[handler.HeaderContentType][0])
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)

			var apiRes handler.APIResponse[handler.RegisterUserResponse]
			if err := json.Unmarshal(rr.Body.Bytes(), &apiRes); err != nil {
				t.Fatal(message.Get("jsonFailed"), err)
			}

			assert.Equal(t, msg, apiRes.Message)
		})
	}
}

func TestUserHandlerHandleUserRegisterDuplicateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := mock.NewMockUserService(ctrl)
	regRequest := handler.RegisterUserRequest{
		Email:           testEmail,
		Password:        testPass,
		PasswordConfirm: testPass,
	}
	regParams := service.RegisterUserParams{
		Email:    testEmail,
		Password: testPass,
	}

	mockService.EXPECT().RegisterUser(handler.NewParamsContext(context.Background(), regRequest), regParams).Return(nil, service.ErrDuplicateUser)
	userHandler := handler.NewUserAPIHandler(mockService)
	r := goexpress.New()
	r.Post(regUrl, userHandler.HandleUserRegister,
		handler.DecodeJSON[handler.RegisterUserRequest](), handler.ValidateInput[handler.RegisterUserRequest](validate))

	reqJSON, err := json.Marshal(regRequest)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, regUrl, bytes.NewBuffer(reqJSON))
	req.Header.Set(handler.HeaderContentType, handler.MimeJSONUTF8)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
	assert.Equal(t, handler.MimeJSONUTF8, res.Header[handler.HeaderContentType][0])

	var apiRes handler.APIResponse[handler.RegisterUserResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &apiRes); err != nil {
		t.Fatal(message.Get("jsonFailed"), err)
	}

	assert.Equal(t, service.ErrDuplicateUser.Error(), apiRes.Message)
}
