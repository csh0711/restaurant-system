package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostOrders(t *testing.T) {
	server := &Server{
		channel:   nil,
		queueName: "orders",
	}

	tests := map[string]struct {
		name           string
		body           string
		expectedStatus int
	}{
		"valid request": {
			body:           `{"tableId":"1","items":["Caesar Salad"]}`,
			expectedStatus: http.StatusCreated,
		},
		"missing tableId": {
			body:           `{"items":["Caesar Salad"]}`,
			expectedStatus: http.StatusBadRequest,
		},
		"invalid menu item": {
			body:           `{"tableId":"1","items":["Pizza Hawaii"]}`,
			expectedStatus: http.StatusBadRequest,
		},
		"missing menu items": {
			body:           `{"tableId":"1"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			server.PostOrders(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
