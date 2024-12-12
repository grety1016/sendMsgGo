package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"sendmsggo/model"
	"sendmsggo/util/mssql"

	"github.com/gin-gonic/gin"
)

type DB = mssql.DBWrapper      //mssql.DBWrapper类型别名
type TodoList = model.TodoList //model.TodoList类型别名

// GetTodoList 获取待办列表
func GetTodoList(c *gin.Context) {
	db := c.MustGet("db").(*DB)

	var todoList []TodoList
	err := db.QueryCollect(&todoList, "SELECT *  FROM getTodoList(@status,@phone)", sql.Named("status", "2"), sql.Named("phone", "15345923407"))
	if err != nil {
		fmt.Printf("GetTodoList error:%v", err)
		ResponseError(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ResponseSuccess(c, http.StatusOK, "success", todoList, len(todoList))
}
