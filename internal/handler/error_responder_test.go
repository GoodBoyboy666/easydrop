package handler

import (
	"errors"
	"net/http"
	"testing"

	"easydrop/internal/pkg/initsecret"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/service"
)

func TestStatusForError(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		status int
	}{
		{name: "validator bad request", err: validator.ErrEmptyUsername, status: http.StatusBadRequest},
		{name: "invalid credentials unauthorized", err: service.ErrInvalidCredentials, status: http.StatusUnauthorized},
		{name: "register closed forbidden", err: service.ErrRegisterClosed, status: http.StatusForbidden},
		{name: "email exists conflict", err: service.ErrEmailExists, status: http.StatusConflict},
		{name: "user not found", err: service.ErrUserNotFound, status: http.StatusNotFound},
		{name: "post not found", err: service.ErrPostNotFound, status: http.StatusNotFound},
		{name: "comment disabled", err: service.ErrPostCommentDisabled, status: http.StatusForbidden},
		{name: "attachment validation", err: service.ErrAttachmentExtensionNotAllowed, status: http.StatusBadRequest},
		{name: "init secret invalid", err: initsecret.ErrInvalid, status: http.StatusForbidden},
		{name: "already initialized", err: service.ErrAlreadyInitialized, status: http.StatusConflict},
		{name: "unknown error", err: errors.New("boom"), status: http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := statusForError(tc.err); got != tc.status {
				t.Fatalf("expected %d, got %d", tc.status, got)
			}
		})
	}
}
