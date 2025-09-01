package datasync

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "registers routes without panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			router := gin.New()
			group := router.Group("/test")

			handler := NewHandler(&mockSyncService{})

			// Should not panic
			assert.NotPanics(t, func() {
				RegisterRoutes(group, handler)
			})
		})
	}
}
