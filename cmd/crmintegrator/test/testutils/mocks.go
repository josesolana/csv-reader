package testutils

import (
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

type MockRows struct {
	mock.Mock
}

func NewMockDB() *MockDB {
	return new(MockDB)
}

func NewMockRows() *MockRows {
	return new(MockRows)
}

func (d *MockDB) Read(limit, offset int64) (*MockRows, error) {
	args := d.Called(limit, offset)
	mockRows := args.Get(0).(*MockRows)
	return mockRows, args.Error(1)
}

func (d *MockDB) Close() error {
	return d.Called().Error(0)
}
