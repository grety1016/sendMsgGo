package main

import (
	"time"

	"sendmsg/logger"
	"sendmsg/mssql"

	// "sendmsg/mssql/mssqldemo"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
)

type DBConfig = mssql.DBConfig

func main() {
	logger.Init()

	// mssqldemo.SqlxDemo() // 测试mssql的例子

	// 数据库连接和连接池的配置参数(正式环境从环境变量中获取)
	dbConfig := DBConfig{
		ConnString:      "server=47.103.31.8;port=1433;user id=kxs_dev;password=kephi;database=Kxs_Interface;encrypt=true;trustServerCertificate=true;connection timeout=30;application name=sendmsg;",
		MaxOpenConns:    200,
		MaxIdleConns:    200,
		ConnMaxLifetime: 60 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	}
	db, err := mssql.InitDB(dbConfig)
	if err != nil {
		logrus.Errorf("[DB] @%s - Failed to open to database: %v }", db.GetDBName(), err)
	}
	defer db.Close()

}
