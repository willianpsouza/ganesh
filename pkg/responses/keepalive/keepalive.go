package keepalive

import (
	"github.com/gin-gonic/gin"
	"time"
)

type KeepAlive struct {
	ClientIP         string `json:"client_ip"`
	ClientRemoteAddr string `json:"client_remote_addr"`
	ClientProtocol   string `json:"client_protocol"`
	ClientVersion    string `json:"client_version"`
	ClientID         string `json:"client_id"`
	Timestamp        int64  `json:"timestamp"`
	HeaderID         string `json:"header_id"`
	RequestID        string `json:"request_id"`
}

func GetKeepAlive(c *gin.Context) KeepAlive {
	RequestID := c.Param("id")
	return KeepAlive{
		ClientIP:         c.ClientIP(),
		ClientID:         c.GetHeader("User-Agent"),
		Timestamp:        time.Now().UnixMilli(),
		ClientProtocol:   "HTTP/1.1",
		ClientVersion:    c.Request.Proto,
		ClientRemoteAddr: c.Request.RemoteAddr,
		HeaderID:         c.Request.Header.Get("X-Request-Id"),
		RequestID:        RequestID,
	}
}
