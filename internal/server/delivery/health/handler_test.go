package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		want *Handler
		name string
	}{
		{
			name: "success/creates_handler",
			want: &Handler{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewHandler()
			require.NotNil(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_HealthCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		setupRequest   func() *http.Request
		name           string
		wantBody       string
		wantStatusCode int
	}{
		{
			name: "success/returns_200_ok",
			setupRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/health", nil)
				return req
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			gin.SetMode(gin.TestMode)
			handler := NewHandler()

			// Create gin router and register endpoint
			router := gin.New()
			router.GET("/health", handler.HealthCheck)

			// Create test request
			req := tt.setupRequest()
			recorder := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(recorder, req)

			// Assertions
			assert.Equal(t, tt.wantStatusCode, recorder.Code)
			if tt.wantBody != "" {
				assert.Contains(t, recorder.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestHandler_HealthCheck_WithServer(t *testing.T) {
	t.Parallel()

	// Test using httptest.Server as recommended in Issue #16
	gin.SetMode(gin.TestMode)
	handler := NewHandler()

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{
			name:           "success/health_check_endpoint",
			method:         http.MethodGet,
			path:           "/health",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Make HTTP request to test server
			req, err := http.NewRequest(tt.method, server.URL+tt.path, nil)
			require.NoError(t, err)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assertions
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}
