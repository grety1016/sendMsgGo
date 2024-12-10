package mssqldemo

//*自己的主函数main中引入 "项目名/sqlx/sqlxdemo"，然后调用sqlxdemo.SqlxDemo()函数即可
import (
	// "net/http"

	"database/sql"
	"fmt"
	"time"

	// "github.com/jmoiron/sqlx"

	// "sendmsg/middleware"
	"sendmsggo/mssql"

	"github.com/sirupsen/logrus"
	// "github.com/gin-gonic/gin"
)

//#region mssqldemo

// 如下函数用于测试sqlx的使用
type UserParams struct {
	Name int64 `db:"name"`
	Age  int   `db:"age"`
	Sex  int   `db:"sex"`
}

// 定义 User 结构体
type User struct {
	UserID   int    `db:"userid"`
	UserName string `db:"username"`
}

type ID struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type TxWrapper = mssql.TxWrapper
type AsyncResult = mssql.AsyncResult
type DBConfig = mssql.DBConfig

func SqlxDemo() {
	// 配置参数
	dbConfig := mssql.SetDBConfig("DB_CONN_STRING", 100, 20, 60*time.Minute, 15*time.Minute)

	// 初始化数据库连接
	db, err := mssql.InitDB(dbConfig)
	//判断初始化数据库连接池是否成功
	if err != nil {
		logrus.Error("Failed to initialize database:", err)
		return
	}
	// 最终结束关闭数据库连接池
	defer func() {
		if err := db.Close(); err != nil {
			logrus.Error("Failed to close database:", err)
		}
	}()

	//匿名切片使用 sql.Named 创建命名参数
	params := []interface{}{
		sql.Named("username", "苏宁绿"),
		sql.Named("userphone", "15345923407"),
		sql.Named("userid", 398),
	}

	//获取单个值并进行声明和赋值
	//*表名不能与列名和条件列名一样用参数绑定，只能用拼接字符
	tableName := "sendMsg_users"
	query := fmt.Sprintf("SELECT createdtime FROM %s where userid = @id", tableName)
	if createdtime, err := db.QueryValue(query, sql.Named("id", 398)); err != nil {
		if err == sql.ErrNoRows { //返回空行需要额外判断处理
			fmt.Println("[QueryValue1]没有查询到结果")
		} else {
			fmt.Printf("[QueryValue1]查询用户ID创建时间时出错: %v\n", err)
		}
	} else {
		if tm, ok := createdtime.(time.Time); ok {
			fmt.Printf("[QueryValue1]获取到的用户ID创建时间为: %s\n", tm) //转换为time.time)
		} else {
			fmt.Printf("[QueryValue1]断言失败,实际类型: %T\n", createdtime)
		}
	}

	query = fmt.Sprintf("SELECT top 1 createdtime FROM %s ", tableName)
	if createdtime, err := db.QueryValue(query, nil); err != nil {
		if err == sql.ErrNoRows { //返回空行需要额外判断处理
			fmt.Println("[QueryValue2]没有查询到结果")
		} else {
			fmt.Printf("[QueryValue2]查询用户ID创建时间时出错: %v\n", err)
		}
	} else {
		if tm, ok := createdtime.(time.Time); ok {
			fmt.Printf("[QueryValue2]获取到的用户ID创建时间为: %s\n", tm) //转换为time.time)
		} else {
			fmt.Printf("[QueryValue2]断言失败,实际类型: %T\n", createdtime)
		}
	}

	// 查询单个值并进行声明和赋值
	if userid, err := db.QueryValue("SELECT top 1 userid FROM sendMsg_users where userPhone = @userphone and username = @username and userid = @userid", params...); err != nil {
		if err == sql.ErrNoRows { //返回空行需要额外判断处理
			fmt.Println("[QueryValue3]没有查询到结果")
		} else {
			fmt.Printf("[QueryValue3]查询时出错: %v\n", err)
		}
	} else {
		if id, ok := userid.(int64); ok {
			fmt.Printf("[QueryValue3]获取到的用户ID为: %d\n", id)
		} else {
			fmt.Printf("[QueryValue3]断言失败,实际类型: %T\n", userid)
		}
	}

	// 查询结果集并映射结构体
	var users []User
	if err := db.QueryCollect(&users, "SELECT userid,username FROM sendMsg_users where userid <= @userid",
		sql.Named("userid", 308)); err != nil {
		fmt.Printf("[QueryValue]查询用户ID时出错: %v\n", err)
	} else {
		fmt.Printf("[QueryValue]获取到的用户列表为: %v\n", users)
	}
	//字典
	usermap := map[string]interface{}{"id": 5, "name": "钱二", "age": 2, "sex": 1}
	// 插入数据(多条) 使用匿名 `map`
	usermaps := []map[string]interface{}{
		{"id": 5, "name": "钱二", "age": 2, "sex": 1},
		{"id": 6, "name": "李四", "age": 2, "sex": 2},
		{"id": 7, "name": "王五", "age": 3, "sex": 3},
	}
	id := ID{ID: 2, Name: "张三"}
	ids := []ID{{ID: 1, Name: "李四"}, {ID: 2, Name: "王五"}, {ID: 3, Name: "赵六"}}

	// 匿名map插入数据(单条)ExecSql非事务版本
	if rows, err := db.ExecSQL("INSERT INTO go_table (id,name) VALUES (:id,:name)", usermap); err != nil {
		fmt.Printf("ExecSQL插入数据时出错: %v\n", err)
	} else {
		fmt.Printf("ExecSQL插入数据成功,影响行数: %d\n", rows)
	}

	// 匿名map插入数据(多条)ExecSql
	if rows, err := db.ExecSQL("INSERT INTO go_table (id,name) VALUES (:id,:name)", usermaps); err != nil {
		fmt.Printf("ExecSQL插入数据时出错: %v\n", err)
	} else {
		fmt.Printf("ExecSQL插入数据成功,影响行数: %d\n", rows)
	}
	if rows, err := db.ExecSQL("SELECT @@VERSION", nil); err != nil {
		fmt.Printf("ExecSQL插入数据时出错: %v\n", err)
	} else {
		fmt.Printf("ExecSQL插入数据成功,影响行数: %d\n", rows)
	}
	// 结构体插入数据(单条)ExecSql
	if rows, err := db.ExecSQL("INSERT INTO go_table (id,name) VALUES (:id,:name)", id); err != nil {
		fmt.Printf("ExecSQL插入数据时出错: %v\n", err)
	} else {
		fmt.Printf("ExecSQL插入数据成功,影响行数: %d\n", rows)
	}
	if rows, err := db.ExecSQLWithTran("SELECT @@VERSION", nil); err != nil {
		fmt.Printf("ExecSQL插入数据时出错: %v\n", err)
	} else {
		fmt.Printf("ExecSQL插入数据成功,影响行数: %d\n", rows)
	}
	// 结构体插入数据(多条)ExecSql
	if rows, err := db.ExecSQL("INSERT INTO go_table (id,name) VALUES (:id,:name)", ids); err != nil {
		fmt.Printf("ExecSQL插入数据时出错: %v\n", err)
	} else {
		fmt.Printf("ExecSQL插入数据成功,影响行数: %d\n", rows)
	}

	// 插入数据(单条)ExecWithTran(用法参考ExecSQL,仅事务区别)
	if rows, err := db.ExecSQLWithTran("INSERT INTO go_table (id,name) VALUES (:id,:name)", id); err != nil {
		fmt.Printf("ExecWithTran插入数据时出错: %v\n", err)
	} else {
		fmt.Printf("ExecWithTran插入数据成功,影响行数: %d\n", rows)
	}
	// 插入数据(多条)ExecWithTran
	if rows, err := db.ExecSQLWithTran("INSERT INTO go_table (id,name) VALUES (:id,:name)", ids); err != nil {
		fmt.Printf("ExecWithTran插入数据时出错: %v\n", err)
	} else {
		fmt.Printf("ExecWithTran插入数据成功,影响行数: %d\n", rows)
	}

	//手动事务执行
	tx, err := db.BeginTran() //开启事务
	if err != nil {
		logrus.Error("Failed to initialize txWrapper:", err)
		return
	}
	//使用完毕后销毁事务对象
	defer func() {
		if err := tx.Close(); err != nil {
			logrus.Error("Failed to close tx:", err)
		}
	}()

	//删除数据(多条)ExecWithTran
	if rows, err := tx.ExecSQL("DELETE FROM dbo.go_table where id = 3", nil); err != nil {
		fmt.Printf("ExecWithTran删除数据时出错: %v\n", err)
	} else {
		fmt.Printf("ExecWithTran删除数据成功,影响行数: %d\n", rows)
	}

	tx.Commit() //提交事务

	// 下面测试使用独立事务开启，暂存点，自动回滚
	if result, err := db.BeginTranAutoCommit(func(tx *TxWrapper) (int64, error) {
		row, err := tx.ExecSQL("INSERT INTO go_table (id, name) VALUES (:id, :name)", id)

		return row, err
	}); err != nil {
		fmt.Printf("Transaction result2: %v\n", err)
	} else {
		fmt.Printf("Transaction result2: %v\n", result)
	}
	// 下面测试使用独立事务开启，暂存点，自动回滚
	if result, err := db.BeginTranAutoRoll(func(tx *TxWrapper) (int64, error) {
		row, err := tx.ExecSQL("INSERT INTO go_table (id, name) VALUES (:id, :name)", id)

		return row, err
	}); err != nil {
		fmt.Printf("Transaction result2: %v\n", err)
	} else {
		fmt.Printf("Transaction result2: %v\n", result)
	}

	//异步执行事务
	results := db.ExecuteAsync(func(tx *TxWrapper) AsyncResult {
		var rows int64
		tx.ExecSQL("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE", nil)
		if err := tx.QueryCollect(&users, "SELECT userid, username FROM sendMsg_users(nolock) WHERE userid <= @userid", sql.Named("userid", 308)); err != nil {
			return AsyncResult{Error: err}
		}

		// 插入操作
		row, err := tx.ExecSQL("INSERT INTO go_table (id, name) VALUES (:id, :name)", usermap)
		if err != nil {
			return AsyncResult{Error: err}
		} else {
			rows += row
		}
		// 插入操作
		row, err = tx.ExecSQL("INSERT INTO go_table (id, name) VALUES (:id, :name)", usermap)
		if err != nil {
			return AsyncResult{Error: err}
		} else {
			rows += row
		}

		//删除数据(多条)ExecWithTran
		if row, err := tx.ExecSQL("DELETE FROM dbo.go_table where id = 1", nil); err != nil {
			return AsyncResult{Error: err}
		} else {
			rows += row
		}
		tx.Commit()
		return AsyncResult{ExecRowsAffected: rows, QueryResult: users}
	})
	fmt.Println("waiting for results...")

	result := <-results
	if result.Error != nil {
		fmt.Printf("Error executing transaction: %v\n", result.Error)
	} else {
		fmt.Printf("Rows affected1: %d\n", result.ExecRowsAffected)
		fmt.Printf("Query result: %v\n", result.QueryResult)
	}

	results = db.ExecuteAsyncNoTx(func() AsyncResult {
		var rows int64
		start := time.Now()
		tx, err = db.BeginTran() // 开启非事务

		if row, err := tx.ExecSQL("INSERT INTO go_table (id, name) VALUES (:id, :name)", ids); err == nil {
			rows += row
		}
		// //删除数据(多条)ExecWithTran
		// if row, err := tx.ExecSQL("DELETE FROM dbo.go_table where id = 2", nil); err != nil {
		// 	return AsyncResult{Error: err}
		// } else {
		// 	rows += row
		// }
		tx.Commit()
		end := time.Since(start).Milliseconds()
		logrus.Infof("execute time:%dms", end)
		return AsyncResult{ExecRowsAffected: rows}

	})

	// 直接读取结果而不使用遍历
	result = <-results
	if result.Error != nil {
		fmt.Printf("Error executing transaction: %v\n", result.Error)
	} else {
		fmt.Printf("Rows affected2: %d\n", result.ExecRowsAffected)
	}

	//判断表名是否存在
	exist, err := db.TableExists("go_table")
	if err != nil {
		fmt.Printf("Failed to check table exist: %v\n", err)
	} else {
		fmt.Printf("Table exist: %v\n", exist)
	}
	//判断列名是否存在
	exist, err = db.ColumnExists("go_table", "id")
	if err != nil {
		fmt.Printf("Failed to check column exist: %v\n", err)
	} else {
		fmt.Printf("Column exist: %v\n", exist)
	}

	//判断锁是否存在
	tables := []string{"T_WF_PROCDEF", "T_WF_PROCINST"}
	for _, table := range tables {
		exist, err = db.LocksExists(table)
		if err != nil {
			fmt.Printf("Failed to check column exist: %v\n", err)
		} else {
			fmt.Printf("locks exist: %v\n", exist)
		}
	}

	results = db.ExecuteAsyncNoTx(func() AsyncResult {
		// 查询结果集并映射结构体
		var users []User
		if err := db.QueryCollect(&users, "SELECT userid,username FROM sendMsg_users where userid <= @userid",
			sql.Named("userid", 308)); err != nil {
			fmt.Printf("[QueryValue]查询用户ID时出错: %v\n", err)
		} else {
			fmt.Printf("[QueryValue]获取到的用户列表为: %v\n", users)
		}

		return AsyncResult{QueryResult: users}

	})

	// 直接读取结果而不使用遍历
	result = <-results
	if result.Error != nil {
		fmt.Printf("Error executing transaction: %v\n", result.Error)
	} else {
		fmt.Printf("Rows affected2: %+v\n", result.QueryResult)
	}
}

//#endregion mssqldemo
