package main

import (
	"sendmsggo/middleware"
	"sendmsggo/router"
	"sendmsggo/util/ddtoken"
	"sendmsggo/util/logger"
	"sendmsggo/util/mssql"
	"sendmsggo/validator"
	"time"

	// "sendmsggo/util/eventful"

	// "sendmsggo/util/mssql/mssqldemo"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 初始化日志
	logger.Init()

	validator.InitValidator() // 初始化验证器

	//token获取
	go func() {
		if err := ddtoken.LocalThread(); err != nil {
			logrus.Errorf("[DD]-获取用户列表更新失败:%v", err)
			return
		}
	}()

	// eventful.EventDemo() // 测试event的例子

	// mssqldemo.SqlxDemo() // 测试mssql的例子

	// 初始化数据库
	db, err := initDB()
	if err != nil {
		logrus.Fatalf("[DB] @%s - Failed to open to database: %v ", db.DBName(), err)
	}
	logrus.Infof("[DB] @%s - Connecte database success ", db.DBName())

	//程序退出时关闭数据库连接
	defer db.Close()

	// gin.SetMode(gin.ReleaseMode) // 设置运行模式

	r := gin.New() // gin实例化

	r.SetTrustedProxies([]string{"192.168.0.31"})

	r.Use(middleware.HttpLogger())     // http请求日志记录中间件&&panic恢复中间件&&请求地址转换成小写
	r.Use(middleware.DBMiddleware(db)) // 数据库注入gin中间件

	router.Router(r) // 路由注册

	r.Run("localhost:3888") // 启动服务

}

// mssql包的别名
type DBConfig = mssql.DBConfig
type DB = mssql.DBWrapper

// 初始化数据库
func initDB() (*DB, error) {
	var dbconfigstr = "server=47.103.31.8;port=1433;user id=kxs_dev;password=kephi;database=Kxs_Interface;encrypt=true;trustServerCertificate=true;connection timeout=30;application name=sendmsg"
	dbConfig := mssql.SetDBConfig(dbconfigstr, 100, 20, 60*time.Minute, 15*time.Minute)
	db, err := mssql.InitDB(dbConfig)
	if err != nil {
		logrus.Fatalf("[DB] @%s - Failed to open to database: %v ", db.DBName(), err)
		return nil, err
	}
	return db, nil
}
