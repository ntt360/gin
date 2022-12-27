package servername

import (
	"net"
	"strings"

	"github.com/ntt360/gin"
	"github.com/ntt360/gin/core/config"
	"github.com/ntt360/gin/utils/array"
)

func ServerName(config *config.Base) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var host string
		var err error
		if !strings.Contains(ctx.Request.Host, ":") ||
			(strings.Contains(ctx.Request.Host, "[") && strings.Contains(ctx.Request.Host, "]") &&
				strings.LastIndexByte(ctx.Request.Host, ']') > strings.LastIndexByte(ctx.Request.Host, ':')) {
			host = ctx.Request.Host
		} else {
			host, _, err = net.SplitHostPort(ctx.Request.Host)
			if err != nil {
				ctx.AbortWithStatus(403)
			}
		}

		if config.Env == "prod" && !array.InEqual(host, config.ServerName) {
			ctx.AbortWithStatus(403)
		}

		ctx.Next()
	}
}
