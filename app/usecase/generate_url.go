package usecase

import (
	"bucket-signer/app/domain"
	"context"
	"fmt"
	"strings"
)

// GenerateURLUseCase representa o caso de uso de geração de URL assinada
type GenerateURLUseCase struct {
	presigner      domain.Presigner
	allowedBuckets map[string]struct{}
	defaultBucket  string
}

// Option configura o caso de uso.
type Option func(*GenerateURLUseCase)

// WithAllowedBuckets restringe os buckets que podem ser assinados. Lista vazia = sem restrição.
func WithAllowedBuckets(buckets []string) Option {
	return func(uc *GenerateURLUseCase) {
		for _, b := range buckets {
			if b = strings.TrimSpace(b); b != "" {
				uc.allowedBuckets[b] = struct{}{}
			}
		}
	}
}

// WithDefaultBucket define o bucket usado quando o chamador não informa um.
func WithDefaultBucket(bucket string) Option {
	return func(uc *GenerateURLUseCase) { uc.defaultBucket = strings.TrimSpace(bucket) }
}

// NewGenerateURLUseCase cria uma nova instância do caso de uso
func NewGenerateURLUseCase(presigner domain.Presigner, opts ...Option) *GenerateURLUseCase {
	uc := &GenerateURLUseCase{presigner: presigner, allowedBuckets: map[string]struct{}{}}
	for _, o := range opts {
		o(uc)
	}
	return uc
}

// RestrictsBuckets indica se há allowlist de buckets configurada.
func (uc *GenerateURLUseCase) RestrictsBuckets() bool { return len(uc.allowedBuckets) > 0 }

// Execute executa o caso de uso para o bucket e cliente informado
func (uc *GenerateURLUseCase) Execute(ctx context.Context, bucketName, clienteID string, upload bool) (string, error) {
	clienteID = strings.TrimSpace(clienteID)
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		bucketName = uc.defaultBucket
	}
	if err := domain.ValidateClientID(clienteID); err != nil {
		return "", fmt.Errorf("clienteID: %w", err)
	}
	if err := domain.ValidateBucketName(bucketName); err != nil {
		return "", fmt.Errorf("bucket: %w", err)
	}
	if uc.RestrictsBuckets() {
		if _, ok := uc.allowedBuckets[bucketName]; !ok {
			return "", domain.ErrBucketNotAllowed
		}
	}
	return uc.presigner.GeneratePresignedURL(ctx, bucketName, clienteID, upload)
}
