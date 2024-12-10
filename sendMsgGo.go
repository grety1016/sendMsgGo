package main

import (
	"sendmsggo/logger"
	"sendmsggo/middleware"
	"sendmsggo/mssql"
	"sendmsggo/router"
	"time"

	// "sendmsggo/eventful"

	// "sendMsgGo/mssql/mssqldemo"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 初始化日志
	logger.Init()

	db, err := initDB()
	if err != nil {
		logrus.Fatalf("[DB] @%s - Failed to open to database: %v }", db.DBName(), err)
	}
	//程序退出时关闭数据库连接
	defer db.Close()

	// gin.SetMode(gin.ReleaseMode)// 设置运行模式

	r := gin.Default()                 // gin实例化
	r.Use(middleware.DBMiddleware(db)) // 数据库注入gin中间件
	r.Use(middleware.HttpLogger())     // http请求日志记录中间件

	router.Router(r) // 路由注册

	r.Run(":3888") // 启动服务
	

	// eventful.EventDemo() // 测试event的例子

	// mssqldemo.SqlxDemo() // 测试mssql的例子

}

// mssql包的别名
type DBConfig = mssql.DBConfig
type DB = mssql.DBWrapper

func initDB() (*DB, error) {
	dbConfig := mssql.SetDBConfig("DB_CONN_STRING", 100, 20, 60*time.Minute, 15*time.Minute)
	db, err := mssql.InitDB(dbConfig)
	if err != nil {
		logrus.Errorf("[DB] @%s - Failed to open to database: %v }", db.DBName(), err)
		return nil, err
	}
	return db, nil
}
