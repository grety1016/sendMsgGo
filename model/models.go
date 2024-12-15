package model

// #region User界面相关数据结构
type LoginResponse struct {
	UserPhone string `json:"userPhone"`
	SmsCode   int    `json:"smsCode"`
	Token     string `json:"token"`
	Code      int    `json:"code"`
	ErrMsg    string `json:"errMsg"`
}
