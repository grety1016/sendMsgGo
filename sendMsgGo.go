package main

import (
	"database/sql"
	"time"

	"sendMsgGo/logger"
	"sendMsgGo/middleware"
	"sendMsgGo/mssql"

	// "sendMsgGo/eventful"

	// "sendMsgGo/mssql/mssqldemo"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
)

type DBConfig = mssql.DBConfig
type DB = mssql.DBWrapper

func main() {
	logger.Init()

	// eventful.EventDemo() // 测试event的例子

	// mssqldemo.SqlxDemo() // 测试mssql的例子

	// 数据库连接和连接池的配置参数(正式环境从环境变量中获取)
	dbConfig := mssql.SetDBConfig("DB_CONN_STRING", 100, 20, 60*time.Minute, 15*time.Minute)
	//初始化数据库
	db, err := mssql.InitDB(dbConfig)
	if err != nil {
		logrus.Errorf("[DB] @%s - Failed to open to database: %v }", db.DBName(), err)
	}

	// defer db.Close() // 关闭数据库连接
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Errorf("[DB] @%s - Failed to close database: %v }", db.DBName(), err)
		}
	}()
	// 启动gin服务
	r := gin.Default()
	r.Use(middleware.DBMiddleware(db))
	r.GET("/ping", func(c *gin.Context) {
		db := c.MustGet("db").(*DB)

		var todoList []TodoList

		db.QueryCollect(&todoList, "SELECT * FROM getTodoList(@status,@phone)", sql.Named("status", "2"), sql.Named("phone", "15345923407"))

		c.JSON(http.StatusOK, todoList)

	})
	r.Run(":3888")

}

// 定义结构体
type TodoList struct {
	EventName      string    `json:"eventName" db:"eventName"`
	RN             int       `json:"rn" db:"rn"`
	FStatus        string    `json:"fStatus" db:"fStatus"`
	FNumber        string    `json:"fNumber" db:"fNumber"`
	FFormID        string    `json:"fFormID" db:"fFormID"`
	FFormType      string    `json:"fFormType" db:"fFormType"`
	FDisplayName   string    `json:"fDisplayName" db:"fDisplayName"`
	TodoStatus     string    `json:"todoStatus" db:"todoStatus"`
	FName          string    `json:"fName" db:"fName"`
	SenderPhone    string    `json:"senderPhone" db:"senderPhone"`
	FReceiverNames string    `json:"fReceiverNames" db:"fReceiverNames"`
	FPhone         string    `json:"fPhone" db:"fPhone"`
	FProcInstID    string    `json:"fProcinstID" db:"fProcinstID"`
	FCreateTime    time.Time `json:"fCreateTime" db:"fCreateTime"`
}
