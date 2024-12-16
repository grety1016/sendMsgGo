package validator

 

import ( 
    "regexp"
    "github.com/go-playground/validator/v10"
) 

// 定义手机号码验证函数
func validatePhoneNumber(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    re := regexp.MustCompile(`^1[3456789]\d{9}$`)
    return re.MatchString(phone)
}

// 定义短信验证码验证函数
func validateSmsCode(fl validator.FieldLevel) bool {
    smsCode := fl.Field().Interface().(int)
    return smsCode >= 1000 && smsCode <= 9999 // 验证是否为4位数字
}


 



 
