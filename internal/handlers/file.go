package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "UploadFile endpoint - to be implemented",
	})
}

func ServeFile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "ServeFile endpoint - to be implemented",
	})
}