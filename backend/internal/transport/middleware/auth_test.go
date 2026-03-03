package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthBearer_HappyPath_AllowsRequestWithValidToken(t *testing.T) {
	t.Parallel()

	expected := "secret-token"

	// next handler that should be reached
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := AuthBearer(expected, map[string]struct{}{})
	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "http://example.local/api/urls", nil)
	req.Header.Set("Authorization", "Bearer "+expected)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want %d. body=%q", rec.Code, http.StatusOK, rec.Body.String())
	}
}
