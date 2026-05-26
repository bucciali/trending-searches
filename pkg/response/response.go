package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorBody struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

func WriteJSON(c *gin.Context, status int, v any) {
	c.JSON(status, v)
}

func WriteError(c *gin.Context, status int, message string) {
	resp := ErrorResponse{
		Error: ErrorBody{
			Status:  status,
			Message: message,
		},
	}
	c.JSON(status, resp)
}

func BadRequest(c *gin.Context, message string) {
	WriteError(c, http.StatusBadRequest, message)
}

func NotFound(c *gin.Context, message string) {
	WriteError(c, http.StatusNotFound, message)
}

func InternalError(c *gin.Context, message string) {
	WriteError(c, http.StatusInternalServerError, message)
}
