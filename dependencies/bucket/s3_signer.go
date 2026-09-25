package bucket

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	// DefaultExpiration é usada quando EXPIRATION_MINUTES está ausente ou inválida.
	DefaultExpiration = 10 * time.Minute
	// MaxExpiration limita a validade de qualquer URL assinada.
	MaxExpiration = 60 * time.Minute
	// DefaultKeyPattern é o layout do objeto: clienteID/clienteID-prompt.json
	DefaultKeyPattern = "%s/%s-prompt.json"
)

type S3Presigner struct {
	client     *s3.Client
	expiration time.Duration
	keyPattern string
}

func NewS3Presigner() (*S3Presigner, error) {
	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("variáveis AWS_ACCESS_KEY_ID ou AWS_SECRET_ACCESS_KEY não definidas")
	}

	pattern, err := ParseKeyPattern(os.Getenv("PROMPT_FILE_PATTERN"))
	if err != nil {
		return nil, err
	}

	// Carrega configuração com credenciais estáticas para AWS S3
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar configuração AWS: %w", err)
	}

	return &S3Presigner{
		client:     s3.NewFromConfig(cfg),
		expiration: ParseExpiration(os.Getenv("EXPIRATION_MINUTES")),
		keyPattern: pattern,
	}, nil
}

// ParseExpiration converte EXPIRATION_MINUTES em duração, aplicando default e teto.
func ParseExpiration(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultExpiration
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		log.Printf("⚠️  EXPIRATION_MINUTES inválido (%q); usando %s", raw, DefaultExpiration)
		return DefaultExpiration
	}
	d := time.Duration(n) * time.Minute
	if d > MaxExpiration {
		log.Printf("⚠️  EXPIRATION_MINUTES=%d excede o máximo; limitado a %s", n, MaxExpiration)
		return MaxExpiration
	}
	return d
}

// ParseKeyPattern valida PROMPT_FILE_PATTERN: apenas verbos %s (1 ou 2), sem "..".
func ParseKeyPattern(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return DefaultKeyPattern, nil
	}
	n := strings.Count(raw, "%s")
	if n < 1 || n > 2 || strings.Count(raw, "%") != n || strings.Contains(raw, "..") || strings.HasPrefix(raw, "/") {
		return "", fmt.Errorf("PROMPT_FILE_PATTERN inválido: use 1 ou 2 ocorrências de %%s e nenhum outro verbo")
	}
	return raw, nil
}

// BuildKey monta a chave do objeto para o clienteID (já validado pelo caso de uso).
func BuildKey(pattern, clienteID string) string {
	args := make([]any, strings.Count(pattern, "%s"))
	for i := range args {
		args[i] = clienteID
	}
	return fmt.Sprintf(pattern, args...)
}

// GeneratePresignedURL gera a URL assinada com base no clienteID
func (p *S3Presigner) GeneratePresignedURL(ctx context.Context, bucketName, clienteID string, upload bool) (string, error) {
	key := BuildKey(p.keyPattern, strings.TrimSpace(clienteID))
	presigner := s3.NewPresignClient(p.client)

	if upload {
		req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(p.expiration))
		if err != nil {
			return "", fmt.Errorf("erro ao gerar presigned URL PUT: %w", err)
		}
		log.Printf("presign PUT bucket=%s key=%s exp=%s", bucketName, key, p.expiration)
		return req.URL, nil
	}

	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(p.expiration))
	if err != nil {
		return "", fmt.Errorf("erro ao gerar presigned URL: %w", err)
	}
	log.Printf("presign GET bucket=%s key=%s exp=%s", bucketName, key, p.expiration)
	return req.URL, nil
}
