package router

import (
	"sendmsggo/controller"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	//用户登录相关路由
	user := r.Group("/user")
	{
		user.GET("getsmscode", controller.GetSmsCode)//获取短信验证码
		user.POST("login", controller.LoginPost)//用户登录
	}
	//流程表单相关路由
	flowform := r.Group("/flowform")
	{
		flowform.GET("getitemlist", controller.GetItemList)//获取表单项列表
		flowform.GET("getflowdetailfybxandclbx", controller.GetItemDetailFybxAndClbx)
	}

}
