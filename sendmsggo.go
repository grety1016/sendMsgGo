package main

import (
	"sendmsggo/cstvalidator" //自定义验证器包
	"sendmsggo/middleware"   //自定义中间件包
	"sendmsggo/router"       //路由注册包
	"sendmsggo/util/ddtoken" //自定义token包
	"sendmsggo/util/logger"  //自定义日志库
	"sendmsggo/util/mssql"   //自定义mssql包
	"time"                   //时间库

	// "sendmsggo/util/eventful"

	// "sendmsggo/util/mssql/mssqldemo"

	_ "github.com/denisenkom/go-mssqldb" //mssql驱动
	"github.com/gin-contrib/gzip"

	//gzip压缩中间件
	"github.com/gin-gonic/gin"   //gin框架
	"github.com/sirupsen/logrus" //日志库
)

func main() {

	logger.Init() // 初始化日志库

	cstvalidator.InitValidator() // 初始化自定义验证器

	//创建goroutine更新用户列表
	go func() {
		if err := ddtoken.LocalThread(); err != nil {
			logrus.Errorf("[DD]-获取用户列表更新失败:%v", err)
			return
		}
	}()

	// 初始化数据库封装实例对象
	db, err := initDB()
	if err != nil {
		logrus.Errorf("[DB] @%s - Failed to open to database: %v ", db.DBName(), err)
	}
	logrus.Infof("[DB] @%s - Connecte database success ", db.DBName())

	//程序退出时关闭数据库连接
	defer db.Close()

	// gin.SetMode(gin.ReleaseMode) // 设置运行模式

	r := gin.New() // gin实例化

	r.SetTrustedProxies([]string{"192.168.0.31"}) //设置信任代理IP地址

	r.Use(middleware.HttpLogger())                           // http请求日志记录中间件&&panic恢复中间件&&请求地址转换成小写
	r.Use(middleware.ValidateJWT())                          // JWT验证中间件
	r.Use(middleware.DBMiddleware(db))                       // 数据库注入gin中间件
	r.StaticFS("/files", gin.Dir("D:/kingdee  File", false)) // 将 /files/ 路径映射到 D:\kingkee files\
	// 使用最快压缩级别
	r.Use(gzip.Gzip(gzip.BestSpeed))

	router.Router(r) // 路由注册

	r.Run("localhost:3888") // 启动服务

}

type DBConfig = mssql.DBConfig // mssql数据库配置对象别名
type DB = mssql.DBWrapper      // mssql数据库封装对象别名

// 本地初始化数据库函数
func initDB() (*DB, error) {
	var dbconfigstr = "server=47.103.31.8;port=1433;user id=kxs_dev;password=kephi;database=Kxs_Interface;encrypt=true;trustServerCertificate=true;connection timeout=30;application name=sendmsg"
	dbConfig := mssql.SetDBConfig(dbconfigstr, 100, 20, 60*time.Minute, 15*time.Minute)
	db, err := mssql.InitDB(dbConfig)
	if err != nil {
		logrus.Errorf("[DB] @%s - Failed to open to database: %v ", db.DBName(), err)
		return nil, err
	}
	return db, nil
}
