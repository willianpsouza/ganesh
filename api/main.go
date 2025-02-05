package api

import (
	"ganesh.provengo.io/pkg/responses/keepalive"
	"github.com/gin-gonic/gin"
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
