package handlers

import (
	"strconv"

	"trending-searches/internal/storageredis"
	"trending-searches/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetTopHandler(redisClient *storageredis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}

		top, err := redisClient.GetTop(c.Request.Context(), limit)
		if err != nil {
			response.InternalError(c, "failed to get top")
			return
		}

		response.Success(c, gin.H{
			"top": top,
		})
	}
}
