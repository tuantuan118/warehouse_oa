package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 返回消息
	Data    interface{} `json:"data"`    // 返回的数据
}

type PaginatedResponse struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 返回消息
	Data    interface{} `json:"data"`    // 返回的数据
	Total   int64       `json:"total"`   // 数据数量
}

// 定义常见的状态码
const (
	SuccessCode    = 200
	ErrorCode      = 500
	BadRequestCode = 400
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode,
		Message: "Success",
		Data:    data,
	})
}

// SuccessWithMessage 自定义消息的成功响应
//func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
//	c.JSON(http.StatusOK, Response{
//		Code:    SuccessCode,
//		Message: message,
//		Data:    data,
//	})
//}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// InternalServerError 通用的服务器内部错误响应
func InternalServerError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    ErrorCode,
		Message: err.Error(),
		Data:    nil,
	})
}

// InternalServerErrorData 通用的服务器内部错误响应
func InternalServerErrorData(c *gin.Context, err error, data interface{}) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    ErrorCode,
		Message: err.Error(),
		Data:    data,
	})
}

// BadRequest 参数错误响应
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    BadRequestCode,
		Message: message,
		Data:    nil,
	})
}
