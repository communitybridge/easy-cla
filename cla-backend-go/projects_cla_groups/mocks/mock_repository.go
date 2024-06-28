// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

// Code generated by MockGen. DO NOT EDIT.
// Source: projects_cla_groups/repository.go

// Package mock_projects_cla_groups is a generated GoMock package.
package mock_projects_cla_groups

import (
	context "context"
	reflect "reflect"

	projects_cla_groups "github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"
	gomock "github.com/golang/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// AssociateClaGroupWithProject mocks base method.
func (m *MockRepository) AssociateClaGroupWithProject(ctx context.Context, claGroupID, projectSFID, foundationSFID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AssociateClaGroupWithProject", ctx, claGroupID, projectSFID, foundationSFID)
	ret0, _ := ret[0].(error)
	return ret0
}

// AssociateClaGroupWithProject indicates an expected call of AssociateClaGroupWithProject.
func (mr *MockRepositoryMockRecorder) AssociateClaGroupWithProject(ctx, claGroupID, projectSFID, foundationSFID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AssociateClaGroupWithProject", reflect.TypeOf((*MockRepository)(nil).AssociateClaGroupWithProject), ctx, claGroupID, projectSFID, foundationSFID)
}

// GetCLAGroup mocks base method.
func (m *MockRepository) GetCLAGroup(ctx context.Context, claGroupID string) (*projects_cla_groups.ProjectClaGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCLAGroup", ctx, claGroupID)
	ret0, _ := ret[0].(*projects_cla_groups.ProjectClaGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCLAGroup indicates an expected call of GetCLAGroup.
func (mr *MockRepositoryMockRecorder) GetCLAGroup(ctx, claGroupID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCLAGroup", reflect.TypeOf((*MockRepository)(nil).GetCLAGroup), ctx, claGroupID)
}

// GetCLAGroupNameByID mocks base method.
func (m *MockRepository) GetCLAGroupNameByID(ctx context.Context, claGroupID string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCLAGroupNameByID", ctx, claGroupID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCLAGroupNameByID indicates an expected call of GetCLAGroupNameByID.
func (mr *MockRepositoryMockRecorder) GetCLAGroupNameByID(ctx, claGroupID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCLAGroupNameByID", reflect.TypeOf((*MockRepository)(nil).GetCLAGroupNameByID), ctx, claGroupID)
}

// GetClaGroupIDForProject mocks base method.
func (m *MockRepository) GetClaGroupIDForProject(ctx context.Context, projectSFID string) (*projects_cla_groups.ProjectClaGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClaGroupIDForProject", ctx, projectSFID)
	ret0, _ := ret[0].(*projects_cla_groups.ProjectClaGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClaGroupIDForProject indicates an expected call of GetClaGroupIDForProject.
func (mr *MockRepositoryMockRecorder) GetClaGroupIDForProject(ctx, projectSFID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClaGroupIDForProject", reflect.TypeOf((*MockRepository)(nil).GetClaGroupIDForProject), ctx, projectSFID)
}

// GetProjectsIdsForAllFoundation mocks base method.
func (m *MockRepository) GetProjectsIdsForAllFoundation(ctx context.Context) ([]*projects_cla_groups.ProjectClaGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProjectsIdsForAllFoundation", ctx)
	ret0, _ := ret[0].([]*projects_cla_groups.ProjectClaGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProjectsIdsForAllFoundation indicates an expected call of GetProjectsIdsForAllFoundation.
func (mr *MockRepositoryMockRecorder) GetProjectsIdsForAllFoundation(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProjectsIdsForAllFoundation", reflect.TypeOf((*MockRepository)(nil).GetProjectsIdsForAllFoundation), ctx)
}

// GetProjectsIdsForClaGroup mocks base method.
func (m *MockRepository) GetProjectsIdsForClaGroup(ctx context.Context, claGroupID string) ([]*projects_cla_groups.ProjectClaGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProjectsIdsForClaGroup", ctx, claGroupID)
	ret0, _ := ret[0].([]*projects_cla_groups.ProjectClaGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProjectsIdsForClaGroup indicates an expected call of GetProjectsIdsForClaGroup.
func (mr *MockRepositoryMockRecorder) GetProjectsIdsForClaGroup(ctx, claGroupID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProjectsIdsForClaGroup", reflect.TypeOf((*MockRepository)(nil).GetProjectsIdsForClaGroup), ctx, claGroupID)
}

// GetProjectsIdsForFoundation mocks base method.
func (m *MockRepository) GetProjectsIdsForFoundation(ctx context.Context, foundationSFID string) ([]*projects_cla_groups.ProjectClaGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProjectsIdsForFoundation", ctx, foundationSFID)
	ret0, _ := ret[0].([]*projects_cla_groups.ProjectClaGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProjectsIdsForFoundation indicates an expected call of GetProjectsIdsForFoundation.
func (mr *MockRepositoryMockRecorder) GetProjectsIdsForFoundation(ctx, foundationSFID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProjectsIdsForFoundation", reflect.TypeOf((*MockRepository)(nil).GetProjectsIdsForFoundation), ctx, foundationSFID)
}

// IsAssociated mocks base method.
func (m *MockRepository) IsAssociated(ctx context.Context, projectSFID, claGroupID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAssociated", ctx, projectSFID, claGroupID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsAssociated indicates an expected call of IsAssociated.
func (mr *MockRepositoryMockRecorder) IsAssociated(ctx, projectSFID, claGroupID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAssociated", reflect.TypeOf((*MockRepository)(nil).IsAssociated), ctx, projectSFID, claGroupID)
}

// IsExistingFoundationLevelCLAGroup mocks base method.
func (m *MockRepository) IsExistingFoundationLevelCLAGroup(ctx context.Context, foundationSFID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsExistingFoundationLevelCLAGroup", ctx, foundationSFID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsExistingFoundationLevelCLAGroup indicates an expected call of IsExistingFoundationLevelCLAGroup.
func (mr *MockRepositoryMockRecorder) IsExistingFoundationLevelCLAGroup(ctx, foundationSFID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsExistingFoundationLevelCLAGroup", reflect.TypeOf((*MockRepository)(nil).IsExistingFoundationLevelCLAGroup), ctx, foundationSFID)
}

// RemoveProjectAssociatedWithClaGroup mocks base method.
func (m *MockRepository) RemoveProjectAssociatedWithClaGroup(ctx context.Context, claGroupID string, projectSFIDList []string, all bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveProjectAssociatedWithClaGroup", ctx, claGroupID, projectSFIDList, all)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveProjectAssociatedWithClaGroup indicates an expected call of RemoveProjectAssociatedWithClaGroup.
func (mr *MockRepositoryMockRecorder) RemoveProjectAssociatedWithClaGroup(ctx, claGroupID, projectSFIDList, all interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveProjectAssociatedWithClaGroup", reflect.TypeOf((*MockRepository)(nil).RemoveProjectAssociatedWithClaGroup), ctx, claGroupID, projectSFIDList, all)
}

// UpdateClaGroupName mocks base method.
func (m *MockRepository) UpdateClaGroupName(ctx context.Context, projectSFID, claGroupName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateClaGroupName", ctx, projectSFID, claGroupName)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateClaGroupName indicates an expected call of UpdateClaGroupName.
func (mr *MockRepositoryMockRecorder) UpdateClaGroupName(ctx, projectSFID, claGroupName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateClaGroupName", reflect.TypeOf((*MockRepository)(nil).UpdateClaGroupName), ctx, projectSFID, claGroupName)
}

// UpdateRepositoriesCount mocks base method.
func (m *MockRepository) UpdateRepositoriesCount(ctx context.Context, projectSFID string, diff int64, reset bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateRepositoriesCount", ctx, projectSFID, diff, reset)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateRepositoriesCount indicates an expected call of UpdateRepositoriesCount.
func (mr *MockRepositoryMockRecorder) UpdateRepositoriesCount(ctx, projectSFID, diff, reset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateRepositoriesCount", reflect.TypeOf((*MockRepository)(nil).UpdateRepositoriesCount), ctx, projectSFID, diff, reset)
}
