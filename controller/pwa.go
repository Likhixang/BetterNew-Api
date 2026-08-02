package controller

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

// GetPwaManifest returns the PWA web app manifest with the live system name,
// so the installed app name follows the site branding — the same source
// (common.SystemName) that drives the page title at runtime.
// When a custom logo URL is configured (common.Logo), the icons point at it;
// otherwise the built-in icons generated from logo.png are used.
// The static manifest file must NOT exist in web/dist, otherwise the
// static.Serve middleware in SetWebRouter will shadow this route.
func GetPwaManifest(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	systemName := common.SystemName
	logo := common.Logo
	common.OptionMapRWMutex.RUnlock()

	if systemName == "" {
		systemName = "New API"
	}

	icons := []gin.H{
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
	}
	if logo != "" {
		// Custom logo: we cannot guarantee its dimensions or a maskable
		// safe zone, so declare it as "any" without a fixed type; the
		// browser fetches it and scales as needed.
		if strings.HasSuffix(strings.ToLower(logo), ".svg") {
			icons = []gin.H{
				{
					"src":     logo,
					"sizes":   "any",
					"type":    "image/svg+xml",
					"purpose": "any",
				},
			}
		} else {
			icons = []gin.H{
				{
					"src":     logo,
					"sizes":   "192x192",
					"purpose": "any",
				},
				{
					"src":     logo,
					"sizes":   "512x512",
					"purpose": "any",
				},
			}
		}
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
		"icons":            icons,
	})
}
