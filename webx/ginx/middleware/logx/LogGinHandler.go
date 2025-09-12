package middleware

import (
	"bytes"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
)

type GinLogx struct {
	logx          logx.Loggerx
	allowReqBody  bool //  是否允许打印请求体
	allowRespBody bool // 是否允许打印响应体
}

// NewGinLogx 自定义Gin日志中间件
func NewGinLogx(logx logx.Loggerx) *GinLogx {
	return &GinLogx{logx: logx}
}

// AllowReqBody 允许打印请求体
func (l *GinLogx) AllowReqBody() *GinLogx {
	l.allowReqBody = true
	return l
}

// AllowRespBody 允许打印响应体
func (l *GinLogx) AllowRespBody() *GinLogx {
	l.allowRespBody = true
	return l
}

// ZerologLogger 自定义Gin日志中间件
//   - 【注意，中间件需在gin的注册中间件最后，否则可能会获取不到请求内容】
func (g *GinLogx) BuildGinHandlerLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		// 初始化AccessLog
		al := initAccessLog(c, startTime)

		//	允许打印请求体
		if g.allowReqBody {
			al.ReqBody = ""
			// 读取请求体
			if c.Request.Body != nil {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				al.ReqBody = string(bodyBytes)
				// 恢复请求体，以便后续处理
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		// 允许打印响应体
		if g.allowRespBody {
			// 使用自定义responseWriter
			writer := &responseWriter{
				ResponseWriter: c.Writer,
				al:             al,
			}
			c.Writer = writer
		}

		// 记录请求日志
		al.ReqLogPrint(g.logx)

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		al.StopTime = endTime.UnixMilli()
		al.Duration = endTime.Sub(startTime).Milliseconds() // 毫秒
		// 打印响应头
		al.RespHeader = c.Writer.Header()

		// 响应日志
		al.RespLogPrint(g.logx)
	}
}

func initAccessLog(c *gin.Context, startTime time.Time) *AccessLog {
	// 初始化AccessLog
	al := &AccessLog{
		StartTime: startTime.UnixMilli(),
		ClientIP:  c.ClientIP(),
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Query:     c.Request.URL.RawQuery,
		Proto:     c.Request.Proto,
		Headers:   c.Request.Header,
		ReqBody:   "Disable Display", // 默认不显示请求体和响应体
		RespBody:  "Disable Display", // 默认不显示请求体和响应体
	}
	if len(al.Path) > 1024 { // 判断 path 路径长度，如果过长，就截取。防止黑客伪造过大 的 path 路径，导致日志内容过大
		al.Path = al.Path[:1024]
	}
	// 请求id，用于追踪请求
	al.LogId = startTime.Format("20060102150405") + fmt.Sprintf("%d", startTime.Nanosecond())
	return al
}

type AccessLog struct {
	LogId      string      `json:"id"`        // 请求id
	Path       string      `json:"path"`      //  请求路径
	Method     string      `json:"method"`    //  请求方法
	Query      string      `json:"query"`     // 请求参数
	Proto      string      `json:"proto"`     // 请求协议
	Headers    http.Header `json:"headers"`   // 请求头
	ReqBody    string      `json:"req_body"`  //  请求体
	ClientIP   string      `json:"client_ip"` // 客户端IP
	Status     int         `json:"status"`
	RespBody   string      `json:"resp_body"`   //  响应体
	RespHeader http.Header `json:"resp_header"` // 响应头
	StartTime  int64       `json:"start_time"`
	StopTime   int64       `json:"stop_time"`
	Duration   int64       `json:"duration"` // 请求耗时
}

// responseWriter 自定义响应写入器，用于捕获响应状态码和响应体
//   - 响应体，因为gin的ctx没有暴漏响应体，但是暴漏了responseWriter，帮我们记录响应体
type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.al.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// ReqLogPrint 请求日志打印
func (a *AccessLog) ReqLogPrint(log logx.Loggerx) {
	// 创建日志事件
	if a.Status >= http.StatusBadRequest && a.Status < http.StatusInternalServerError {
		log.Warn("HTTP request", logx.String("log_id", a.LogId),
			logx.String("client_ip", a.ClientIP),
			logx.String("proto", a.Proto),
			logx.String("method", a.Method),
			logx.String("path", a.Path),
			logx.Any("headers", a.Headers),
			logx.String("req_body", a.ReqBody),
			logx.Int64("start_time", a.StartTime),
		)
	} else if a.Status >= http.StatusInternalServerError {
		log.Error("HTTP request", logx.String("log_id", a.LogId),
			logx.String("client_ip", a.ClientIP),
			logx.String("proto", a.Proto),
			logx.String("method", a.Method),
			logx.String("path", a.Path),
			logx.Any("headers", a.Headers),
			logx.String("req_body", a.ReqBody),
			logx.Int64("start_time", a.StartTime),
		)
	} else {
		log.Info("HTTP request, 日志解析状态码错误", logx.String("log_id", a.LogId),
			logx.String("client_ip", a.ClientIP),
			logx.String("proto", a.Proto),
			logx.String("method", a.Method),
			logx.String("path", a.Path),
			logx.Any("headers", a.Headers),
			logx.String("req_body", a.ReqBody),
			logx.Int64("start_time", a.StartTime),
		)
	}
}

// RespLogPrint 响应日志打印
func (a *AccessLog) RespLogPrint(log logx.Loggerx) {
	// 创建日志事件
	if a.Status >= http.StatusBadRequest && a.Status < http.StatusInternalServerError {
		log.Warn("HTTP response", logx.String("log_id", a.LogId),
			logx.Int("status", a.Status),
			logx.String("method", a.Method),
			logx.Any("headers", a.RespHeader),
			logx.String("resp_body", a.RespBody),
			logx.Int64("start_time", a.StartTime),
			logx.Int64("end_time", a.StopTime),
			logx.Int64("duration", a.Duration),
		)
	} else if a.Status >= http.StatusInternalServerError {
		log.Error("HTTP response", logx.String("log_id", a.LogId),
			logx.Int("status", a.Status),
			logx.String("method", a.Method),
			logx.Any("headers", a.RespHeader),
			logx.String("resp_body", a.RespBody),
			logx.Int64("start_time", a.StartTime),
			logx.Int64("end_time", a.StopTime),
			logx.Int64("duration", a.Duration),
		)
	} else {
		log.Info("HTTP response, 日志解析状态码错误", logx.String("log_id", a.LogId),
			logx.Int("status", a.Status),
			logx.String("method", a.Method),
			logx.Any("headers", a.RespHeader),
			logx.String("resp_body", a.RespBody),
			logx.Int64("start_time", a.StartTime),
			logx.Int64("end_time", a.StopTime),
			logx.Int64("duration", a.Duration),
		)
	}
}
