// Package utils 提供通用工具函数
package utils

import (
	"reflect"
	"strings"
	"sync"

	"go_wails_project_manager/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// GetValidator 获取验证器单例
func GetValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()

		// 使用JSON标签作为字段名
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return fld.Name
			}
			return name
		})

		// 注册自定义验证器
		registerCustomValidators(validate)
	})
	return validate
}

// registerCustomValidators 注册自定义验证器
func registerCustomValidators(v *validator.Validate) {
	// 手机号验证
	v.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
		mobile := fl.Field().String()
		if len(mobile) != 11 {
			return false
		}
		return mobile[0] == '1'
	})

	// 用户名验证（字母开头，允许字母数字下划线）
	v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) < 3 || len(username) > 20 {
			return false
		}
		if !isLetter(rune(username[0])) {
			return false
		}
		for _, c := range username {
			if !isLetter(c) && !isDigit(c) && c != '_' {
				return false
			}
		}
		return true
	})
}

func isLetter(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

// ValidateStruct 验证结构体
func ValidateStruct(obj interface{}) error {
	return GetValidator().Struct(obj)
}

// ValidationError 验证错误信息
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatValidationErrors 格式化验证错误
func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

// getErrorMessage 获取错误消息
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + "不能为空"
	case "email":
		return "邮箱格式不正确"
	case "min":
		return e.Field() + "长度不能小于" + e.Param()
	case "max":
		return e.Field() + "长度不能大于" + e.Param()
	case "len":
		return e.Field() + "长度必须为" + e.Param()
	case "mobile":
		return "手机号格式不正确"
	case "username":
		return "用户名格式不正确（3-20位，字母开头，只允许字母数字下划线）"
	case "gte":
		return e.Field() + "必须大于等于" + e.Param()
	case "lte":
		return e.Field() + "必须小于等于" + e.Param()
	case "gt":
		return e.Field() + "必须大于" + e.Param()
	case "lt":
		return e.Field() + "必须小于" + e.Param()
	case "oneof":
		return e.Field() + "必须是以下值之一: " + e.Param()
	case "url":
		return "URL格式不正确"
	case "alphanum":
		return e.Field() + "只能包含字母和数字"
	default:
		return e.Field() + "验证失败"
	}
}

// GetFirstErrorMessage 获取第一个错误消息
func GetFirstErrorMessage(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		if len(validationErrors) > 0 {
			return getErrorMessage(validationErrors[0])
		}
	}
	return "参数验证失败"
}

// BindAndValidate 绑定并验证请求参数
func BindAndValidate(c *gin.Context, obj interface{}) bool {
	// 尝试绑定JSON
	if err := c.ShouldBindJSON(obj); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return false
	}

	// 验证参数
	if err := ValidateStruct(obj); err != nil {
		response.ValidationError(c, GetFirstErrorMessage(err))
		return false
	}

	return true
}

// BindQueryAndValidate 绑定并验证Query参数
func BindQueryAndValidate(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return false
	}

	if err := ValidateStruct(obj); err != nil {
		response.ValidationError(c, GetFirstErrorMessage(err))
		return false
	}

	return true
}

// BindURIAndValidate 绑定并验证URI参数
func BindURIAndValidate(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindUri(obj); err != nil {
		response.BadRequest(c, "请求参数格式错误")
		return false
	}

	if err := ValidateStruct(obj); err != nil {
		response.ValidationError(c, GetFirstErrorMessage(err))
		return false
	}

	return true
}
