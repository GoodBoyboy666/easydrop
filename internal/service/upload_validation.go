package service

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"easydrop/internal/consts"
)

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

// validateAttachmentUpload 按配置的扩展名与 MIME 规则校验附件上传请求。
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

// validateAvatarUpload 按头像固定规则校验头像上传请求。
func validateAvatarUpload(originalFilename, declaredContentType string, content []byte) error {
	allowedExtensions := make([]string, 0, len(avatarAllowedExtensions))
	allowedExtensions = append(allowedExtensions, avatarAllowedExtensions...)

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

// validateUpload 执行统一上传校验流程：扩展名、声明 MIME、内容探测 MIME。
func validateUpload(originalFilename, declaredContentType string, content []byte, allowedExtensions []string, allowedMIMETypes map[string][]string, extensionErr error, mimeErr error) error {
	// 先校验扩展名是否被允许。
	extension := normalizeFileExtension(originalFilename)
	if extension == "" || !slices.Contains(allowedExtensions, extension) {
		return extensionErr
	}

	// 再校验声明的 MIME 类型是否与扩展名匹配。
	allowedTypes := allowedMIMETypes[extension]
	if len(allowedTypes) == 0 {
		return mimeErr
	}

	declaredType := normalizeContentType(declaredContentType)
	if declaredType == "" || !slices.Contains(allowedTypes, declaredType) {
		return mimeErr
	}

	// 最后基于内容探测 MIME，防止仅伪造 Content-Type。
	detectedType := normalizeContentType(http.DetectContentType(content))
	if detectedType == "" || !slices.Contains(allowedTypes, detectedType) {
		return mimeErr
	}

	return nil
}

// getAllowedAttachmentExtensions 从动态配置读取并解析允许的附件扩展名。
func getAllowedAttachmentExtensions(ctx context.Context, settings SettingService) ([]string, error) {
	if settings == nil {
		return nil, ErrAttachmentExtensionsNotConfigured
	}

	value, found, err := settings.GetValue(ctx, consts.AttachmentAllowedExtensionsSettingKey)
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

// parseAllowedAttachmentExtensions 解析逗号分隔扩展名并做去重清洗。
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

// normalizeFileExtension 提取文件扩展名并转换为不带点的小写形式。
func normalizeFileExtension(value string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(strings.TrimSpace(value))))
	return strings.TrimPrefix(ext, ".")
}

// normalizeExtensionValue 规范化配置中的扩展名值。
func normalizeExtensionValue(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.TrimPrefix(normalized, ".")
}

// normalizeContentType 规范化 MIME 字符串并去除参数部分。
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
