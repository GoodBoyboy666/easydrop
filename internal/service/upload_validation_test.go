package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestParseAllowedAttachmentExtensions(t *testing.T) {
	t.Parallel()

	got := parseAllowedAttachmentExtensions(" .PDF, png ,jpg,.jpg, ,WEBP ")
	want := []string{"pdf", "png", "jpg", "webp"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseAllowedAttachmentExtensions() = %#v, want %#v", got, want)
	}
}

func TestValidateAttachmentUploadRequiresConfiguredExtensions(t *testing.T) {
	t.Parallel()

	err := validateAttachmentUpload(context.Background(), &mockInitSettingService{values: map[string]string{}}, "report.pdf", "application/pdf", samplePDFData())
	if !errors.Is(err, ErrAttachmentExtensionsNotConfigured) {
		t.Fatalf("expected ErrAttachmentExtensionsNotConfigured, got %v", err)
	}
}

func TestValidateAttachmentUploadRejectsDisallowedExtension(t *testing.T) {
	t.Parallel()

	err := validateAttachmentUpload(context.Background(), &mockInitSettingService{
		values: map[string]string{attachmentAllowedExtensionsSettingKey: "pdf,png"},
	}, "evil.html", "text/html", []byte("<html></html>"))
	if !errors.Is(err, ErrAttachmentExtensionNotAllowed) {
		t.Fatalf("expected ErrAttachmentExtensionNotAllowed, got %v", err)
	}
}

func TestValidateAttachmentUploadRejectsMIMEMismatch(t *testing.T) {
	t.Parallel()

	err := validateAttachmentUpload(context.Background(), &mockInitSettingService{
		values: map[string]string{attachmentAllowedExtensionsSettingKey: "png"},
	}, "avatar.png", "image/png", samplePlainTextData())
	if !errors.Is(err, ErrAttachmentMIMETypeNotAllowed) {
		t.Fatalf("expected ErrAttachmentMIMETypeNotAllowed, got %v", err)
	}
}

func TestValidateAttachmentUploadAcceptsAllowedFile(t *testing.T) {
	t.Parallel()

	err := validateAttachmentUpload(context.Background(), &mockInitSettingService{
		values: map[string]string{attachmentAllowedExtensionsSettingKey: "pdf"},
	}, "report.PDF", "application/pdf", samplePDFData())
	if err != nil {
		t.Fatalf("validateAttachmentUpload returned error: %v", err)
	}
}

func TestValidateAvatarUploadAcceptsAllowedImage(t *testing.T) {
	t.Parallel()

	if err := validateAvatarUpload("avatar.JPG", "image/jpeg", sampleJPEGData()); err != nil {
		t.Fatalf("validateAvatarUpload returned error: %v", err)
	}
}

func TestValidateAvatarUploadRejectsDisallowedExtension(t *testing.T) {
	t.Parallel()

	err := validateAvatarUpload("avatar.gif", "image/gif", sampleGIFData())
	if !errors.Is(err, ErrAvatarExtensionNotAllowed) {
		t.Fatalf("expected ErrAvatarExtensionNotAllowed, got %v", err)
	}
}

func TestValidateAvatarUploadRejectsMIMEMismatch(t *testing.T) {
	t.Parallel()

	err := validateAvatarUpload("avatar.png", "image/png", samplePlainTextData())
	if !errors.Is(err, ErrAvatarMIMETypeNotAllowed) {
		t.Fatalf("expected ErrAvatarMIMETypeNotAllowed, got %v", err)
	}
}
