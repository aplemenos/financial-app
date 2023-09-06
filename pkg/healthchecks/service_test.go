package healthchecks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHealthcheckRepository struct {
	mock.Mock
}

func (m *MockHealthcheckRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestService_Alive(t *testing.T) {
	mockRepository := new(MockHealthcheckRepository)
	service := &service{healthcheck: mockRepository}

	testCases := []struct {
		Name          string
		MockRepoError error
		ExpectedError error
	}{
		{
			Name:          "Repository Alive",
			MockRepoError: nil,
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockRepository.On("Ping", mock.Anything).Return(tc.MockRepoError)

			err := service.Alive(context.Background())

			assert.Equal(t, tc.ExpectedError, err)
		})
	}
}
