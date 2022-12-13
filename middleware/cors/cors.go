package cors

import (
	"github.com/ntt360/gin/utils/array"
	"net/http"

	"github.com/ntt360/gin"
)

// Cors 设置允许的跨域请求origin
// 如果允许所有来源跨域，设置 allowOrigins = ["*"]
func Cors(allowOrigins []string, aclHeaders map[string]string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.Request.Header.Get("Origin")
		if allowOrigins[0] == "*" || array.InEqual(origin, allowOrigins) {
			ctx.Header("Access-Control-Allow-Origin", origin)

			for headKey, headVal := range aclHeaders {
				ctx.Header(headKey, headVal)
			}
			/** 添加默认跨域返回头 */
			if _, ok := aclHeaders["Access-Control-Allow-Methods"]; !ok {
				ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			}
			if _, ok := aclHeaders["Access-Control-Allow-Headers"]; !ok {
				ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			}
			if _, ok := aclHeaders["Access-Control-Allow-Credentials"]; !ok {
				ctx.Header("Access-Control-Allow-Credentials", "true")
			}
			if _, ok := aclHeaders["Access-Control-Max-Age"]; !ok {
				ctx.Header("Access-Control-Max-Age", "3600")
			}

			/* 设置预检请求返回 */
			if ctx.Request.Method == "OPTIONS" {
				ctx.AbortWithStatus(http.StatusNoContent)
			}
		}
	}
}
