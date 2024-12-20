package main

import (
	"database/sql"
	"time"

	"sendmsggo/logger"
	"sendmsggo/middleware"
	"sendmsggo/modeltype"
	"sendmsggo/mssql"

	"sendmsggo/eventful"

	// "sendMsgGo/mssql/mssqldemo"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
)

// mssql包的别名
type DBConfig = mssql.DBConfig
type DB = mssql.DBWrapper

// modeltype包的别名
type TodoList = modeltype.TodoList

func main() {
	logger.Init()

	eventful.EventDemo() // 测试event的例子

	// mssqldemo.SqlxDemo() // 测试mssql的例子

	// 数据库连接和连接池的配置参数(正式环境从环境变量中获取)
	dbConfig := mssql.SetDBConfig("DB_CONN_STRING", 100, 20, 60*time.Minute, 15*time.Minute)
	//初始化数据库
	db, err := mssql.InitDB(dbConfig)
	if err != nil {
		logrus.Errorf("[DB] @%s - Failed to open to database: %v }", db.DBName(), err)
	}

	// defer db.Close() // 确保程序退出关闭数据库连接
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Errorf("[DB] @%s - Failed to close database: %v }", db.DBName(), err)
		}
	}()

	// 启动gin服务
	r := gin.Default() //创建gin实例

	r.Use(middleware.DBMiddleware(db)) //将数据库连接对象注入到gin的上下文中

	r.Use(middleware.HttpLogger()) //http请求日志记录

	// 注册路由
	r.GET("/ping", func(c *gin.Context) {
		db := c.MustGet("db").(*DB)

		var todoList []TodoList

		db.QueryCollect(&todoList, "SELECT * FROM getTodoList(@status,@phone)", sql.Named("status", "2"), sql.Named("phone", "15345923407"))

		c.JSON(http.StatusOK, todoList)

	})
	// 启动服务
	r.Run(":3888")
}
