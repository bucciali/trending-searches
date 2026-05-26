package handlers

import (
	"trending-searches/pkg/response"

	"github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
	})
}
