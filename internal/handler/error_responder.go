package handler

import (
	"errors"
	"net/http"

	"easydrop/internal/dto"
	"easydrop/internal/pkg/initsecret"
	"easydrop/internal/pkg/oauth"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// ErrorResponder 统一处理业务错误到 HTTP 响应的映射。
type ErrorResponder interface {
	Respond(c *gin.Context, err error)
}

type defaultErrorResponder struct{}

// NewErrorResponder 创建统一错误响应器。
func NewErrorResponder() ErrorResponder {
	return defaultErrorResponder{}
}

func ensureErrorResponder(errorResponder ErrorResponder) ErrorResponder {
	if errorResponder != nil {
		return errorResponder
	}
	return NewErrorResponder()
}

func (defaultErrorResponder) Respond(c *gin.Context, err error) {
	if err == nil {
		return
	}

	c.JSON(statusForError(err), dto.ErrorResponse{Message: err.Error()})
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, validator.ErrEmptyUsername),
		errors.Is(err, validator.ErrUsernameTooShort),
		errors.Is(err, validator.ErrUsernameTooLong),
		errors.Is(err, validator.ErrInvalidUsernameFormat),
		errors.Is(err, validator.ErrEmptyPassword),
		errors.Is(err, validator.ErrPasswordTooShort),
		errors.Is(err, validator.ErrPasswordContainsSpace),
		errors.Is(err, validator.ErrPasswordMissingLetter),
		errors.Is(err, validator.ErrPasswordMissingNumber),
		errors.Is(err, validator.ErrEmptyEmail),
		errors.Is(err, validator.ErrInvalidEmailFormat),
		errors.Is(err, service.ErrEmptyAccount),
		errors.Is(err, service.ErrCaptchaRequired),
		errors.Is(err, service.ErrCaptchaFailed),
		errors.Is(err, service.ErrInvalidSiteSetting),
		errors.Is(err, service.ErrInvalidPassword),
		errors.Is(err, service.ErrInvalidPasswordReset),
		errors.Is(err, service.ErrInvalidEmailVerify),
		errors.Is(err, service.ErrInvalidEmailChange),
		errors.Is(err, service.ErrEmptyAvatarContent),
		errors.Is(err, service.ErrEmptyAvatarFilename),
		errors.Is(err, service.ErrAvatarExtensionNotAllowed),
		errors.Is(err, service.ErrAvatarMIMETypeNotAllowed),
		errors.Is(err, service.ErrEmptyPostContent),
		errors.Is(err, service.ErrInvalidPostUser),
		errors.Is(err, service.ErrTagNameTooLong),
		errors.Is(err, service.ErrEmptyCommentContent),
		errors.Is(err, service.ErrInvalidCommentPost),
		errors.Is(err, service.ErrInvalidCommentUser),
		errors.Is(err, service.ErrInvalidCommentParent),
		errors.Is(err, service.ErrInvalidAttachmentBizType),
		errors.Is(err, service.ErrInvalidFileSize),
		errors.Is(err, service.ErrEmptyAttachmentContent),
		errors.Is(err, service.ErrAttachmentExtensionsNotConfigured),
		errors.Is(err, service.ErrAttachmentExtensionNotAllowed),
		errors.Is(err, service.ErrAttachmentMIMETypeNotAllowed),
		errors.Is(err, initsecret.ErrRequired),
		errors.Is(err, service.ErrSettingKeyRequired),
		errors.Is(err, service.ErrPasskeyLimitReached),
		errors.Is(err, service.ErrPasskeyNameEmpty),
		errors.Is(err, service.ErrPasskeyNameTooLong):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, service.ErrRegisterClosed),
		errors.Is(err, service.ErrUserDisabled),
		errors.Is(err, service.ErrEmailNotVerified),
		errors.Is(err, service.ErrStorageQuotaExceeded),
		errors.Is(err, service.ErrPostCommentDisabled),
		errors.Is(err, initsecret.ErrInvalid):
		return http.StatusForbidden
	case errors.Is(err, service.ErrUsernameExists),
		errors.Is(err, service.ErrEmailExists),
		errors.Is(err, service.ErrEmailChanged),
		errors.Is(err, service.ErrAlreadyInitialized):
		return http.StatusConflict
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrPostNotFound),
		errors.Is(err, service.ErrCommentNotFound),
		errors.Is(err, service.ErrAttachmentNotFound),
		errors.Is(err, service.ErrPasskeyNotFound),
		errors.Is(err, service.ErrOAuthBindNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrOAuthStateMismatch):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrOAuthEmailExistsUnbound):
		return http.StatusConflict
	case errors.Is(err, service.ErrOAuthBindAlreadyExists),
		errors.Is(err, service.ErrOAuthProviderBindAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, service.ErrOAuthNotConfigured):
		return http.StatusNotFound
	case errors.Is(err, oauth.ErrProviderDisabled),
		errors.Is(err, oauth.ErrProviderNotConfigured):
		return http.StatusNotFound
	case errors.Is(err, oauth.ErrTokenExchangeFailed),
		errors.Is(err, oauth.ErrFetchUserInfoFailed),
		errors.Is(err, oauth.ErrEmailNotReturned):
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
