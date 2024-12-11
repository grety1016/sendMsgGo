package middleware

import (
	"sendmsggo/util/mssql"

	"github.com/gin-gonic/gin"
)

// #region 数据库中间件

// 中间件函数，将数据库连接注入到上下文中
type DB = mssql.DBWrapper

func DBMiddleware(db *DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// #endregion 数据库中间件
