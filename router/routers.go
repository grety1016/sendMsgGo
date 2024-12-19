package router

import (
	"sendmsggo/controller"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	//用户登录相关路由
	user := r.Group("/user")
	{
		user.GET("getsmscode", controller.GetSmsCode) //获取短信验证码
		user.POST("login", controller.LoginPost)      //用户登录
	}
	//流程表单相关路由
	flowform := r.Group("/flowform")
	{
		flowform.GET("getitemlist", controller.GetItemList)                           //获取待办-已办-我发起的列表
		flowform.GET("getflowdetailfybxandclbx", controller.GetItemDetailFybxAndClbx) //获取费用报销-差旅报销详情
		flowform.GET("getflowdetailrowsfybx", controller.GetFlowDetailRowsFybx)       //获取费用报销明细行
		flowform.GET("getflowdetailflowchart", controller.GetFlowDetailFlowChart)     //获取费用报销流程图
		flowform.GET("checkfileexist", controller.CheckFileExist)                     //检查文件在获取明细时时否转换成功
	}

}
