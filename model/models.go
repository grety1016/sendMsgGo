package model

// #region User界面相关数据结构
type LoginUser struct {
	UserPhone string `json:"userPhone" validate:"userPhone"`
	SmsCode   int    `json:"smsCode" validate:"smsCode"` 
}
 