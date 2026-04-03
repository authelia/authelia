package duo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	. "github.com/authelia/authelia/v4/internal/duo"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestAPIImpl_Call(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mock := mocks.NewMockDuoBaseProvider(ctrl)

	impl := NewDuoAPI(mock)

	assert.NotNil(t, impl.BaseProvider)
}
