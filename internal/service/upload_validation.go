package service

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
)

const attachmentAllowedExtensionsSettingKey = "storage.allowed_attachment_extensions"

var (
	ErrAttachmentExtensionsNotConfigured = errors.New("未配置允许上传的附件扩展名")
	ErrAttachmentExtensionNotAllowed     = errors.New("附件扩展名不允许上传")
	ErrAttachmentMIMETypeNotAllowed      = errors.New("附件 MIME 类型不允许上传")
	ErrAvatarExtensionNotAllowed         = errors.New("头像扩展名不允许上传")
	ErrAvatarMIMETypeNotAllowed          = errors.New("头像 MIME 类型不允许上传")
)

var attachmentAllowedMIMETypes = map[string][]string{
	"gif":  {"image/gif"},
	"jpeg": {"image/jpeg"},
	"jpg":  {"image/jpeg"},
	"mp3":  {"audio/mpeg"},
	"mp4":  {"video/mp4"},
	"pdf":  {"application/pdf"},
	"png":  {"image/png"},
	"txt":  {"text/plain"},
	"webp": {"image/webp"},
	"zip":  {"application/x-zip-compressed", "application/zip"},
}

var avatarAllowedExtensions = []string{"jpg", "jpeg", "png", "webp"}

func validateAttachmentUpload(ctx context.Context, settings SettingService, originalFilename, declaredContentType string, content []byte) error {
	allowedExtensions, err := getAllowedAttachmentExtensions(ctx, settings)
	if err != nil {
		return err
	}

	return validateUpload(
		originalFilename,
		declaredContentType,
		content,
		allowedExtensions,
		attachmentAllowedMIMETypes,
		ErrAttachmentExtensionNotAllowed,
		ErrAttachmentMIMETypeNotAllowed,
	)
}

func validateAvatarUpload(originalFilename, declaredContentType string, content []byte) error {
	allowedExtensions := make([]string, 0, len(avatarAllowedExtensions))
	for _, ext := range avatarAllowedExtensions {
		allowedExtensions = append(allowedExtensions, ext)
	}

	return validateUpload(
		originalFilename,
		declaredContentType,
		content,
		allowedExtensions,
		attachmentAllowedMIMETypes,
		ErrAvatarExtensionNotAllowed,
		ErrAvatarMIMETypeNotAllowed,
	)
}

func validateUpload(originalFilename, declaredContentType string, content []byte, allowedExtensions []string, allowedMIMETypes map[string][]string, extensionErr error, mimeErr error) error {
	extension := normalizeFileExtension(originalFilename)
	if extension == "" || !slices.Contains(allowedExtensions, extension) {
		return extensionErr
	}

	allowedTypes := allowedMIMETypes[extension]
	if len(allowedTypes) == 0 {
		return mimeErr
	}

	declaredType := normalizeContentType(declaredContentType)
	if declaredType == "" || !slices.Contains(allowedTypes, declaredType) {
		return mimeErr
	}

	detectedType := normalizeContentType(http.DetectContentType(content))
	if detectedType == "" || !slices.Contains(allowedTypes, detectedType) {
		return mimeErr
	}

	return nil
}

func getAllowedAttachmentExtensions(ctx context.Context, settings SettingService) ([]string, error) {
	if settings == nil {
		return nil, ErrAttachmentExtensionsNotConfigured
	}

	value, found, err := settings.GetValue(ctx, attachmentAllowedExtensionsSettingKey)
	if err != nil {
		return nil, ErrInternal
	}
	if !found {
		return nil, ErrAttachmentExtensionsNotConfigured
	}

	extensions := parseAllowedAttachmentExtensions(value)
	if len(extensions) == 0 {
		return nil, ErrAttachmentExtensionsNotConfigured
	}

	return extensions, nil
}

func parseAllowedAttachmentExtensions(value string) []string {
	items := strings.Split(value, ",")
	if len(items) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(items))
	extensions := make([]string, 0, len(items))
	for _, item := range items {
		normalized := normalizeExtensionValue(item)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		extensions = append(extensions, normalized)
	}

	return extensions
}

func normalizeFileExtension(value string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(strings.TrimSpace(value))))
	return strings.TrimPrefix(ext, ".")
}

func normalizeExtensionValue(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.TrimPrefix(normalized, ".")
}

func normalizeContentType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	if idx := strings.Index(normalized, ";"); idx >= 0 {
		normalized = strings.TrimSpace(normalized[:idx])
	}
	return normalized
}
