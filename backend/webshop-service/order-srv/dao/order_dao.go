package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"webshop-service/order-srv/common/global"
	"webshop-service/order-srv/model"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
在操作mysql语句时，有时需要通过affected_rows来判断语句执行的情况。
例如在事务操作中，就可以通过affected_rows来判断事务是否执行成功，以进一步执行事务的提交或者回滚操作。
对于SELECT操作，mysql_affected_rows()等价于mysql_num_rows()，即查询结果的行数，但是显示使用mysql_num_rows()更加合适。
因此mysql_affected_rows()一般用来在DELETE, INSERT , REPLACE , UPDATE语句执行完成之后判断数据表中变化的行数（如果数据表没有变化，则行数为0）。
DELETE语句执行成功，返回删除的行数，INSERT INTO TABLE VALUES 或者 INSERT INTO TABLES SET 都是返回插入成功的行数，这些是比较明确的。
UPDATE语句执行成功时，则有可能也为0。如果要更新的值与原来的值相同，则affected_rows为0；否则，为更新的行数。
INSERT INTO TABLE VALUES 或者 INSERT INTO TABLES SET 都是返回插入成功的行数，插入成功则返回1，否则返回0 。
INSERT INTO TABLE VALUES … ON DUPLICATE KEY UPDATE … 语句执行成功后，则会有3种情况，当不存在唯一索引冲突时，执行INSERT操作，affected_rows结果为1；当存在主键冲突时，执行UPDATE操作，如果要更新的值与原来的相同，则affected_rows为0，否则为2。
REPLACE INTO TABLE VALUES执行成功 ，如果没有存在唯一索引的冲突，则与INSERT操作没有什么区别affected_rows为1 ；如果存在主键冲突，则会DELETE再INSERT，所以affected_rows的值为2 。
INSERT INTO TABLE VALUES … ON DUPLICATE KEY UPDATE ，当存在唯一索引重复，并成功更新数据之后，受到影响的行数实际上是1，但是affected_rows的值为2。这个数值是mysql手册上规定的，个人猜测应该是因为该语句直接INSERT操作时的affected_rows是1，为了区分两种情况。
*/

//defer是go语言中的关键字，延迟指定函数的执行。通常在资源释放、连接关闭、函数结束时调用。多个defer为堆栈结构，先进后出，也就是先进的后执行。defer可用于异常抛出后的处理。
//panic是go语言中的内置函数，抛出异常(类似java中的throw)
//recover() 是go语言中的内置函数，获取异常(类似java中的catch)，多次调用时，只有第一次能获取值。

//增加订单所属商品出库历史表
func AddInventoryHistory(invHistory *model.InventoryHistory) (err error) {

	defer func() {
		if p := recover(); p != nil {
			errmsg := p.(string)
			err = errors.New(errmsg)

			debug.PrintStack() //堆栈的异常调用信息
		}
	}()

	var sys_err error = nil

	sqlStr := "INSERT inventory_history (order_sn, order_inv_detail, add_time) VALUES (?, ?, NOW())"
	sqlResult, sys_err := global.DBMgr.DB.Exec(sqlStr, invHistory.OrderSn, invHistory.OrderInvDetail)
	if sys_err != nil {
		// 查询数据库出错
		return status.Errorf(codes.Internal, "内部错误[DB.Exec] %s", sys_err.Error())
	}

	var affected int64
	affected, err = sqlResult.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		//Insert未执行
		return errors.New("操作未执行")
	}

	return nil
}

// 提交事务
func CommitTransaction(tx *sqlx.Tx) error {
	return tx.Commit()
}

// 回滚事务
func RollbackTransaction(tx *sqlx.Tx) error {
	return tx.Rollback()
}

// 设置事务隔离级别，只对当前会话有效
func SetTransaction() {

}

func TransactionExec(stmt *sqlx.Stmt, args ...interface{}) error {

	// 添加数据 Exec、MustExec
	// MustExec遇到错误的时候直接抛出一个panic错误，程序就退出了；
	// Exec是将错误和执行结果一起返回，由我们自己处理错误。 推荐使用！
	sqlResult, sys_err := stmt.Exec(args...)
	if sys_err != nil {
		// 查询数据库出错
		return status.Errorf(codes.Internal, "内部错误[DB.Exec] %s", sys_err.Error())
	}

	var affected int64
	affected, err := sqlResult.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		//Insert未执行
		return errors.New("操作未执行")
	}
	return nil
}

func TransactionGet(stmt *sqlx.Stmt, dest interface{}, args ...interface{}) error {

	// Get、QueryRowx: 查询一条数据
	// QueryRowx可以指定到不同的数据类型中

	// Select、Queryx：查询多条数据
	// Queryx可以指定到不同的数据类型中

	//Stmt.Get] sql: converting argument $1 type: unsupported type []interface {}, a slice of interface
	//字面意思是sqlx在解析两个占位符并试图填入参数时，第一个参数类型是空指针的切片，而预期是args这个可变参数中的第一个。
	//当...Type作为参数时，本质上函数会把参数转化成一个Type类型的切片，于是在上述代码中，Service层调以可变参数形式传入两个参数，
	//在Insert中的args就已经是[]interface{}类型了，如果直接把args作为func (db *DB) Exec(query string, args ...interface{}) (Result, error)的参数，
	//对于Exec来说，收到的args就只有一个长度为1的切片，其元素类型为[]interface{}，于是就有了上述的报错，解决办法很简单，
	//就是在一个slice后加上...，这样就能把它拆包成一个可变参数的形式传入函数。
	err := stmt.Get(dest, args...)
	if err != nil && err != sql.ErrNoRows {
		// 查询数据库出错
		return status.Errorf(codes.Internal, "内部错误[Stmt.Get] %s", err.Error())
	}

	if err == sql.ErrNoRows {
		return err
	}

	return nil
	/*
		var goods_id uint32
		var stocks uint32
		var version uint32
		var freeze uint32
		err := stmt.QueryRowx(args...).Scan(&goods_id, &stocks, &version, &freeze)
		if err != nil {
			// 查询数据库出错
			return status.Errorf(codes.Internal, "内部错误[Stmt.Get] %s", err.Error())
		}
	*/

}

func TransactionSelect(stmt *sqlx.Stmt, dest interface{}, args ...interface{}) error {

	// Get、QueryRowx: 查询一条数据
	// QueryRowx可以指定到不同的数据类型中

	// Select、Queryx：查询多条数据
	// Queryx可以指定到不同的数据类型中

	//Stmt.Get] sql: converting argument $1 type: unsupported type []interface {}, a slice of interface
	//字面意思是sqlx在解析两个占位符并试图填入参数时，第一个参数类型是空指针的切片，而预期是args这个可变参数中的第一个。
	//当...Type作为参数时，本质上函数会把参数转化成一个Type类型的切片，于是在上述代码中，Service层调以可变参数形式传入两个参数，
	//在Insert中的args就已经是[]interface{}类型了，如果直接把args作为func (db *DB) Exec(query string, args ...interface{}) (Result, error)的参数，
	//对于Exec来说，收到的args就只有一个长度为1的切片，其元素类型为[]interface{}，于是就有了上述的报错，解决办法很简单，
	//就是在一个slice后加上...，这样就能把它拆包成一个可变参数的形式传入函数。
	err := stmt.Select(dest, args...)
	if err != nil && err != sql.ErrNoRows {
		// 查询数据库出错
		return status.Errorf(codes.Internal, "内部错误[Stmt.Select] %s", err.Error())
	}

	if err == sql.ErrNoRows {
		return err
	}

	return nil
	/*
		var goods_id uint32
		var stocks uint32
		var version uint32
		var freeze uint32
		err := stmt.QueryRowx(args...).Scan(&goods_id, &stocks, &version, &freeze)
		if err != nil {
			// 查询数据库出错
			return status.Errorf(codes.Internal, "内部错误[Stmt.Get] %s", err.Error())
		}
	*/

}

func TransactionPrepare(tx *sqlx.Tx, sql string) (*sql.Stmt, error) {
	return tx.Prepare(sql)
}

func SimplePanic(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Println(file, line, err)
		//runtime.Goexit()
	}
}

/*
func logRollback(tx *sql.Tx) {
	err := tx.Rollback()
	if err != tx.ErrTxDone && err != nil {
		log.Error(err.Error())
	}
}
*/
