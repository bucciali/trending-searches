package api

import (
	"trending-searches/internal/handlers"
	"trending-searches/internal/metrics"
	"trending-searches/internal/storageredis"
	"trending-searches/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(redisClient *storageredis.RedisClient) *gin.Engine {
	r := gin.Default()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(metrics.MetricsMiddleware())
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	v1 := r.Group("/api/v1")
	{
		v1.GET("/top", handlers.GetTopHandler(redisClient))
		v1.POST("/stoplist", handlers.AddToStoplistHandler(redisClient))
		v1.DELETE("/stoplist/:word", handlers.RemoveFromStoplistHandler(redisClient))
	}
	return r
}
