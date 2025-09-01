package about

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBuildInfoOperator implements BuildInfoOperator interface for testing.
type MockBuildInfoOperator struct {
	VersionFunc func() string
	DateFunc    func() time.Time
	CommitFunc  func() string
}

func (m *MockBuildInfoOperator) Version() string {
	if m.VersionFunc != nil {
		return m.VersionFunc()
	}
	return "1.0.0"
}

func (m *MockBuildInfoOperator) Date() time.Time {
	if m.DateFunc != nil {
		return m.DateFunc()
	}
	return time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC)
}

func (m *MockBuildInfoOperator) Commit() string {
	if m.CommitFunc != nil {
		return m.CommitFunc()
	}
	return "abc123def"
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		info BuildInfoOperator
		want *Handler
		name string
	}{
		{
			name: "success/creates_handler_with_build_info",
			info: &MockBuildInfoOperator{},
			want: &Handler{info: &MockBuildInfoOperator{}},
		},
		{
			name: "success/nil_build_info",
			info: nil,
			want: &Handler{info: nil},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewHandler(tt.info)
			require.NotNil(t, got)
			assert.Equal(t, tt.info, got.info)
		})
	}
}

func TestHandler_AboutInfo(t *testing.T) {
	t.Parallel()

	testDate := time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		setupMock      func(*MockBuildInfoOperator)
		wantResponse   BuildInfo
		name           string
		wantStatusCode int
	}{
		{
			name: "success/returns_build_info",
			setupMock: func(m *MockBuildInfoOperator) {
				m.VersionFunc = func() string { return "1.2.3" }
				m.DateFunc = func() time.Time { return testDate }
				m.CommitFunc = func() string { return "abc123def456" }
			},
			wantStatusCode: http.StatusOK,
			wantResponse: BuildInfo{
				Version: "1.2.3",
				Date:    testDate,
				Commit:  "abc123def456",
			},
		},
		{
			name: "success/empty_build_info",
			setupMock: func(m *MockBuildInfoOperator) {
				m.VersionFunc = func() string { return "" }
				m.DateFunc = func() time.Time { return time.Time{} }
				m.CommitFunc = func() string { return "" }
			},
			wantStatusCode: http.StatusOK,
			wantResponse: BuildInfo{
				Version: "",
				Date:    time.Time{},
				Commit:  "",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			gin.SetMode(gin.TestMode)
			mockInfo := &MockBuildInfoOperator{}
			if tt.setupMock != nil {
				tt.setupMock(mockInfo)
			}

			handler := NewHandler(mockInfo)

			// Create gin router and register endpoint
			router := gin.New()
			router.GET("/about", handler.AboutInfo)

			// Create test request
			req, err := http.NewRequest(http.MethodGet, "/about", nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(recorder, req)

			// Assertions
			assert.Equal(t, tt.wantStatusCode, recorder.Code)

			var response BuildInfo
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.wantResponse, response)
		})
	}
}

func TestHandler_AboutInfo_WithServer(t *testing.T) {
	t.Parallel()

	// Test using httptest.Server as recommended in Issue #16
	gin.SetMode(gin.TestMode)
	testDate := time.Date(2023, 12, 1, 10, 0, 0, 0, time.UTC)

	mockInfo := &MockBuildInfoOperator{
		VersionFunc: func() string { return "v1.0.0" },
		DateFunc:    func() time.Time { return testDate },
		CommitFunc:  func() string { return "commit123" },
	}

	handler := NewHandler(mockInfo)
	router := gin.New()
	router.GET("/about", handler.AboutInfo)

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	tests := []struct {
		name           string
		method         string
		path           string
		wantVersion    string
		wantCommit     string
		wantStatusCode int
	}{
		{
			name:           "success/about_info_endpoint",
			method:         http.MethodGet,
			path:           "/about",
			wantStatusCode: http.StatusOK,
			wantVersion:    "v1.0.0",
			wantCommit:     "commit123",
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

			// Parse response
			var response BuildInfo
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, tt.wantVersion, response.Version)
			assert.Equal(t, tt.wantCommit, response.Commit)
			assert.Equal(t, testDate, response.Date)
		})
	}
}

func TestHandler_AboutInfo_NilBuildInfo(t *testing.T) {
	t.Parallel()

	// Test behavior when BuildInfoOperator is nil
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil)

	router := gin.New()
	router.GET("/about", func(c *gin.Context) {
		// This will panic if handler tries to call methods on nil info
		defer func() {
			if r := recover(); r != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "nil build info"})
			}
		}()
		handler.AboutInfo(c)
	})

	req, err := http.NewRequest(http.MethodGet, "/about", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return 500 due to nil pointer
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
