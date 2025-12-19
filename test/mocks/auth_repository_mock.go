package mocks

import (
	model "UASBE/app/model/Postgresql"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) FindUserByEmailOrUsername(identifier string) (*model.Users, string, error) {
	args := m.Called(identifier)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*model.Users), args.String(1), args.Error(2)
}

func (m *MockAuthRepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]string, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
