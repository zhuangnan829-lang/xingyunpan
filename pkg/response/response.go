// 路径: pkg/response/response.go
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 业务状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// PageData 分页数据结构
type PageData struct {
	List       interface{} `json:"list"`        // 数据列表
	Total      int64       `json:"total"`       // 总记录数
	Page       int         `json:"page"`        // 当前页码
	PageSize   int         `json:"page_size"`   // 每页大小
	TotalPages int         `json:"total_pages"` // 总页数
}

// 业务状态码定义
const (
	CodeSuccess            = 200   // 成功
	CodeInvalidParams      = 400   // 请求参数错误
	CodeUnauthorized       = 401   // 未授权
	CodeForbidden          = 403   // 禁止访问
	CodeNotFound           = 404   // 资源不存在
	CodeInternalError      = 500   // 服务器内部错误
	CodeDatabaseError      = 501   // 数据库错误
	CodeCacheError         = 502   // 缓存错误
	CodeStorageError       = 503   // 存储错误
	CodeUserNotFound       = 1001  // 用户不存在
	CodeUserAlreadyExists  = 1002  // 用户已存在
	CodePasswordIncorrect  = 1003  // 密码错误
	CodeTokenInvalid       = 1004  // Token 无效
	CodeTokenExpired       = 1005  // Token 过期
	CodeFileNotFound       = 2001  // 文件不存在
	CodeFileAlreadyExists  = 2002  // 文件已存在
	CodeInsufficientSpace  = 2003  // 空间不足
	CodeUploadFailed       = 2004  // 上传失败
	CodeDownloadFailed     = 2005  // 下载失败
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "操作成功",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithData 错误响应（带数据）
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// InvalidParams 参数错误响应
func InvalidParams(c *gin.Context, message string) {
	Error(c, CodeInvalidParams, message)
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
		Data:    nil,
	})
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
		Data:    nil,
	})
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
		Data:    nil,
	})
}

// InternalError 服务器内部错误响应
func InternalError(c *gin.Context, message string) {
	Error(c, CodeInternalError, message)
}

// PageSuccess 分页成功响应
func PageSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "操作成功",
		Data: PageData{
			List:       list,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}
