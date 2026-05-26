package handlers

import (
	"trending-searches/internal/storageredis"
	"trending-searches/pkg/response"

	"github.com/gin-gonic/gin"
)

type AddStoplistRequest struct {
	Word string `json:"word" binding:"required"`
}

func AddToStoplistHandler(redisClient *storageredis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddStoplistRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "word is required")
			return
		}

		if err := redisClient.AddToStoplist(c.Request.Context(), req.Word); err != nil {
			response.InternalError(c, "failed to add word")
			return
		}

		response.Success(c, gin.H{
			"message": "word added to stoplist",
			"word":    req.Word,
		})
	}
}

func RemoveFromStoplistHandler(redisClient *storageredis.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		word := c.Param("word")
		if word == "" {
			response.BadRequest(c, "word is required")
			return
		}

		if err := redisClient.RemoveFromStoplist(c.Request.Context(), word); err != nil {
			response.InternalError(c, "failed to remove word")
			return
		}

		response.Success(c, gin.H{
			"message": "word removed from stoplist",
			"word":    word,
		})
	}
}
