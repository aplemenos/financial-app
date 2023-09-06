package healthchecks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Alive(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHealthcheckHandler_AliveCheck(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	handler := &HealthcheckHandler{Service: mockService, Logger: logger.Sugar()}

	r := gin.Default()
	r.GET("/", handler.aliveCheck)

	testCases := []struct {
		Name          string
		ExpectedError error
		ExpectedCode  int
	}{
		{
			Name:          "Service Alive",
			ExpectedError: nil,
			ExpectedCode:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockService.On("Alive", mock.Anything).Return(tc.ExpectedError)

			req, _ := http.NewRequest("GET", "/", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.ExpectedCode, rr.Code)

			// If an error is expected, assert the error in the body
			if tc.ExpectedError != nil {
				errorResponse := strings.TrimSuffix(rr.Body.String(), "\n")
				assert.Equal(t, tc.ExpectedError.Error(), errorResponse)
			}
		})
	}
}
