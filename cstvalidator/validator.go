package cstvalidator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once // 保证只初始化一次
)

// InitValidator 初始化全局 validator 实例
func InitValidator() {
	once.Do(func() {
		validate = validator.New()
		// 自定义验证器
		validate.RegisterValidation("userPhone", validatePhoneNumber) // 注册手机号码自定义验证器
		validate.RegisterValidation("smsCode", validateSmsCode)       // 注册短信验证码自定义验证器
		validate.RegisterValidation("itemStatus", validateItemStatus) // 注册列表状态码自定义验证器
	})
}

// GetValidator 返回全局 validator 实例
func GetValidator() *validator.Validate {
	return validate
}
