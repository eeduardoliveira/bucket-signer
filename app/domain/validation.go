package domain

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidInput marks errors caused by caller-supplied input (mapped to HTTP 400).
var ErrInvalidInput = errors.New("entrada inválida")

// ErrBucketNotAllowed marks a request for a bucket outside the configured allowlist (HTTP 403).
var ErrBucketNotAllowed = errors.New("bucket não permitido")

// clienteID: starts alphanumeric, then alphanumerics and . _ @ - (no '/', '\', '%', spaces).
var clientIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._@-]{0,127}$`)

// S3 bucket naming rules (simplified): 3-63 chars, lowercase, digits, '.', '-'.
var bucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)

// ValidateClientID rejects values that could escape the tenant prefix in the object key
// (e.g. "../other", "a/b", encoded separators).
func ValidateClientID(id string) error {
	if !clientIDPattern.MatchString(id) || strings.Contains(id, "..") {
		return ErrInvalidInput
	}
	return nil
}

// ValidateBucketName enforces S3 bucket naming rules.
func ValidateBucketName(name string) error {
	if !bucketPattern.MatchString(name) || strings.Contains(name, "..") {
		return ErrInvalidInput
	}
	return nil
}
