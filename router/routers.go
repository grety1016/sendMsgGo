package router

import (
	"database/sql"
	"net/http"
	"sendmsggo/mssql"
	"time"

	"github.com/gin-gonic/gin"
)

type DB = mssql.DBWrapper

func Router(r *gin.Engine) *gin.Engine {
 
	r.GET("/", func(ctx *gin.Context) {
		db := ctx.MustGet("db").(*DB)

		var todoList []TodoList
		err := db.QueryCollect(&todoList, "SELECT * FROM getTodoList(@status,@phone)", sql.Named("status", "2"), sql.Named("phone", "15345923407"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, todoList)
	})

	return r
}

// 如下为类型定义
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
