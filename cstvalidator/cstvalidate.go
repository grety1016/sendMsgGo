package cstvalidator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// 自定义手机号码验证函数
func validatePhoneNumber(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	re := regexp.MustCompile(`^1[3456789]\d{9}$`)
	return re.MatchString(phone)
}

// 自定义短信验证码验证函数
func validateSmsCode(fl validator.FieldLevel) bool {
	smsCode := fl.Field().String()
	re := regexp.MustCompile(`^\d{4}$`)
	return re.MatchString(smsCode)
}

// 自定义待办列表状态验证函数
func validateItemStatus(fl validator.FieldLevel) bool {
	itemStatus := fl.Field().String()
	re := regexp.MustCompile(`^[012]$`)
	return re.MatchString(itemStatus)
}
