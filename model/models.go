package model

import "time"

//用户待办列表
type TodoList struct {
	EventName      string    `json:"eventName" db:"eventName"`
	RN             int       `json:"rn" db:"rn"`
	FStatus        string    `json:"fStatus" db:"fStatus"`
	FNumber        string    `json:"fNumber" db:"fNumber"`
	FFormID        string    `json:"fFormID" db:"fFormID"`
	FFormType      string    `json:"fFormType" db:"fFormType"`
	FDisplayName   string    `json:"fDisplayName" db:"fDisplayName"`
	TodoStatus     string    `json:"todoStatus" db:"todoStatus"`
	FName          string    `json:"fName" db:"fName"`
	SenderPhone    string    `json:"senderPhone" db:"senderPhone"`
	FReceiverNames string    `json:"fReceiverNames" db:"fReceiverNames"`
	FPhone         string    `json:"fPhone" db:"fPhone"`
	FProcInstID    string    `json:"fProcinstID" db:"fProcinstID"`
	FCreateTime    time.Time `json:"fCreateTime" db:"fCreateTime"`
}
