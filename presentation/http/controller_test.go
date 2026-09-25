package http

import (
	"bucket-signer/app/usecase"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubPresigner struct {
	url string
	err error
	n   int
}

func (s *stubPresigner) GeneratePresignedURL(context.Context, string, string, bool) (string, error) {
	s.n++
	return s.url, s.err
}

func newMux(p *stubPresigner, token string, opts ...usecase.Option) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterRoutes(mux, &SignedURLController{UseCase: usecase.NewGenerateURLUseCase(p, opts...)}, token)
	return mux
}

func do(mux http.Handler, method, target string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestAuthRequiredWhenTokenConfigured(t *testing.T) {
	p := &stubPresigner{url: "https://s3/x"}
	mux := newMux(p, "s3cret")
	for _, h := range []map[string]string{nil, {"Authorization": "Bearer wrong"}, {"X-API-Key": "nope"}} {
		if rec := do(mux, "GET", "/signed-url?bucket=my-bucket&clienteID=c1", h); rec.Code != http.StatusUnauthorized {
			t.Fatalf("headers %v: want 401, got %d", h, rec.Code)
		}
	}
	if p.n != 0 {
		t.Fatal("presigner called without auth")
	}
	for _, h := range []map[string]string{{"Authorization": "Bearer s3cret"}, {"X-API-Key": "s3cret"}} {
		if rec := do(mux, "GET", "/signed-url?bucket=my-bucket&clienteID=c1", h); rec.Code != http.StatusOK {
			t.Fatalf("headers %v: want 200, got %d", h, rec.Code)
		}
	}
}

func TestLegacyModeWithoutToken(t *testing.T) {
	rec := do(newMux(&stubPresigner{url: "https://s3/x"}, ""), "GET", "/signed-url?bucket=my-bucket&clienteID=c1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestOnlyGETAllowed(t *testing.T) {
	rec := do(newMux(&stubPresigner{}, ""), "POST", "/signed-url?bucket=my-bucket&clienteID=c1", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}

func TestTraversalRejected(t *testing.T) {
	p := &stubPresigner{}
	rec := do(newMux(p, ""), "GET", "/signed-url?bucket=my-bucket&clienteID=..%2Fother", nil)
	if rec.Code != http.StatusBadRequest || p.n != 0 {
		t.Fatalf("want 400 and no presign, got %d (calls=%d)", rec.Code, p.n)
	}
}

func TestBucketAllowlist(t *testing.T) {
	rec := do(newMux(&stubPresigner{}, "", usecase.WithAllowedBuckets([]string{"allowed"})), "GET", "/signed-url?bucket=other-bucket&clienteID=c1", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", rec.Code)
	}
}

func TestResponseIsValidJSONAndErrorsAreNotLeaked(t *testing.T) {
	p := &stubPresigner{url: `https://s3/x?a=1&b="2"`}
	rec := do(newMux(p, ""), "GET", "/signed-url?bucket=my-bucket&clienteID=c1", nil)
	var body signedURLResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.URL != p.url {
		t.Fatalf("bad JSON %q: %v", rec.Body.String(), err)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing Cache-Control: no-store")
	}

	p = &stubPresigner{err: errors.New("AccessDenied: arn:aws:iam::123:user/secret-internal")}
	rec = do(newMux(p, ""), "GET", "/signed-url?bucket=my-bucket&clienteID=c1", nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
	if json.Valid(rec.Body.Bytes()) == false || strings.Contains(rec.Body.String(), "arn:aws") {
		t.Fatalf("internal error leaked or invalid JSON: %s", rec.Body.String())
	}
}
