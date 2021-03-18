// Code generated by MockGen. DO NOT EDIT.
// Source: contract.go

// Package github_test is a generated GoMock package.
package github_test

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	git "go.octolab.org/toolset/maintainer/internal/model/git"
	github "go.octolab.org/toolset/maintainer/internal/model/github"
)

// MockGit is a mock of Git interface.
type MockGit struct {
	ctrl     *gomock.Controller
	recorder *MockGitMockRecorder
}

// MockGitMockRecorder is the mock recorder for MockGit.
type MockGitMockRecorder struct {
	mock *MockGit
}

// NewMockGit creates a new mock instance.
func NewMockGit(ctrl *gomock.Controller) *MockGit {
	mock := &MockGit{ctrl: ctrl}
	mock.recorder = &MockGitMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGit) EXPECT() *MockGitMockRecorder {
	return m.recorder
}

// Remotes mocks base method.
func (m *MockGit) Remotes() (git.Remotes, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remotes")
	ret0, _ := ret[0].(git.Remotes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Remotes indicates an expected call of Remotes.
func (mr *MockGitMockRecorder) Remotes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remotes", reflect.TypeOf((*MockGit)(nil).Remotes))
}

// MockGitHub is a mock of GitHub interface.
type MockGitHub struct {
	ctrl     *gomock.Controller
	recorder *MockGitHubMockRecorder
}

// MockGitHubMockRecorder is the mock recorder for MockGitHub.
type MockGitHubMockRecorder struct {
	mock *MockGitHub
}

// NewMockGitHub creates a new mock instance.
func NewMockGitHub(ctrl *gomock.Controller) *MockGitHub {
	mock := &MockGitHub{ctrl: ctrl}
	mock.recorder = &MockGitHubMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGitHub) EXPECT() *MockGitHubMockRecorder {
	return m.recorder
}

// Labels mocks base method.
func (m *MockGitHub) Labels(arg0 context.Context, arg1 github.GitHub) ([]github.Label, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Labels", arg0, arg1)
	ret0, _ := ret[0].([]github.Label)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Labels indicates an expected call of Labels.
func (mr *MockGitHubMockRecorder) Labels(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Labels", reflect.TypeOf((*MockGitHub)(nil).Labels), arg0, arg1)
}
