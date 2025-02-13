package api

import (
	localStructs "ganesh.provengo.io/internal/structs"
	"ganesh.provengo.io/pkg/responses/keepalive"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "pong"})
	}
}

func KeepAlive() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, keepalive.GetKeepAlive(c))
	}
}

func Default() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	}
}

func SendUser(channel chan localStructs.DataLogin) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := localStructs.DataLogin{}
		err := c.BindJSON(&data)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			channel <- data
			c.JSON(200, gin.H{"status": "ok"})
		}
	}
}
