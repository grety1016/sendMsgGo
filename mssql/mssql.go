package mssql

import (
	"database/sql"
	"fmt"
	"sendmsg/logger"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// 数据库连接和连接池的配置参数
type DBConfig struct {
	ConnString      string        // 数据库连接字符串，用于指定数据库连接的具体信息
	MaxOpenConns    int           // 最大打开连接数，控制连接池中允许打开的最大数据库连接数
	MaxIdleConns    int           // 最大空闲连接数，控制连接池中保持空闲的最大数据库连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期，控制连接池中连接的最长生命周期
	ConnMaxIdleTime time.Duration // 连接最大空闲时间，控制连接池中连接的最大空闲时间
}

// AsyncResult 包装 SQL(ExecuteAsync\ExecuteAsyncNoTx) 执行结果
type AsyncResult struct {
	ExecRowsAffected int64
	QueryResult      interface{}
	Error            error
}

// DBWrapper包装sqlx.DB 和相关配置
type DBWrapper struct {
	db     *sqlx.DB
	dbName string
}

// TxWrapper包装sqlx.Tx,*DBWrapper
type TxWrapper struct {
	tx          *sqlx.Tx
	dbWrapper   *DBWrapper
	isCommitted int32
	txCount     int32
}

// 初始化数据库连接并配置连接池参数
func InitDB(config DBConfig) (*DBWrapper, error) {
	db, err := sqlx.Open("sqlserver", config.ConnString)
	if err != nil {
		logrus.Errorf("Error opening database: %v", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	// 测试数据库连接
	if err = db.Ping(); err != nil {
		logrus.Errorf("Error pinging database: %v", err)
		return nil, err
	}

	// 配置连接池参数
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 创建 DBWrapper 并设置数据库名称
	dbWrapper := &DBWrapper{
		db:     db,
		dbName: GetDatabaseName(config.ConnString),
	}

	return dbWrapper, nil
}

// 获取数据库名称
func GetDatabaseName(connStr string) string {
	params := strings.Split(connStr, ";")
	for _, param := range params {
		keyValue := strings.SplitN(param, "=", 2)
		if len(keyValue) == 2 && strings.TrimSpace(strings.ToLower(keyValue[0])) == "database" {
			return strings.TrimSpace(keyValue[1])
		}
	}
	return ""
}

// 外部获取数据库名
func (db *DBWrapper) GetDBName() string {
	return db.dbName
}

// InitTX 初始化事务包装器
func (dbWrapper *DBWrapper) initTX() (*TxWrapper, error) {
	logrus.Infof("[DB] @%s - executing, sql: BeginTran { Begin transaction }", dbWrapper.dbName)
	start := time.Now() // 执行开始时间
	tx, err := dbWrapper.db.Beginx()
	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - executed:%dms, sql: BeginTran { Failed to begin transaction, Error: %v }", dbWrapper.dbName, elapsedMillis, err)
		return nil, err
	}
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - executed:%dms, sql: BeginTran { Transaction begun successfully }", dbWrapper.dbName, elapsedMillis)
	return &TxWrapper{tx: tx, dbWrapper: dbWrapper, isCommitted: 0, txCount: 1}, nil
}

// Execer 接口
type Execer interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
	QueryValue(query string, args ...interface{}) (interface{}, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	DBName() string
}

func (dbWrapper *DBWrapper) NamedExec(query string, arg interface{}) (sql.Result, error) {
	return dbWrapper.db.NamedExec(query, arg)
}
func (dbWrapper *DBWrapper) Get(dest interface{}, query string, args ...interface{}) error {
	return dbWrapper.db.Get(dest, query, args...)
}

func (dbWrapper *DBWrapper) Select(dest interface{}, query string, args ...interface{}) error {
	return dbWrapper.db.Select(dest, query, args...)
}

func (dbWrapper *DBWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return dbWrapper.db.Exec(query, args...)
}

func (dbWrapper *DBWrapper) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return dbWrapper.db.Queryx(query, args...)
}

func (dbWrapper *DBWrapper) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return dbWrapper.db.QueryRowx(query, args...)
}

func (dbWrapper *DBWrapper) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	return dbWrapper.db.NamedQuery(query, arg)
}

func (dbWrapper *DBWrapper) BindNamed(query string, arg interface{}) (string, []interface{}, error) {
	return dbWrapper.db.BindNamed(query, arg)
}

func (dbWrapper *DBWrapper) DBName() string {
	return dbWrapper.dbName
}

func (txWrapper *TxWrapper) NamedExec(query string, arg interface{}) (sql.Result, error) {
	return txWrapper.tx.NamedExec(query, arg)
}

func (txWrapper *TxWrapper) Get(dest interface{}, query string, args ...interface{}) error {
	return txWrapper.tx.Get(dest, query, args...)
}

func (txWrapper *TxWrapper) Select(dest interface{}, query string, args ...interface{}) error {
	return txWrapper.tx.Select(dest, query, args...)
}

func (txWrapper *TxWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return txWrapper.tx.Exec(query, args...)
}

func (txWrapper *TxWrapper) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return txWrapper.tx.Queryx(query, args...)
}

func (txWrapper *TxWrapper) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return txWrapper.tx.QueryRowx(query, args...)
}

func (txWrapper *TxWrapper) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	return txWrapper.tx.NamedQuery(query, arg)
}

func (txWrapper *TxWrapper) BindNamed(query string, arg interface{}) (string, []interface{}, error) {
	return txWrapper.tx.BindNamed(query, arg)
}

func (txWrapper *TxWrapper) DBName() string {
	return txWrapper.dbWrapper.dbName
}

// 为DBWrapper实现 Execer 接口（实现 sqlx.Ext 接口的所有方法）

var _ Execer = (*DBWrapper)(nil)
var _ Execer = (*TxWrapper)(nil)

// QueryValue 函数用于查询返回单个值（仅限单列查询，多列查询请使用QueryCollect，否则报错），使用前需断言，空行异常需判断,否则报错
func (txWrapper *TxWrapper) QueryValue(query string, args ...interface{}) (interface{}, error) {
	return queryValue(txWrapper, query, args...)
}

// QueryValue 函数用于查询返回单个值（仅限单列查询，多列查询请使用QueryCollect，否则报错），使用前需断言，空行异常需判断,否则报错
func (dbWrapper *DBWrapper) QueryValue(query string, args ...interface{}) (interface{}, error) {
	return queryValue(dbWrapper, query, args...)
}

// QueryValue 函数用于查询返回单个值（仅限单列查询，多列查询请使用QueryCollect，否则报错），使用前需断言，空行异常需判断,否则报错
func queryValue(exec Execer, query string, args ...interface{}) (interface{}, error) {
	var value interface{}
	dbName := exec.DBName()

	// 记录开始查询语句
	logrus.Infof("[DB] @%s - executing, sql: QueryValue { sql: \"%s\", Params: [%v] }", dbName, query, logger.FormatNamedArgs(args))
	start := time.Now() // 执行开始时间

	// 执行查询并获取单个值
	err := exec.Get(&value, query, args...)
	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - elapsed:%dms, sql: QueryValue { sql: \"%s\", Params: [%v], Error: %v }", dbName, elapsedMillis, query, logger.FormatNamedArgs(args), err)
		return value, err
	}
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - elapsed:%dms, sql: QueryValue { sql: \"%s\", Params: [%v] }", dbName, elapsedMillis, query, logger.FormatNamedArgs(args))
	// 记录结束查询语句

	return value, nil
}

// QueryCollect 查询数据结果集并将其映射到结构体切片中
func (dbWrapper *DBWrapper) QueryCollect(dest interface{}, query string, args ...interface{}) error {
	return queryCollect(dbWrapper, dest, query, args...)

}

// QueryCollect 查询数据结果集并将其映射到结构体切片中
func (txWrapper *TxWrapper) QueryCollect(dest interface{}, query string, args ...interface{}) error {
	return queryCollect(txWrapper, dest, query, args...)

}

// QueryCollect 查询数据结果集并将其映射到结构体切片中
func queryCollect(exec Execer, dest interface{}, query string, args ...interface{}) error {
	dbName := exec.DBName()

	// 记录开始查询语句
	logrus.Infof("[DB] @%s - executing, sql: QueryCollect { sql: \"%s\", Params: [%v] }", dbName, query, logger.FormatNamedArgs(args))
	start := time.Now() // 执行开始时间

	// 执行查询数据
	err := exec.Select(dest, query, args...)
	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - elapsed:%dms, sql: QueryCollect { sql: \"%s\", Params: [%v], Error: %v }", dbName, elapsedMillis, query, logger.FormatNamedArgs(args), err)
		return err
	}
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - elapsed:%dms, sql: QueryCollect { sql: \"%s\", Params: [%v] }", dbName, elapsedMillis, query, logger.FormatNamedArgs(args))
	// 记录结束查询语句

	return nil
}

// 执行插入、更新和删除操作，并返回受影响的行数，仅支持批量插入操作(非事务版本)
func (txWrapper *TxWrapper) ExecSQL(query string, args interface{}) (int64, error) {
	return exec(txWrapper, query, args)
}

// ExecSql 执行插入、更新和删除操作，并返回受影响的行数，仅支持批量插入操作(非事务版本)
func (dbWrapper *DBWrapper) ExecSQL(query string, args interface{}) (int64, error) {
	return exec(dbWrapper, query, args)
}

// Exec 函数用于执行任意 SQL 语句，包括插入、更新和删除操作，仅支持(插入Insert)批量操作(非事务版本)
func exec(exec Execer, query string, args interface{}) (int64, error) {
	dbName := exec.DBName()

	// 记录开始执行操作
	logrus.Infof("[DB] @%s - executing, sql: Exec { sql: \"%s\", Params: [%+v] }", dbName, query, args)
	start := time.Now()

	var totalRowsAffected int64
	var err error
	defer func() {
		elapsedMillis := time.Since(start).Milliseconds()
		if err != nil {
			logrus.Errorf("[DB] @%s - elapsed:%dms, sql: Exec { sql: \"%s\", Params: [%+v], Error: %v }", dbName, elapsedMillis, query, args, err)
		} else {
			logrus.Infof("[DB] @%s - elapsed:%dms, sql: Exec { sql: \"%s\", Params: [%+v], RowsAffected: %d }", dbName, elapsedMillis, query, args, totalRowsAffected)
		}
	}()

	// 处理 args 为 nil 的情况
	if args == nil {
		args = struct{}{}
	}

	// 直接调用 NamedExec 处理批量操作或单个元素操作
	res, execErr := exec.NamedExec(query, args)
	if execErr != nil {
		err = execErr
		return 0, err
	}

	// 获取受影响的行数
	rowsAffected, execErr := res.RowsAffected()
	if execErr != nil {
		err = execErr
		return 0, err
	}
	totalRowsAffected = rowsAffected

	return totalRowsAffected, nil
}

// ExecSQLWithTran 执行插入、更新和删除操作，并返回受影响的行数，仅支持批量插入操作(事务版本)
func (dbWrapper *DBWrapper) ExecSQLWithTran(query string, args interface{}) (int64, error) {
	return execWithTran(dbWrapper, query, args)
}

// execWithTran 执行插入、更新和删除操作，并返回受影响的行数，仅支持批量插入操作(事务版本)
func execWithTran(dbWrapper *DBWrapper, query string, args interface{}) (int64, error) {
	// 记录开始执行操作
	start := time.Now()
	// 开启事务
	tx, err := dbWrapper.db.Beginx()
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - executed:%dms, sql: ExecSQLWithTran { Transaction begun successfully }", dbWrapper.dbName, elapsedMillis)

	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - executed:%dms, sql: ExecSQLWithTran { Failed to begin transaction, Error: %v }", dbWrapper.dbName, elapsedMillis, err)
		return 0, err
	}

	defer func() {
		start := time.Now() // 执行开始时间
		if p := recover(); p != nil {
			tx.Rollback() // 错误处理或发生panic时回滚事务
			elapsedMillis = time.Since(start).Milliseconds()
			logrus.Errorf("[DB] @%s - executed:%dms, sql: ExecSQLWithTran { Failed to Rollback, Error: %v }", dbWrapper.dbName, elapsedMillis, err)
		} else if err != nil {
			tx.Rollback() // 错误处理
			elapsedMillis = time.Since(start).Milliseconds()
			logrus.Errorf("[DB] @%s - executed:%dms, sql: ExecSQLWithTran { Failed to Rollback, Error: %v }", dbWrapper.dbName, elapsedMillis, err)
		} else {
			err = tx.Commit() // 成功时提交事务
			if err != nil {
				elapsedMillis = time.Since(start).Milliseconds()
				logrus.Errorf("[DB] @%s - executed:%dms, sql: ExecSQLWithTran { Failed to commit, Error: %v }", dbWrapper.dbName, elapsedMillis, err)
			}
			elapsedMillis = time.Since(start).Milliseconds()
			logrus.Infof("[DB] @%s - executed:%dms, sql: ExecSQLWithTran { Transaction committed successfully }", dbWrapper.dbName, elapsedMillis)
		}
	}()

	if args == nil {
		args = struct{}{}
	}
	logrus.Infof("[DB] @%s - executing, sql: ExecSQLWithTran { sql: \"%s\", Params: [%+v] }", dbWrapper.dbName, query, logger.FormatNamedArgs(args))
	start = time.Now() // 执行开始时间
	// 执行 NamedExec 并获取受影响的行数
	res, execErr := tx.NamedExec(query, args)
	if execErr != nil {
		elapsedMillis = time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - elapsed:%dms, sql: ExecSQLWithTran { sql: \"%s\", Params: [%+v], Error: %v }", dbWrapper.dbName, elapsedMillis, query, logger.FormatNamedArgs(args), execErr)
		err = execErr
		return 0, err
	}

	rowsAffected, execErr := res.RowsAffected()
	if execErr != nil {
		elapsedMillis = time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - elapsed:%dms, sql: ExecSQLWithTran { sql: \"%s\", Params: [%+v], Error: %v }", dbWrapper.dbName, elapsedMillis, query, logger.FormatNamedArgs(args), execErr)
		err = execErr
		return 0, err
	}

	totalRowsAffected := rowsAffected
	elapsedMillis = time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - elapsed:%dms, sql: ExecSQLWithTran { sql: \"%s\", Params: [%+v], RowsAffected: %d }", dbWrapper.dbName, elapsedMillis, query, logger.FormatNamedArgs(args), totalRowsAffected)

	return totalRowsAffected, nil
}

// ExecuteAsync 异步执行 SQL 语句，支持事务处理，内部所有操作均以*TxWrapper对象进行，否则会有不可预知的行为；
// ExecuteAsync内部已经开启事务，请勿在调用处的闭包中再次使用携带事务的函数，因为事务实例是独立的，重复开启不能保证事务一致性
// ExecuteAsync 异步执行事务并增加死锁重试逻辑，该函数仅限于同步环境调用，异步环境可能造成死锁或其他不可预知的行为；
func (dbWrapper *DBWrapper) ExecuteAsync(fn func(*TxWrapper) AsyncResult) chan AsyncResult {
	results := make(chan AsyncResult, 1) // 创建带缓冲的通道，避免阻塞
	var resultsClosed bool
	var mu sync.Mutex // 保护 resultsClosed 的互斥锁

	go func() {
		start := time.Now() // 执行开始时间
		defer func() {
			if p := recover(); p != nil {
				elapsedMillis := time.Since(start).Milliseconds()
				logrus.Errorf("[DB] @%s - executed:%dms, sql: ExecuteAsync { Execution failed due to panic: %v }", dbWrapper.dbName, elapsedMillis, p)
				mu.Lock()
				if !resultsClosed {
					results <- AsyncResult{Error: fmt.Errorf("execution failed due to panic: %v", p)}
					close(results)
					resultsClosed = true
				}
				mu.Unlock()
			}
			mu.Lock()
			if !resultsClosed {
				close(results)
				resultsClosed = true
			}
			mu.Unlock()
		}()

		const maxRetries = 5                        // 定义最大重试次数为5次
		const initialDelay = 100 * time.Millisecond // 初始等待时间为100毫秒
		const backoffFactor = 2                     // 退避因子为2

		for i := 0; i < maxRetries; i++ {
			// 开启事务
			txWrapper, err := dbWrapper.initTX()
			if err != nil {
				mu.Lock()
				if !resultsClosed {
					results <- AsyncResult{Error: err}
					close(results)
					resultsClosed = true
				}
				mu.Unlock()
				return
			}

			var result AsyncResult
			var committed bool

			defer func() {
				if p := recover(); p != nil {
					txWrapper.Rollback()
					elapsedMillis := time.Since(start).Milliseconds()
					logrus.Errorf("[DB] @%s - executed:%dms, sql: ExecuteAsync { Transaction rolled back due to panic: %v }", dbWrapper.dbName, elapsedMillis, p)
					mu.Lock()
					if !committed && !resultsClosed {
						results <- AsyncResult{Error: fmt.Errorf("transaction rolled back due to panic: %v", p)}
					}
					mu.Unlock()
				} else if err != nil {
					txWrapper.Rollback()
					mu.Lock()
					if !committed && !resultsClosed {
						results <- AsyncResult{Error: err}
					}
					mu.Unlock()
				} else {
					commitErr := txWrapper.Commit()
					if commitErr != nil {
						elapsedMillis := time.Since(start).Milliseconds()
						logrus.Errorf("[DB] @%s - executed:%dms, sql: ExecuteAsync { Failed to commit transaction: %v }", dbWrapper.dbName, elapsedMillis, commitErr)
						mu.Lock()
						if !committed && !resultsClosed {
							results <- AsyncResult{Error: commitErr}
						}
						mu.Unlock()
					} else {
						atomic.StoreInt32(&txWrapper.isCommitted, 1) // 标记事务已提交
						committed = true
					}
				}
				mu.Lock()
				if !resultsClosed {
					close(results)
					resultsClosed = true
				}
				mu.Unlock()
			}()

			result = fn(txWrapper)
			if result.Error != nil {
				if sqlErr, ok := result.Error.(mssql.Error); ok && sqlErr.Number == 1205 { // 检测死锁错误号
					elapsedMillis := time.Since(start).Milliseconds()
					logrus.Warnf("[DB] @%s - executed:%dms, sql: ExecuteAsync { Deadlock detected, retrying transaction... Attempt %d/%d }", dbWrapper.dbName, elapsedMillis, i+1, maxRetries)
					// 计算等待时间
					waitTime := initialDelay * time.Duration(backoffFactor^i)
					time.Sleep(waitTime) // 增加退避时间
					continue
				}
				err = result.Error
			} else {
				committed = true
			}
			mu.Lock()
			if !resultsClosed {
				results <- result
				close(results)
				resultsClosed = true
			}
			mu.Unlock()
			return
		}
	}()

	return results
}

// ExecuteAsyncNoTx 异步执行 SQL 语句，不支持事务处理，闭包内部可自行创建事务对象进行操作，事务提交回滚等用户自行控制
// ExecuteAsyncNoTx 异步执行 SQL 语句，不支持事务处理
func (dbWrapper *DBWrapper) ExecuteAsyncNoTx(fn func() AsyncResult) chan AsyncResult {
	results := make(chan AsyncResult, 1) // 创建带缓冲的通道，避免阻塞
	var resultsClosed bool
	var mu sync.Mutex // 保护 resultsClosed 的互斥锁

	go func() {
		defer func() {
			if p := recover(); p != nil {
				logrus.Errorf("[DB] @%s - executed, sql: ExecuteAsyncNoTx { Execution failed due to panic: %v }", dbWrapper.dbName, p)
				mu.Lock()
				if !resultsClosed {
					results <- AsyncResult{Error: fmt.Errorf("execution failed due to panic: %v", p)}
					close(results)
					resultsClosed = true
				}
				mu.Unlock()
			}
			mu.Lock()
			if !resultsClosed {
				close(results)
				resultsClosed = true
			}
			mu.Unlock()
		}()

		const maxRetries = 5                        // 定义最大重试次数为5次
		const initialDelay = 100 * time.Millisecond // 初始等待时间为100毫秒
		const backoffFactor = 2                     // 退避因子为2

		var result AsyncResult

		for i := 0; i < maxRetries; i++ {
			// 记录操作开始时间
			start := time.Now()

			// 执行异步任务
			result = fn()

			// 记录操作结束时间
			elapsed := time.Since(start)

			if result.Error != nil {
				if sqlErr, ok := result.Error.(mssql.Error); ok && sqlErr.Number == 1205 { // 检测死锁错误号
					logrus.Warnf("[DB] @%s - executed, sql: ExecuteAsyncNoTx { Deadlock detected, retrying transaction... Attempt %d/%d }", dbWrapper.dbName, i+1, maxRetries)
					// 计算等待时间
					// 指数退避
					waitTime := initialDelay * time.Duration(backoffFactor^i)
					// 确保退避时间大于最短事务时间
					if waitTime < elapsed {
						waitTime = elapsed
					}
					time.Sleep(waitTime) // 增加退避时间
					continue
				}
				break
			}
			break
		}

		mu.Lock()
		if !resultsClosed {
			results <- result
			close(results)
			resultsClosed = true
		}
		mu.Unlock()
	}()

	return results
}

// BeginTran 开启事务（全部操作成功自动提交事务，过程失败或异常自动回滚事务）
// 该函数仅限于同步环境调用，异步环境可能造成死锁或其他不可预知的行为；
func (dbWrapper *DBWrapper) BeginTranAutoCommit(fn func(*TxWrapper) (int64, error)) (int64, error) {
	var results int64
	txWrapper, err := dbWrapper.initTX()
	if err != nil {
		return results, err
	}
	defer func() {
		if p := recover(); p != nil {
			txWrapper.Rollback() // 回滚事务
			logrus.Errorf("[DB] @%s - executed, sql: BeginTranAutoCommit { Transaction rolled back due to panic: %v }", dbWrapper.dbName, p)
			err = fmt.Errorf("transaction rolled back due to panic: %v", p) // 将 panic 信息转换为错误
		} else if err != nil {
			rollbackErr := txWrapper.Rollback() // 回滚事务
			if rollbackErr != nil {
				logrus.Errorf("[DB] @%s - executed, sql: BeginTranAutoCommit { Failed to rollback transaction: %v }", dbWrapper.dbName, rollbackErr)
				err = fmt.Errorf("transaction rolled back due to error: %v, rollback error: %v", err, rollbackErr)
			} else {
				logrus.Errorf("[DB] @%s - executed, sql: BeginTranAutoCommit { Transaction rollbaced,reason: %v }", dbWrapper.dbName, err)
				err = fmt.Errorf("transaction rollbaced,reason: %v", err)
			}
		} else {
			commitErr := txWrapper.Commit() // 提交事务
			if commitErr != nil {
				logrus.Errorf("[DB] @%s - executed, sql: BeginTranAutoCommit { Failed to commit transaction: %v }", dbWrapper.dbName, commitErr)
				err = commitErr
			}
		}
	}()

	results, err = fn(txWrapper)
	return results, err
}

// BeginTranAutoRoll 开启事务，非主动提交事务，自动回滚事务(适用于存在保存点的场景不会自动提交保持事务原子性)
func (dbWrapper *DBWrapper) BeginTranAutoRoll(fn func(*TxWrapper) (int64, error)) (int64, error) {
	var results int64
	txWrapper, err := dbWrapper.initTX()
	if err != nil {
		return results, err
	}
	defer func() {
		if p := recover(); p != nil {
			txWrapper.Rollback() // 回滚事务
			logrus.Errorf("[DB] @%s - executed, sql: BeginTranAutoRoll { Transaction rolled back due to panic: %v }", dbWrapper.dbName, p)
			err = fmt.Errorf("transaction rolled back due to panic: %v", p) // 将 panic 信息转换为错误
		} else if atomic.LoadInt32(&txWrapper.isCommitted) == 0 {
			rollbackErr := txWrapper.Rollback() // 回滚事务
			if rollbackErr != nil {
				logrus.Errorf("[DB] @%s - executed, sql: BeginTranAutoRoll { Failed to rollback transaction: %v }", dbWrapper.dbName, rollbackErr)

			} else {
				logrus.Infof("[DB] @%s - executed, sql: BeginTranAutoRoll { Transaction rolled back automatically as it was not committed }", dbWrapper.dbName)
			}
		}
	}()

	results, err = fn(txWrapper)
	return results, err
}

// BeginTran 开启事务
func (dbWrapper *DBWrapper) BeginTran() (*TxWrapper, error) {
	txWrapper, err := dbWrapper.initTX()
	if err != nil {
		return nil, err
	}

	return txWrapper, nil
}

// Commit 提交事务
func (txWrapper *TxWrapper) Commit() error {
	start := time.Now() // 执行开始时间
	if !txWrapper.HasTran() {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Infof("[DB] @%s - executed:%dms, sql: HasTran(emit-%s) { No active transaction }", txWrapper.dbWrapper.dbName, elapsedMillis, "Commit")
		return nil
	} else {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Infof("[DB] @%s - executed:%dms, sql: HasTran(emit-%s) { Transaction is active }", txWrapper.dbWrapper.dbName, elapsedMillis, "Commit")
	}

	err := txWrapper.tx.Commit()
	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - executed:%dms, sql: Commit { Failed to commit, Error: %v }", txWrapper.dbWrapper.dbName, elapsedMillis, err)
		return err
	}
	atomic.AddInt32(&txWrapper.txCount, -1)
	atomic.StoreInt32(&txWrapper.isCommitted, 1)
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - executed:%dms, sql: Commit { Transaction committed successfully }", txWrapper.dbWrapper.dbName, elapsedMillis)

	return nil
}

// Rollback 回滚事务
func (txWrapper *TxWrapper) Rollback() error {
	start := time.Now() // 执行开始时间
	if !txWrapper.HasTran() {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Infof("[DB] @%s - executed:%dms, sql: HasTran(emit-%s) { No active transaction }", txWrapper.dbWrapper.dbName, elapsedMillis, "Rollback")
		return nil
	} else {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Infof("[DB] @%s - executed:%dms, sql: HasTran(emit-%s) { Transaction is active }", txWrapper.dbWrapper.dbName, elapsedMillis, "Rollback")
	}

	err := txWrapper.tx.Rollback()
	if err != nil && err != sql.ErrTxDone {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - executed:%dms, sql: Rollback { Failed to rollback, Error: %v }", txWrapper.dbWrapper.dbName, elapsedMillis, err)
	} else {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Infof("[DB] @%s - executed:%dms, sql: Rollback { Transaction rolled back successfully }", txWrapper.dbWrapper.dbName, elapsedMillis)
	}
	// 确保在任何情况下都减少计数器
	atomic.AddInt32(&txWrapper.txCount, -1)

	return err
}

// 判断当前是否存在事务
func (txWrapper *TxWrapper) HasTran() bool {

	active := atomic.LoadInt32(&txWrapper.txCount) > 0

	return active
}

// 设置事务保存点
func (txWrapper *TxWrapper) SaveTran(name string) error {
	start := time.Now() // 执行开始时间
	_, err := txWrapper.tx.Exec(fmt.Sprintf("SAVE TRANSACTION %s", name))
	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - executed:%dms, sql: SaveTran { Failed to create savepoint \"%s\", Error: %v }", txWrapper.dbWrapper.dbName, elapsedMillis, name, err)
		return err
	}
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - executed:%dms, sql: SaveTran { Savepoint \"%s\" created successfully }", txWrapper.dbWrapper.dbName, elapsedMillis, name)
	return nil
}

// 回滚到保存点
func (txWrapper *TxWrapper) RollbackToSave(name string) error {
	start := time.Now() // 执行开始时间
	_, err := txWrapper.tx.Exec(fmt.Sprintf("ROLLBACK TRANSACTION %s", name))
	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - executed:%dms, sql: RollbackToSave { Failed to rollback to savepoint \"%s\", Error: %v }", txWrapper.dbWrapper.dbName, elapsedMillis, name, err)
		return err
	}
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - executed:%dms, sql: RollbackToSave { Transaction rolled back to savepoint \"%s\" successfully }", txWrapper.dbWrapper.dbName, elapsedMillis, name)
	return nil
}

// Close 关闭数据库连接
func (dbWrapper *DBWrapper) Close() error {
	return dbWrapper.db.Close()
}

// Close 方法，用于释放事务资源
func (txWrapper *TxWrapper) Close() error {
	start := time.Now() // 执行开始时间
	if p := recover(); p != nil {
		if txWrapper.HasTran() {
			txWrapper.Rollback()
			elapsedMillis := time.Since(start).Milliseconds()
			logrus.Errorf("[DB] @%s - executed:%dms, sql: Close { transaction rolled back due to panic: %v }", txWrapper.dbWrapper.dbName, elapsedMillis, p)
		}
		return fmt.Errorf("transaction rolled back due to panic: %v", p)
	}

	if !txWrapper.IsCommitted() && txWrapper.HasTran() { // 使用 IsCommitted 和 HasTran 检查事务状态
		txWrapper.Rollback()
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Infof("[DB] @%s - executed:%dms, sql: Close { Transaction was not committed, automatically rolled back }", txWrapper.dbWrapper.dbName, elapsedMillis)
	}
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - executed:%dms, sql: Close { Transaction closed successfully }", txWrapper.dbWrapper.dbName, elapsedMillis)
	return nil
}

// IsCommitted 检查事务是否已提交
func (txWrapper *TxWrapper) IsCommitted() bool {
	return atomic.LoadInt32(&txWrapper.isCommitted) == 1
}

// TableExists 检查指定的表是否存在
func (dbWrapper *DBWrapper) TableExists(tableName string) (bool, error) {
	return tableExists(dbWrapper, tableName)
}

// TableExists 检查指定的表是否存在
func (txWrapper *TxWrapper) TableExists(tableName string) (bool, error) {
	return tableExists(txWrapper, tableName)
}

// TableExists 检查指定的表是否存在
func (dbWrapper *DBWrapper) ColumnExists(tableName string, columnName string) (bool, error) {
	return columnExists(dbWrapper, tableName, columnName)
}

// TableExists 检查指定的表是否存在
func (txWrapper *TxWrapper) ColumnExists(tableName string, columnName string) (bool, error) {
	return columnExists(txWrapper, tableName, columnName)
}

// tableExists 检查指定的表是否存在
func tableExists(exec Execer, tableName string) (bool, error) {
	var exists bool
	dbName := exec.DBName()
	logrus.Infof("[DB] @%s - executing, sql: TableExists { TableExists? }", dbName)

	query := "SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = @p1"
	start := time.Now() // 执行开始时间
	err := exec.QueryRowx(query, tableName).Scan(&exists)
	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - executed:%dms, sql: TableExists { Failed to check if table exists, Error: %v }", dbName, elapsedMillis, err)
		return false, fmt.Errorf("failed to check if table exists: %w", err)
	}
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - executed:%dms, sql: TableExists { TableExists: %v }", dbName, elapsedMillis, exists)

	return exists, nil
}

// columnExists 检查指定表中的字段是否存在
func columnExists(exec Execer, tableName string, columnName string) (bool, error) {
	var exists bool
	dbName := exec.DBName()
	logrus.Infof("[DB] @%s - executing, sql: ColumnExists { ColumnExists? }", dbName)
	query := `
        SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_NAME = @p1 AND COLUMN_NAME = @p2
    `
	start := time.Now() // 执行开始时间
	err := exec.QueryRowx(query, tableName, columnName).Scan(&exists)
	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - executed:%dms, sql: ColumnExists { Failed to check if column exists, Error: %v }", dbName, elapsedMillis, err)
		return false, fmt.Errorf("failed to check if column exists: %w", err)
	}
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - executed:%dms, sql: ColumnExists { ColumnExists: %v }", dbName, elapsedMillis, exists)
	return exists, nil
}

// 执行插入、更新和删除操作，并返回受影响的行数，仅支持批量插入操作(非事务版本)
func (txWrapper *TxWrapper) LocksExists(tableNames string) (bool, error) {
	return lockExists(txWrapper, tableNames)
}

// ExecSql 执行插入、更新和删除操作，并返回受影响的行数，仅支持批量插入操作(非事务版本)
func (dbWrapper *DBWrapper) LocksExists(tableNames string) (bool, error) {
	return lockExists(dbWrapper, tableNames)
}

// lockExists 判断单个表是否存在锁
func lockExists(execer Execer, tableName string) (bool, error) {
	var exists bool
	dbName := execer.DBName()

	query := `
        SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END
        FROM sys.dm_exec_sessions AS s
        JOIN sys.dm_tran_locks AS t ON s.session_id = t.request_session_id
        WHERE t.resource_associated_entity_id = OBJECT_ID(@p1)
    `
	start := time.Now() // 执行开始时间
	err := execer.QueryRowx(query, tableName).Scan(&exists)
	if err != nil {
		elapsedMillis := time.Since(start).Milliseconds()
		logrus.Errorf("[DB] @%s - executed:%dms, sql: LockExists { Failed to check if table %s is locked, Error: %v }", dbName, elapsedMillis, tableName, err)
		return false, fmt.Errorf("failed to check if table [%s] is locked: %w", tableName, err)
	}
	elapsedMillis := time.Since(start).Milliseconds()
	logrus.Infof("[DB] @%s - executed:%dms, sql: LockExists { Table [%s] locked is: %v }", dbName, elapsedMillis, tableName, exists)
	return exists, nil
}

// HACK：表示使用了不推荐或临时的解决方案，可能需要在将来修复。

// BUG：表示代码中存在已知的错误，需要修复。
// FIXME：表示代码中存在已知的错误，需要修复。
// TODO：表示代码中存在未完成的工作，需要完成。
// *如下函数由于SQLX不支持超时取消，所以暂时不使用(仅能取消本地任务执行，但已经发送的查询请求仍然会继续执行)
