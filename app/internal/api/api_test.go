package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, cfg Config, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	NewHandler(cfg).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestHealth(t *testing.T) {
	rec := get(t, Config{}, "/health")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestVersionReportsBuildAndEnvironment(t *testing.T) {
	rec := get(t, Config{Version: "abc1234", Environment: "dev"}, "/version")
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["version"] != "abc1234" || body["environment"] != "dev" {
		t.Fatalf("unexpected body %v", body)
	}
}

func TestVersionNeverReturnsTheSecret(t *testing.T) {
	rec := get(t, Config{APIKeyPresent: true}, "/version")
	if !strings.Contains(rec.Body.String(), `"api_key_configured":true`) {
		t.Fatalf("flag missing: %s", rec.Body)
	}
}

func TestItems(t *testing.T) {
	var got []item
	if err := json.Unmarshal(get(t, Config{}, "/api/items").Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("want 3 items, got %d", len(got))
	}
}

func TestMetricsExposed(t *testing.T) {
	if rec := get(t, Config{}, "/metrics"); rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestUnknownMethodRejected(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandler(Config{}).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/health", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got %d", rec.Code)
	}
}
