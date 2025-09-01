package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCtxExtractor(t *testing.T) {
	t.Parallel()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	extractor := NewCtxExtractor(c)

	require.NotNil(t, extractor)
	assert.Equal(t, c, extractor.c)
}

func TestCtxExtractor_UserID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	type contextSetup struct {
		value interface{}
		key   string
	}
	tests := []struct {
		setup   *contextSetup
		name    string
		want    uuid.UUID
		wantErr bool
	}{
		{
			name: "valid_user_id",
			setup: &contextSetup{
				key:   consts.CtxKeyUserID,
				value: userID,
			},
			want:    userID,
			wantErr: false,
		},
		{
			name: "nil_user_id",
			setup: &contextSetup{
				key:   consts.CtxKeyUserID,
				value: uuid.Nil,
			},
			want:    uuid.Nil,
			wantErr: false,
		},
		{
			name:    "user_id_not_found",
			setup:   nil,
			want:    uuid.Nil,
			wantErr: true,
		},
		{
			name: "user_id_wrong_type_string",
			setup: &contextSetup{
				key:   consts.CtxKeyUserID,
				value: "not-a-uuid",
			},
			want:    uuid.Nil,
			wantErr: true,
		},
		{
			name: "user_id_wrong_type_int",
			setup: &contextSetup{
				key:   consts.CtxKeyUserID,
				value: 12345,
			},
			want:    uuid.Nil,
			wantErr: true,
		},
		{
			name: "user_id_wrong_type_nil",
			setup: &contextSetup{
				key:   consts.CtxKeyUserID,
				value: nil,
			},
			want:    uuid.Nil,
			wantErr: true,
		},
		{
			name: "different_key",
			setup: &contextSetup{
				key:   "differentKey",
				value: userID,
			},
			want:    uuid.Nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tt.setup != nil {
				c.Set(tt.setup.key, tt.setup.value)
			}

			extractor := NewCtxExtractor(c)
			got, err := extractor.UserID()

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, uuid.Nil, got)
				switch {
				case tt.setup == nil:
					assert.Contains(t, err.Error(), "user ID not found in context")
				case tt.setup.key != consts.CtxKeyUserID:
					assert.Contains(t, err.Error(), "user ID not found in context")
				default:
					assert.Contains(t, err.Error(), "user ID in context is not a uuid.UUID")
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCtxExtractor_BindJSON(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	validJSON := testStruct{Name: "test", Value: 42}
	validJSONBytes, _ := json.Marshal(validJSON)

	tests := []struct {
		dest     interface{}
		want     interface{}
		name     string
		jsonBody string
		wantErr  bool
	}{
		{
			name:     "valid_json_binding",
			jsonBody: string(validJSONBytes),
			dest:     &testStruct{},
			want:     &validJSON,
			wantErr:  false,
		},
		{
			name:     "empty_json_object",
			jsonBody: "{}",
			dest:     &testStruct{},
			want:     &testStruct{},
			wantErr:  false,
		},
		{
			name:     "invalid_json_malformed",
			jsonBody: `{"name": "test", "value":}`,
			dest:     &testStruct{},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid_json_wrong_type",
			jsonBody: `{"name": "test", "value": "not-a-number"}`,
			dest:     &testStruct{},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "empty_body",
			jsonBody: "",
			dest:     &testStruct{},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "null_json",
			jsonBody: "null",
			dest:     &testStruct{},
			want:     &testStruct{},
			wantErr:  false,
		},
		{
			name:     "array_instead_of_object",
			jsonBody: `[{"name": "test"}]`,
			dest:     &testStruct{},
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tt.jsonBody))
			req.Header.Set("Content-Type", "application/json")

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = req

			extractor := NewCtxExtractor(c)
			err := extractor.BindJSON(tt.dest)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to bind JSON")
			} else {
				require.NoError(t, err)
				if tt.want != nil {
					assert.Equal(t, tt.want, tt.dest)
				}
			}
		})
	}
}

func TestCtxExtractor_BindURI(t *testing.T) {
	t.Parallel()

	type uriStruct struct {
		ID   string `uri:"id"`
		Name string `uri:"name"`
	}

	tests := []struct {
		dest      interface{}
		want      interface{}
		uriParams map[string]string
		name      string
		urlPath   string
		wantErr   bool
	}{
		{
			name:    "valid_uri_binding",
			urlPath: "/test/123/item/testname",
			uriParams: map[string]string{
				"id":   "123",
				"name": "testname",
			},
			dest:    &uriStruct{},
			want:    &uriStruct{ID: "123", Name: "testname"},
			wantErr: false,
		},
		{
			name:    "partial_uri_binding",
			urlPath: "/test/456",
			uriParams: map[string]string{
				"id": "456",
			},
			dest:    &uriStruct{},
			want:    &uriStruct{ID: "456", Name: ""},
			wantErr: false,
		},
		{
			name:      "no_params",
			urlPath:   "/test",
			uriParams: map[string]string{},
			dest:      &uriStruct{},
			want:      &uriStruct{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.urlPath, nil)

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = req

			// Set URI parameters in Gin context
			for key, value := range tt.uriParams {
				c.Params = append(c.Params, gin.Param{Key: key, Value: value})
			}

			extractor := NewCtxExtractor(c)
			err := extractor.BindURI(tt.dest)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to bind URI")
			} else {
				require.NoError(t, err)
				if tt.want != nil {
					assert.Equal(t, tt.want, tt.dest)
				}
			}
		})
	}
}

func TestCtxExtractor_Integration(t *testing.T) {
	t.Parallel()

	// Integration test combining multiple operations
	userID := uuid.New()

	type requestBody struct {
		Data string `json:"data"`
	}

	type uriParams struct {
		ItemID string `uri:"item_id"`
	}

	reqBody := requestBody{Data: "test data"}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/items/123", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	c.Set(consts.CtxKeyUserID, userID)
	c.Params = append(c.Params, gin.Param{Key: "item_id", Value: "123"})

	extractor := NewCtxExtractor(c)

	// Test UserID extraction
	gotUserID, err := extractor.UserID()
	require.NoError(t, err)
	assert.Equal(t, userID, gotUserID)

	// Test JSON binding
	var gotBody requestBody
	err = extractor.BindJSON(&gotBody)
	require.NoError(t, err)
	assert.Equal(t, reqBody, gotBody)

	// Test URI binding
	var gotURI uriParams
	err = extractor.BindURI(&gotURI)
	require.NoError(t, err)
	assert.Equal(t, uriParams{ItemID: "123"}, gotURI)
}
