package logger

import "github.com/ntt360/gin"

// GinxLogFormatterParams ginx custom web log formatter params struct.
type GinxLogFormatterParams struct {
	gin.LogFormatterParams
	Format   string // 数据输出格式：text，json
	Env      string
	Idc      string // 服务idc
	PrjName  string
	LocalIP  string
	Hostname string
}
