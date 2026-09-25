package usecase

import (
	"bucket-signer/app/domain"
	"context"
	"errors"
	"testing"
)

type fakePresigner struct {
	calls  int
	bucket string
	id     string
}

func (f *fakePresigner) GeneratePresignedURL(_ context.Context, bucket, id string, _ bool) (string, error) {
	f.calls++
	f.bucket, f.id = bucket, id
	return "https://signed", nil
}

func TestExecuteRejectsTraversalAndOddClientIDs(t *testing.T) {
	for _, id := range []string{"../other", "a/b", "..", "a..b", "%2e%2e", "a\\b", " ", "", ".hidden", "a b"} {
		f := &fakePresigner{}
		_, err := NewGenerateURLUseCase(f).Execute(context.Background(), "my-bucket", id, false)
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("clienteID %q: want ErrInvalidInput, got %v", id, err)
		}
		if f.calls != 0 {
			t.Errorf("clienteID %q: presigner must not be called", id)
		}
	}
}

func TestExecuteAcceptsNormalClientIDs(t *testing.T) {
	for _, id := range []string{"123", "cliente-a", "c_1.v2", "user@example.com", "550e8400-e29b-41d4-a716-446655440000"} {
		f := &fakePresigner{}
		if _, err := NewGenerateURLUseCase(f).Execute(context.Background(), "my-bucket", id, true); err != nil {
			t.Errorf("clienteID %q: unexpected error %v", id, err)
		}
	}
}

func TestExecuteRejectsInvalidBucketNames(t *testing.T) {
	for _, b := range []string{"UPPER", "a", "bucket/../x", "bucket?x=1", "-bad"} {
		_, err := NewGenerateURLUseCase(&fakePresigner{}).Execute(context.Background(), b, "c1", false)
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("bucket %q: want ErrInvalidInput, got %v", b, err)
		}
	}
}

func TestExecuteEnforcesBucketAllowlist(t *testing.T) {
	f := &fakePresigner{}
	uc := NewGenerateURLUseCase(f, WithAllowedBuckets([]string{"chatbot-prompt", " "}))
	if _, err := uc.Execute(context.Background(), "someone-elses-bucket", "c1", false); !errors.Is(err, domain.ErrBucketNotAllowed) {
		t.Fatalf("want ErrBucketNotAllowed, got %v", err)
	}
	if _, err := uc.Execute(context.Background(), "chatbot-prompt", "c1", false); err != nil {
		t.Fatalf("allowed bucket rejected: %v", err)
	}
}

func TestExecuteFallsBackToDefaultBucket(t *testing.T) {
	f := &fakePresigner{}
	uc := NewGenerateURLUseCase(f, WithDefaultBucket("chatbot-prompt"), WithAllowedBuckets([]string{"chatbot-prompt"}))
	if _, err := uc.Execute(context.Background(), "", "c1", false); err != nil {
		t.Fatal(err)
	}
	if f.bucket != "chatbot-prompt" {
		t.Fatalf("want default bucket, got %q", f.bucket)
	}
}
