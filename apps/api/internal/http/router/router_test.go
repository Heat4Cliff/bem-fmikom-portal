package router

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRouter() http.Handler {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return New(logger)
}

func TestHealthzIsOperationalAndReturnsCommonEnvelope(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	testRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var body struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || body.Data.Status != "ok" {
		t.Fatalf("unexpected health response: %+v", body)
	}
}

func TestUnknownRouteReturnsJSONNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	testRouter().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("404 response should be JSON: %v", err)
	}
	if body["success"] != false || body["message"] != "Resource not found" {
		t.Fatalf("unexpected not-found response: %#v", body)
	}
}

func TestRequestIDIsReturnedAndUnsafeIncomingIDIsReplaced(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Request-ID", "unsafe id")
	testRouter().ServeHTTP(recorder, request)

	requestID := recorder.Header().Get("X-Request-ID")
	if requestID == "" || requestID == "unsafe id" {
		t.Fatalf("request ID was not replaced safely: %q", requestID)
	}
}
