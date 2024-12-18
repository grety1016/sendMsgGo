package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"sendmsggo/cstvalidator"
	"sendmsggo/model"

	"github.com/gin-gonic/gin"
)

// GetItemList 待办列表
func GetItemList(c *gin.Context) {
	var msg string //返回信息
	var code int   //返回状态码

	var getItemList model.ItemListRequest

	err := c.ShouldBindQuery(&getItemList)
	if err != nil {
		msg = "请求参数不正确"
		code = http.StatusBadRequest
		ResponseSuccess(c, code, msg, "", 1)
		return
	}

	//从validator自定义验证器中获取验证器实例调用正则验证手机号及验证码格式
	err = cstvalidator.GetValidator().Struct(getItemList) //验证结构体
	if err != nil {
		msg = "手机号或待办状态码不正确!"
		code = http.StatusBadRequest
		ResponseSuccess(c, code, msg, "", 1)
		return
	}

	//查询数据库获取数据
	db := c.MustGet("db").(*DB)

	var itemList []model.FlowItemList //待办列表接收对象

	params := []interface{}{sql.Named("itemstatus", getItemList.ItemStatus), sql.Named("userphone", getItemList.UserPhone)} //参数化查询

	query := "SELECT * FROM getTodoList(@itemstatus,@userphone)" //查询语句

	err = db.QueryCollect(&itemList, query, params...) //查询执行结果
	if err != nil {
		msg = "获取待办列表失败!"
		code = http.StatusInternalServerError
		ResponseSuccess(c, code, msg, "", 1)
		return
	} else {
		ResponseSuccess(c, http.StatusOK, "获取成功", itemList, len(itemList))
	}

}

func GetItemDetailFybxAndClbx(c *gin.Context) {
	// var msg string //返回信息
	// var code int   //返回状态码
	// var flowdetailfybxandclbx model.FlowItemDetailFybxAndClbx

	var fProcinstID = c.Query("fprocinstid")
	// var fPhone = c.MustGet("fphone")
	fmt.Println("fProcinstID:", fProcinstID)
	// fmt.Println("fPhone:", fPhone)

	// db := c.MustGet("db").(*DB)

	// //查询数据库获取数据
	// query := "SELECT * FROM getFlowDetailFybxAndClbx(@p1,@p2)" //查询语句
	// db.QueryCollect(&flowdetailfybxandclbx, query string, args ...interface{})

}
