package model

import "encoding/json"

// #region User界面相关数据结构
type LoginUser struct { //用户登录提交表单
	UserPhone json.Number `json:"userphone" validate:"userPhone"`
	SmsCode   json.Number `json:"smscode" validate:"smsCode"`
}

// #endregion User界面相关数据结构

// #region flowform相关数据结构

//定义获取待办事项列表所需要的IItem参数
type ItemListRequest struct {
	UserPhone  string `form:"userphone" validate:"userPhone"`
	ItemStatus string `form:"itemstatus" validate:"itemStatus"`
}

//获取流程事项列表返回数据接口
type FlowItemList struct {
	EventName      string `json:"eventName" db:"eventName"`           // 事项名称
	Rn             int32  `json:"rn" db:"rn"`                         // 类型修改为 int32
	FStatus        string `json:"fStatus" db:"fStatus"`               // （运行中、挂起、终止、暂停）
	FNumber        string `json:"fNumber" db:"fNumber"`               // 实例编码
	FFormID        string `json:"fFormID" db:"fFormID"`               // 流程单据ID
	FFormType      string `json:"fFormType" db:"fFormType"`           // 流程单据类型
	FDisplayName   string `json:"fDisplayName" db:"fDisplayName"`     // 实例名称
	TodoStatus     int32  `json:"todoStatus" db:"todoStatus"`         // 处理状态（未处理、已处理）0：未处理 1：已处理 2：我发起
	FName          string `json:"fName" db:"fName"`                   // 流程发起者
	SenderPhone    string `json:"senderPhone" db:"senderPhone"`       // 发起者电话
	FReceiverNames string `json:"fReceiverNames" db:"fReceiverNames"` // 流程接收者
	FPhone         string `json:"fPhone" db:"fPhone"`                 // 接收者手机号
	FProcinstID    string `json:"fProcinstID" db:"fProcinstID"`       // 实例内码
	FCreateTime    string `json:"fCreateTime" db:"fCreateTime"`       // 流程创建日期
}

//定义流程事项明细表头信息接口(费用报销-差旅报销)
type FlowItemDetailFybxAndClbx struct {
	Available          int     `json:"available" db:"available"`
	FBillNo            string  `json:"fBillNo" db:"fBillNo"`
	FOrgID             string  `json:"fOrgID" db:"fOrgID"`
	FRequestDeptID     string  `json:"fRequestDeptID" db:"fRequestDeptID"`
	FProposerID        string  `json:"fProposerID" db:"fProposerID"`
	FExpenseOrgID      string  `json:"fExpenseOrgID" db:"fExpenseOrgID"`
	FExpenseDeptID     string  `json:"fExpenseDeptID" db:"fExpenseDeptID"`
	FCurrency          string  `json:"fCurrency" db:"fCurrency"`
	FReqReimbAmountSum float64 `json:"fReqReimbAmountSum" db:"fReqReimbAmountSum"`
	FExpAmountSum      float64 `json:"fExpAmountSum" db:"fExpAmountSum"`
	FCausa             string  `json:"fCausa" db:"fCausa"`
	Years              string  `json:"years" db:"years"`
	Status             string  `json:"status" db:"status"`
}

// #endregion flowform相关数据结构
