package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCorsMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Wrap it with CORS middleware
	handler := corsMiddleware(testHandler)

	// Test normal request
	t.Run("normal request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		// Check CORS headers
		if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
		}

		if methods := resp.Header.Get("Access-Control-Allow-Methods"); methods != "GET, POST, PUT, DELETE, OPTIONS" {
			t.Errorf("Expected Access-Control-Allow-Methods 'GET, POST, PUT, DELETE, OPTIONS', got '%s'", methods)
		}

		if headers := resp.Header.Get("Access-Control-Allow-Headers"); headers != "Content-Type, Authorization" {
			t.Errorf("Expected Access-Control-Allow-Headers 'Content-Type, Authorization', got '%s'", headers)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// Test OPTIONS request (preflight)
	t.Run("OPTIONS preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		// Check that OPTIONS returns OK without calling the handler
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// CORS headers should still be present
		if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
		}
	})

	// Test POST request
	t.Run("POST request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		// Check CORS headers are present
		if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// Test PUT request
	t.Run("PUT request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// Test DELETE request
	t.Run("DELETE request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}
