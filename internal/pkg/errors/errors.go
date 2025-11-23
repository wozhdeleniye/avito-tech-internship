package serverrors

import (
	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
)

type ServiceError struct {
	HTTPCode int
	Code     openapi.ErrorResponseErrorCode
	Message  string
}

func (e *ServiceError) Error() string {
	return e.Message
}

var (
	ErrTeamExists = &ServiceError{HTTPCode: 400, Code: "TEAM_EXISTS", Message: "team_name already in use"}

	ErrUserNotFound = &ServiceError{HTTPCode: 404, Code: "USER_NOT_FOUND", Message: "user not found"}
	ErrUserExists   = &ServiceError{HTTPCode: 409, Code: "USER_EXISTS", Message: "user exists"}
)
var (
	ErrUnknown = &ServiceError{HTTPCode: 520, Code: "UNKNOWN_ERROR", Message: "unknown error"}
)
var (
	ErrTeamNotFound = &ServiceError{HTTPCode: 404, Code: "TEAM_NOT_FOUND", Message: "team not found"}
)
var (
	ErrPRNotFound = &ServiceError{HTTPCode: 404, Code: "PR_NOT_FOUND", Message: "pull request not found"}
)
var (
	ErrPRExists = &ServiceError{HTTPCode: 409, Code: "PR_EXISTS", Message: "PR id already exists"}
)
var (
	ErrNoAvailableReviewers = &ServiceError{Code: "NO_AVAILABLE_REVIEWERS", Message: "no available reviewers in the team"}
)
var (
	ErrUnauthorized = &ServiceError{Code: "UNAUTHORIZED", Message: "unauthorized access"}
)
var (
	ErrInvalidToken = &ServiceError{Code: "INVALID_TOKEN", Message: "invalid token"}
)

var (
	ErrPRMerged    = &ServiceError{HTTPCode: 409, Code: "PR_MERGED", Message: "cannot reassign on merged PR"}
	ErrNotAssigned = &ServiceError{HTTPCode: 409, Code: "NOT_ASSIGNED", Message: "reviewer is not assigned to this PR"}
	ErrNoCandidate = &ServiceError{HTTPCode: 409, Code: "NO_CANDIDATE", Message: "no active replacement candidate in team"}
)
