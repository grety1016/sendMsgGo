package main

import (
	"time"

	"sendMsgGo/logger"
	"sendMsgGo/mssql"
	// "sendMsgGo/eventful"

	// "sendMsgGo/mssql/mssqldemo"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
)

type DBConfig = mssql.DBConfig

func main() {
	logger.Init()

	// eventful.EventDemo() // 测试event的例子

	// mssqldemo.SqlxDemo() // 测试mssql的例子

	// 数据库连接和连接池的配置参数(正式环境从环境变量中获取)
	dbConfig := mssql.SetDBConfig("DB_CONN_STRING", 100, 20, 60*time.Minute, 15*time.Minute)
	//初始化数据库
	db, err := mssql.InitDB(dbConfig)
	if err != nil {
		logrus.Errorf("[DB] @%s - Failed to open to database: %v }", db.GetDBName(), err)
	}

	// defer db.Close() // 关闭数据库连接
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Errorf("[DB] @%s - Failed to close database: %v }", db.GetDBName(), err)
		}
	}()

	 
}
