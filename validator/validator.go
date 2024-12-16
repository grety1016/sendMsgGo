package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// InitValidator 初始化全局 validator 实例
func InitValidator() {
	once.Do(func() {
		validate = validator.New()
		// 自定义验证器
		validate.RegisterValidation("userPhone", validatePhoneNumber)
		validate.RegisterValidation("smsCode", validateSmsCode)
	})
}

// GetValidator 返回全局 validator 实例
func GetValidator() *validator.Validate {
	return validate
}
