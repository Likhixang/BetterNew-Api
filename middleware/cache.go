package middleware

import (
	"github.com/gin-gonic/gin"
)

func Cache() func(c *gin.Context) {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// Never cache the service worker: browsers only check for SW updates
		// while the previous response is fresh, so a 7-day cache would leave
		// every deployment stale for a week. The PWA manifest is dynamic
		// (serves common.SystemName), so it must revalidate as well.
		if path == "/" || path == "/sw.js" || path == "/manifest.webmanifest" {
			c.Header("Cache-Control", "no-cache")
		} else {
			c.Header("Cache-Control", "max-age=604800") // one week
		}
		c.Header("Cache-Version", "b688f2fb5be447c25e5aa3bd063087a83db32a288bf6a4f35f2d8db310e40b14")
		c.Next()
	}
}
