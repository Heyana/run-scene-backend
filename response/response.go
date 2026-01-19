package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseCode 响应状态码
type ResponseCode int

const (
	// 成功状态码
	CodeSuccess ResponseCode = 200

	// 客户端错误状态码 (400-499)
	CodeBadRequest          ResponseCode = 400
	CodeUnauthorized        ResponseCode = 401
	CodeForbidden           ResponseCode = 403
	CodeNotFound            ResponseCode = 404
	CodeMethodNotAllowed    ResponseCode = 405
	CodeConflict            ResponseCode = 409
	CodeUnprocessableEntity ResponseCode = 422
	CodeTooManyRequests     ResponseCode = 429

	// 服务器错误状态码 (500-599)
	CodeInternalServerError ResponseCode = 500
	CodeNotImplemented      ResponseCode = 501
	CodeBadGateway          ResponseCode = 502
	CodeServiceUnavailable  ResponseCode = 503
)

// Response 统一响应结构
type Response struct {
	Code      ResponseCode `json:"code"`      // 业务状态码
	Msg       string       `json:"msg"`       // 消息
	Data      interface{}  `json:"data"`      // 数据
	Timestamp int64        `json:"timestamp"` // 时间戳
}

// PaginationData 分页数据结构
type PaginationData struct {
	List     interface{} `json:"list"`      // 列表数据
	Total    int64       `json:"total"`     // 总数
	Page     int         `json:"page"`      // 当前页
	PageSize int         `json:"page_size"` // 每页数量
	Pages    int64       `json:"pages"`     // 总页数
}

// ErrorDetail 错误详情结构
type ErrorDetail struct {
	Code      ResponseCode      `json:"code"`      // 业务状态码
	Msg       string            `json:"msg"`       // 消息
	Details   map[string]string `json:"details"`   // 字段错误详情
	Timestamp int64             `json:"timestamp"` // 时间戳
}

// NewResponse 创建新的响应
func NewResponse(code ResponseCode, msg string, data interface{}) *Response {
	return &Response{
		Code:      code,
		Msg:       msg,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	response := NewResponse(CodeSuccess, "success", data)
	c.JSON(http.StatusOK, response)
}

// SuccessWithMsg 带消息的成功响应
func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	response := NewResponse(CodeSuccess, msg, data)
	c.JSON(http.StatusOK, response)
}

// Error 错误响应
func Error(c *gin.Context, code ResponseCode, msg string) {
	response := NewResponse(code, msg, nil)
	c.JSON(http.StatusOK, response) // 统一返回200状态码
}

// BadRequest 400错误
func BadRequest(c *gin.Context, msg string) {
	if msg == "" {
		msg = "请求参数错误"
	}
	Error(c, CodeBadRequest, msg)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, msg string) {
	if msg == "" {
		msg = "未认证的用户"
	}
	Error(c, CodeUnauthorized, msg)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, msg string) {
	if msg == "" {
		msg = "权限不足"
	}
	Error(c, CodeForbidden, msg)
}

// NotFound 404错误
func NotFound(c *gin.Context, msg string) {
	if msg == "" {
		msg = "资源不存在"
	}
	Error(c, CodeNotFound, msg)
}

// Conflict 409错误
func Conflict(c *gin.Context, msg string) {
	if msg == "" {
		msg = "资源冲突"
	}
	Error(c, CodeConflict, msg)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, msg string) {
	if msg == "" {
		msg = "系统错误"
	}
	Error(c, CodeInternalServerError, msg)
}

// ValidationError 422验证错误
func ValidationError(c *gin.Context, msg string) {
	if msg == "" {
		msg = "数据验证失败"
	}
	Error(c, CodeUnprocessableEntity, msg)
}

// SuccessWithPagination 分页响应
func SuccessWithPagination(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	pages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		pages++
	}
	data := PaginationData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}
	response := NewResponse(CodeSuccess, "success", data)
	c.JSON(http.StatusOK, response)
}

// ErrorWithDetails 带详情的错误响应（用于表单验证）
func ErrorWithDetails(c *gin.Context, code ResponseCode, msg string, details map[string]string) {
	response := &ErrorDetail{
		Code:      code,
		Msg:       msg,
		Details:   details,
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

// ValidationErrorWithDetails 带字段详情的验证错误
func ValidationErrorWithDetails(c *gin.Context, details map[string]string) {
	ErrorWithDetails(c, CodeUnprocessableEntity, "数据验证失败", details)
}

// TooManyRequests 429请求过于频繁
func TooManyRequests(c *gin.Context, msg string) {
	if msg == "" {
		msg = "请求过于频繁"
	}
	Error(c, CodeTooManyRequests, msg)
}

// 状态码对应的默认消息
var defaultMessages = map[ResponseCode]string{
	CodeSuccess:             "success",
	CodeBadRequest:          "请求参数错误",
	CodeUnauthorized:        "未认证的用户",
	CodeForbidden:           "权限不足",
	CodeNotFound:            "资源不存在",
	CodeMethodNotAllowed:    "方法不允许",
	CodeConflict:            "资源冲突",
	CodeUnprocessableEntity: "数据验证失败",
	CodeTooManyRequests:     "请求过于频繁",
	CodeInternalServerError: "系统错误",
	CodeNotImplemented:      "功能未实现",
	CodeBadGateway:          "网关错误",
	CodeServiceUnavailable:  "服务不可用",
}

// GetDefaultMessage 获取默认消息
func GetDefaultMessage(code ResponseCode) string {
	if msg, exists := defaultMessages[code]; exists {
		return msg
	}
	return "未知错误"
}
