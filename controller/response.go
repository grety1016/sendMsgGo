package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const ABORTWITHSTATUS = -1 //用于全局模拟错误终止路由流转

type JsonSuccess struct {
	Code  int         `json:"code"`
	Msg   interface{} `json:"msg"`
	Data  interface{} `json:"data"`
	Count int         `json:"count"`
}

type JsonError struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

func ResponseSuccess(c *gin.Context, code int, msg interface{}, data interface{}, count int) {
	json := JsonSuccess{
		Code: code, Msg: msg, Data: data, Count: count,
	}
	if code == -1 {
		c.AbortWithStatusJSON(http.StatusOK, json)
	} else {
		c.JSON(http.StatusOK, json)
	}

}

func ResponseError(c *gin.Context, code int, msg interface{}) {
	json := JsonError{
		Code: code, Msg: msg,
	}

	// 设置 HTTP 状态码
	c.Writer.WriteHeader(code)
	// 将错误信息记录到 gin.Context 中
	c.Error(fmt.Errorf("%v", msg))
	// 使用 c.AbortWithStatusJSON 确保中止后续中间件的执行
	c.AbortWithStatusJSON(c.Writer.Status(), json)
}
