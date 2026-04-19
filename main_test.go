package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetRate(t *testing.T) {
	createServer := func(handler http.HandlerFunc) (*httptest.Server, *ExchangeService) {
		server := httptest.NewServer(handler)
		service := NewExchangeService(server.URL)
		return server, service
	}

	t.Run("successful scenario", func(t *testing.T) {
		server, service := createServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(RateResponse{Base: "USD", Target: "EUR", Rate: 0.85})
		})
		defer server.Close()

		rate, err := service.GetRate("USD", "EUR")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rate != 0.85 {
			t.Errorf("expected rate 0.85, got %f", rate)
		}
	})

	t.Run("api business error", func(t *testing.T) {
		server, service := createServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(RateResponse{ErrorMsg: "invalid currency pair"})
		})
		defer server.Close()

		_, err := service.GetRate("USD", "XXX")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "api error: invalid currency pair" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		server, service := createServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("this is not json"))
		})
		defer server.Close()

		_, err := service.GetRate("USD", "EUR")
		if err == nil {
			t.Fatal("expected error for malformed json, got nil")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		server, service := createServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		defer server.Close()

		_, err := service.GetRate("USD", "EUR")
		if err == nil {
			t.Fatal("expected error for empty body, got nil")
		}
	})

	t.Run("server panic / 500", func(t *testing.T) {
		server, service := createServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer server.Close()

		_, err := service.GetRate("USD", "EUR")
		if err == nil {
			t.Fatal("expected error for 500 status, got nil")
		}
	})

	t.Run("slow response / timeout", func(t *testing.T) {
		server, service := createServer(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(6 * time.Second)
			w.WriteHeader(http.StatusOK)
		})
		defer server.Close()

		_, err := service.GetRate("USD", "EUR")
		if err == nil {
			t.Fatal("expected timeout error, got nil")
		}
	})
}
