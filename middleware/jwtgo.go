package middleware

import (
	"net/http"
	"sendmsggo/controller"
	"sendmsggo/util/jwtgo"
	"strings"

	"github.com/gin-gonic/gin"
)

func ValidateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization") //获取token
		if authorization != "" {
			authorization = strings.Replace(authorization, "Bearer ", "", 1)
		}

		if c.Request.URL.Path != "/user/getsmscode" && c.Request.URL.Path != "/user/login" && !strings.Contains(c.Request.URL.Path, "/files/") {
			//验证token是否有效
			isValid, err := jwtgo.ValidateJWT(authorization)
			if err != nil {
				controller.ResponseError(c, http.StatusForbidden, "您当前访问未经授权验证,请前往 http://47.103.31.8:3189 登录")
				return
			} else {
				c.Set("fphone", isValid.UserPhone)
			}
		}
		//继续处理请求
		c.Next()
	}

}
