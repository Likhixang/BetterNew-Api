package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPwaManifestUsesSystemName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/manifest.webmanifest", GetPwaManifest)

	originalName := common.SystemName
	t.Cleanup(func() {
		common.SystemName = originalName
	})

	tests := []struct {
		name       string
		systemName string
		want       string
	}{
		{
			name:       "default name",
			systemName: "New API",
			want:       "New API",
		},
		{
			name:       "custom branding follows system name",
			systemName: "Acme AI Hub",
			want:       "Acme AI Hub",
		},
		{
			name:       "empty name falls back",
			systemName: "",
			want:       "New API",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.SystemName = tt.systemName

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/manifest.webmanifest", nil)
			engine.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			assert.Contains(
				t,
				w.Header().Get("Content-Type"),
				"application/manifest+json",
			)

			var body map[string]any
			require.NoError(t, common.Unmarshal(w.Body.Bytes(), &body))
			assert.Equal(t, tt.want, body["name"])
			assert.Equal(t, tt.want, body["short_name"])
			assert.Equal(t, "standalone", body["display"])
			assert.Equal(t, "/", body["start_url"])

			icons, ok := body["icons"].([]any)
			require.True(t, ok)
			assert.Len(t, icons, 3)
			assert.Equal(t, "/pwa-maskable-512x512.png",
				icons[2].(map[string]any)["src"])
		})
	}
}
