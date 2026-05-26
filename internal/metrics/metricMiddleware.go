package metrics

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("MetricsMiddleware called")
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		RequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), strconv.Itoa(status)).Inc()
		RequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}
