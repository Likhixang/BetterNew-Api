package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

// GetPwaManifest returns the PWA web app manifest with the live system name,
// so the installed app name follows the site branding — the same source
// (common.SystemName) that drives the page title at runtime.
// The static manifest file must NOT exist in web/dist, otherwise the
// static.Serve middleware in SetWebRouter will shadow this route.
func GetPwaManifest(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	systemName := common.SystemName
	common.OptionMapRWMutex.RUnlock()

	if systemName == "" {
		systemName = "New API"
	}

	c.Header("Content-Type", "application/manifest+json; charset=utf-8")
	c.JSON(http.StatusOK, gin.H{
		"name":             systemName,
		"short_name":       systemName,
		"description":      "Unified AI API gateway and admin dashboard.",
		"lang":             "zh-CN",
		"start_url":        "/",
		"scope":            "/",
		"display":          "standalone",
		"background_color": "#ffffff",
		"theme_color":      "#ffffff",
		"icons": []gin.H{
			{
				"src":     "/pwa-192x192.png",
				"sizes":   "192x192",
				"type":    "image/png",
				"purpose": "any",
			},
			{
				"src":     "/pwa-512x512.png",
				"sizes":   "512x512",
				"type":    "image/png",
				"purpose": "any",
			},
			{
				"src":     "/pwa-maskable-512x512.png",
				"sizes":   "512x512",
				"type":    "image/png",
				"purpose": "maskable",
			},
		},
	})
}
