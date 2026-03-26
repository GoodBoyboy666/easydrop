package handler

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

const uploadSniffSize = 512

func openUploadFile(fileHeader *multipart.FileHeader) (multipart.File, []byte, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, nil, "", err
	}

	sample := make([]byte, uploadSniffSize)
	n, err := file.Read(sample)
	if err != nil && !errors.Is(err, io.EOF) {
		_ = file.Close()
		return nil, nil, "", err
	}
	sample = sample[:n]

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		_ = file.Close()
		return nil, nil, "", err
	}

	contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if contentType == "" && len(sample) > 0 {
		contentType = http.DetectContentType(sample)
	}

	return file, sample, contentType, nil
}

func isRequestTooLargeError(err error) bool {
	if err == nil {
		return false
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return true
	}

	return strings.Contains(strings.ToLower(err.Error()), "request body too large")
}
