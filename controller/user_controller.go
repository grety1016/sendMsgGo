package controller

import (
	"database/sql"

	"math/rand"
	"net/http"
	"sendmsggo/cstvalidator"
	"sendmsggo/model"
	"sendmsggo/util/ddtoken"
	"sendmsggo/util/jwtgo"
	"sendmsggo/util/logger"
	"sendmsggo/util/mssql"

	"github.com/gin-gonic/gin"
)

type DB = mssql.DBWrapper //mssql.DBWrapper类型别名

// 处理获取短信验证码的请求
func GetSmsCode(c *gin.Context) {
	code := http.StatusOK //返回状态码
	msg := "验证码发送成功"      //错误信息
	var smsCode int = 0   //短信验证码

	userphone := c.Query("userphone") //获取url查询参数 userphone

	// 获取手机号并验证手机号格式
	if err := cstvalidator.GetValidator().Var(userphone, "userPhone"); err != nil {
		if userphone == "" {
			msg = "参数不正确!"
		} else {
			msg = "手机号格式错误!"
		}
		code = http.StatusBadRequest
		ResponseSuccess(c, code, msg, "", 1)
		return
	}

	db := c.MustGet("db").(*DB) //从中间件获取db对象

	// 查询当前手机是否在消息用户列表中存在有效验证码
	query := "SELECT DATEDIFF(SECOND, createdtime, GETDATE()) FROM dbo.sendMsg_users WHERE userPhone = @userphone"

	seconds, err := db.QueryValue(query, sql.Named("userphone", userphone))
	if err != nil {
		if err == sql.ErrNoRows {
			msg = "该手机号未注册!"
			code = http.StatusNotFound
			ResponseSuccess(c, code, msg, "", 1)
			return
		} else {
			msg = err.Error()
			code = http.StatusInternalServerError
			ResponseError(c, code, msg)
			return
		}
	}
	// 验证码有效期为60秒
	if seconds.(int64) < 60 {
		code = http.StatusTooManyRequests
		msg = "操作过于频繁，请复制最近一次验证码或一分钟后重试"
	} else {
		smsCode = rand.Intn(9000) + 1000
		query = "UPDATE dbo.sendMsg_users SET smsCode = :smsCode, createdtime = GETDATE() WHERE userPhone = :userphone"
		_, err := db.ExecSQLWithTran(query, map[string]interface{}{"smsCode": smsCode, "userphone": userphone})
		if err != nil {
			msg = "验证码生成失败!"
			code = http.StatusInternalServerError
			ResponseSuccess(c, code, msg, "", 1)
			return
		} else {
			// 验证码生成成功
			var ddmsg = []ddtoken.SmsMessage{}
			query = "SELECT  '' as ddtoken,dduserid,userphone,robotcode,smscode   FROM sendMsg_users  WITH(NOLOCK)  WHERE userphone = @userphone"
			db.QueryCollect(&ddmsg, query, sql.Named("userphone", userphone))
			if ddmsg[0].RobotCode == "dingrw2omtorwpetxqop" {
				gzym_ddtoken := ddtoken.NewDDToken(
					"https://oapi.dingtalk.com/gettoken",
					"dingrw2omtorwpetxqop",
					"Bcrn5u6p5pQg7RvLDuCP71VjIF4ZxuEBEO6kMiwZMKXXZ5AxQl_I_9iJD0u4EQ-N")
				ddmsg[0].DDToken, err = gzym_ddtoken.GetToken()
				if err != nil {
					msg = "获取钉钉token失败!"
					code = http.StatusInternalServerError
					ResponseSuccess(c, code, msg, "", 1)
					return
				}

			} else {
				zb_ddtoken := ddtoken.NewDDToken(
					"https://oapi.dingtalk.com/gettoken",
					"dingzblrl7qs6pkygqcn",
					"26GGYRR_UD1VpHxDBYVixYvxbPGDBsY5lUB8DcRqpSgO4zZax427woZTmmODX4oU")
				ddmsg[0].DDToken, err = zb_ddtoken.GetToken()
				if err != nil {
					msg = "获取钉钉token失败!"
					code = http.StatusInternalServerError
					ResponseSuccess(c, code, msg, "", 1)
					return
				}
			}
			// 发送钉钉消息
			// err = ddmsg[0].SendSmsCode()
			// if err != nil {
			// 	msg = "验证码发送失败!"
			// 	code = http.StatusInternalServerError
			// }

		}

	}
	// 验证码发送成功
	ResponseSuccess(c, code, msg, "", 1)
}

// 处理登录请求
func LoginPost(c *gin.Context) {
	var msg string //返回信息
	var code int   //返回状态码

	// 绑定json数据到结构体
	var user model.LoginUser
	// 绑定json数据到结构体，验证字段是否匹配
	err := c.ShouldBindJSON(&user)
	if err != nil {
		msg = "请求参数格式错误!"
		code = http.StatusBadRequest
		ResponseSuccess(c, code, msg, "", 1)
		return
	}
	//从validator自定义验证器中获取验证器实例调用正则验证手机号及验证码格式
	err = cstvalidator.GetValidator().Struct(user) //验证结构体
	if err != nil {
		msg = "手机号或验证码格式错误!"
		code = http.StatusBadRequest
		ResponseSuccess(c, code, msg, "", 1)
		return
	}

	// 从中间件获取db对象
	db := c.MustGet("db").(*DB)

	// 查询当前手机是否在消息用户列表中存在有效验证码
	query := "SELECT 1 FROM dbo.sendMsg_users WHERE userPhone = @userphone AND smsCode = @smsCode and DATEDIFF(SECOND, createdtime, GETDATE()) < 180"
	params := []interface{}{sql.Named("userphone", user.UserPhone), sql.Named("smsCode", user.SmsCode)}
	exist, err := db.QueryValue(query, params...)
	if err != nil {
		if err == sql.ErrNoRows {
			msg = "手机号或验证码错误!"
			code = http.StatusNotFound
			ResponseSuccess(c, code, msg, "", 1)
			return
		} else {
			msg = err.Error()
			code = http.StatusInternalServerError
			ResponseError(c, code, msg)
			return
		}
	}
	//验证成功
	if exist.(int64) == 1 {
		// 生成token
		token, err := jwtgo.GenerateJWT(user.UserPhone.String()) //因为user.UserPhone是json.Number类型，所以需要转换成string类型
		if err != nil {
			logger.InitHTTPLogger().Errorf("[HTTP]-Error generating token:%s", err.Error())
			ResponseSuccess(c, http.StatusInternalServerError, err.Error(), "", 1)
			return
		} else {
			// 设置响应头
			c.Header("Authorization", "Bearer "+token)
			ResponseSuccess(c, http.StatusOK, "登录成功", "", 1)
		}
	}

}
