package middleware

import (  
	"sendmsggo/util/jwtgo"
	"strings"
	"sendmsggo/controller"
	"github.com/gin-gonic/gin"
)

func ValidateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")//获取token
		if authorization != "" {
			authorization = strings.Replace(authorization, "Bearer ", "", 1)
		}


		if c.Request.URL.Path != "/user/getsmscode" && c.Request.URL.Path != "/user/login" {
			//验证token是否有效
			isValid, err := jwtgo.ValidateJWT(authorization)
			if err != nil {
				controller.ResponseSuccess(c,controller.ABORTWITHSTATUS, "您当前访问未经授权验证,请前往登录","",1)
				return
			}else{ 
				c.Set("fphone", isValid.UserPhone)
			}
		}
		//继续处理请求
		c.Next()
	}

}
