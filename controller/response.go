package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
	c.JSON(http.StatusOK, json)
}

func ResponseError(c *gin.Context, code int, msg interface{}) {
	json := JsonError{
		Code: code, Msg: msg,
	}
	c.JSON(http.StatusOK, json)
}
