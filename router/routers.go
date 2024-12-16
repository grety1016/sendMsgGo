package router

import (
	"sendmsggo/controller"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {

	user := r.Group("/user")
	{
		user.GET("getsmscode", controller.GetSmsCode)
		user.POST("/login", controller.LoginPost)
	}

}
