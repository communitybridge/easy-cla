// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

// Code generated by MockGen. DO NOT EDIT.
// Source: users/repository.go

// Package mock_users is a generated GoMock package.
package mock_users

import (
	reflect "reflect"

	models "github.com/communitybridge/easycla/cla-backend-go/gen/v1/models"
	gomock "github.com/golang/mock/gomock"
)

// MockUserRepository is a mock of UserRepository interface.
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository.
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance.
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// CreateUser mocks base method.
func (m *MockUserRepository) CreateUser(user *models.User) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", user)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUserRepositoryMockRecorder) CreateUser(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUserRepository)(nil).CreateUser), user)
}

// Delete mocks base method.
func (m *MockUserRepository) Delete(userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockUserRepositoryMockRecorder) Delete(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUserRepository)(nil).Delete), userID)
}

// GetUser mocks base method.
func (m *MockUserRepository) GetUser(userID string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", userID)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockUserRepositoryMockRecorder) GetUser(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockUserRepository)(nil).GetUser), userID)
}

// GetUserByEmail mocks base method.
func (m *MockUserRepository) GetUserByEmail(userEmail string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByEmail", userEmail)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByEmail indicates an expected call of GetUserByEmail.
func (mr *MockUserRepositoryMockRecorder) GetUserByEmail(userEmail interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByEmail", reflect.TypeOf((*MockUserRepository)(nil).GetUserByEmail), userEmail)
}

// GetUserByExternalID mocks base method.
func (m *MockUserRepository) GetUserByExternalID(userExternalID string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByExternalID", userExternalID)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByExternalID indicates an expected call of GetUserByExternalID.
func (mr *MockUserRepositoryMockRecorder) GetUserByExternalID(userExternalID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByExternalID", reflect.TypeOf((*MockUserRepository)(nil).GetUserByExternalID), userExternalID)
}

// GetUserByGitHubID mocks base method.
func (m *MockUserRepository) GetUserByGitHubID(gitHubID string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByGitHubID", gitHubID)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByGitHubID indicates an expected call of GetUserByGitHubID.
func (mr *MockUserRepositoryMockRecorder) GetUserByGitHubID(gitHubID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByGitHubID", reflect.TypeOf((*MockUserRepository)(nil).GetUserByGitHubID), gitHubID)
}

// GetUserByGitHubUsername mocks base method.
func (m *MockUserRepository) GetUserByGitHubUsername(gitHubUsername string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByGitHubUsername", gitHubUsername)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByGitHubUsername indicates an expected call of GetUserByGitHubUsername.
func (mr *MockUserRepositoryMockRecorder) GetUserByGitHubUsername(gitHubUsername interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByGitHubUsername", reflect.TypeOf((*MockUserRepository)(nil).GetUserByGitHubUsername), gitHubUsername)
}

// GetUserByGitLabUsername mocks base method.
func (m *MockUserRepository) GetUserByGitLabUsername(gitlabUsername string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByGitLabUsername", gitlabUsername)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByGitLabUsername indicates an expected call of GetUserByGitLabUsername.
func (mr *MockUserRepositoryMockRecorder) GetUserByGitLabUsername(gitlabUsername interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByGitLabUsername", reflect.TypeOf((*MockUserRepository)(nil).GetUserByGitLabUsername), gitlabUsername)
}

// GetUserByGitlabID mocks base method.
func (m *MockUserRepository) GetUserByGitlabID(gitlabID int) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByGitlabID", gitlabID)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByGitlabID indicates an expected call of GetUserByGitlabID.
func (mr *MockUserRepositoryMockRecorder) GetUserByGitlabID(gitlabID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByGitlabID", reflect.TypeOf((*MockUserRepository)(nil).GetUserByGitlabID), gitlabID)
}

// GetUserByLFUserName mocks base method.
func (m *MockUserRepository) GetUserByLFUserName(lfUserName string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByLFUserName", lfUserName)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByLFUserName indicates an expected call of GetUserByLFUserName.
func (mr *MockUserRepositoryMockRecorder) GetUserByLFUserName(lfUserName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByLFUserName", reflect.TypeOf((*MockUserRepository)(nil).GetUserByLFUserName), lfUserName)
}

// GetUserByUserName mocks base method.
func (m *MockUserRepository) GetUserByUserName(userName string, fullMatch bool) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByUserName", userName, fullMatch)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByUserName indicates an expected call of GetUserByUserName.
func (mr *MockUserRepositoryMockRecorder) GetUserByUserName(userName, fullMatch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByUserName", reflect.TypeOf((*MockUserRepository)(nil).GetUserByUserName), userName, fullMatch)
}

// GetUsersByEmail mocks base method.
func (m *MockUserRepository) GetUsersByEmail(userEmail string) ([]*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUsersByEmail", userEmail)
	ret0, _ := ret[0].([]*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsersByEmail indicates an expected call of GetUsersByEmail.
func (mr *MockUserRepositoryMockRecorder) GetUsersByEmail(userEmail interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsersByEmail", reflect.TypeOf((*MockUserRepository)(nil).GetUsersByEmail), userEmail)
}

// Save mocks base method.
func (m *MockUserRepository) Save(user *models.UserUpdate) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", user)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Save indicates an expected call of Save.
func (mr *MockUserRepositoryMockRecorder) Save(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockUserRepository)(nil).Save), user)
}

// SearchUsers mocks base method.
func (m *MockUserRepository) SearchUsers(searchField, searchTerm string, fullMatch bool) (*models.Users, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchUsers", searchField, searchTerm, fullMatch)
	ret0, _ := ret[0].(*models.Users)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchUsers indicates an expected call of SearchUsers.
func (mr *MockUserRepositoryMockRecorder) SearchUsers(searchField, searchTerm, fullMatch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchUsers", reflect.TypeOf((*MockUserRepository)(nil).SearchUsers), searchField, searchTerm, fullMatch)
}

// UpdateUser mocks base method.
func (m *MockUserRepository) UpdateUser(userID string, updates map[string]interface{}) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", userID, updates)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockUserRepositoryMockRecorder) UpdateUser(userID, updates interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockUserRepository)(nil).UpdateUser), userID, updates)
}

// UpdateUserCompanyID mocks base method.
func (m *MockUserRepository) UpdateUserCompanyID(userID, companyID, note string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserCompanyID", userID, companyID, note)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUserCompanyID indicates an expected call of UpdateUserCompanyID.
func (mr *MockUserRepositoryMockRecorder) UpdateUserCompanyID(userID, companyID, note interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserCompanyID", reflect.TypeOf((*MockUserRepository)(nil).UpdateUserCompanyID), userID, companyID, note)
}
