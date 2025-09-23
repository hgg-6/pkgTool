package prometheusx

import "github.com/gin-gonic/gin"

// PrometheusGinBuilder 接口
type PrometheusGinBuilder interface {
	// BuildResponseTime 构建响应时间
	BuildResponseTime() gin.HandlerFunc

	// BuildActiveRequest 构建活跃请求
	BuildActiveRequest() gin.HandlerFunc
}
