package dao

import (
	utils "common/Utils"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"
	"webshop-service/user-srv/common/errno"
	"webshop-service/user-srv/common/global"
	pb "webshop-service/user-srv/proto"

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

// 创建用户
func RegisterUser(createUser *pb.UserRegisterRequest) (lastID uint64, err error) {

	//我们在调用recover的延迟函数中以最合理的方式响应该异常：
	//1. 打印堆栈的异常调用信息和关键的业务信息，以便这些问题保留可见；
	//2. 将异常转换为错误，以便调用者让程序恢复到健康状态并继续安全运行。
	defer func() {
		if p := recover(); p != nil {
			lastID = 0
			errmsg := p.(string)
			err = errors.New(errmsg)

			debug.PrintStack() //堆栈的异常调用信息
		}
	}()

	var sys_err error = nil
	//启动事务

	//判断用户是否存在
	sqlStr := "select user_mobile from user where user_mobile = ?"
	mobile := new(string)
	sys_err = global.DBMgr.DB.Get(mobile, sqlStr, createUser.Mobile)
	if sys_err != nil && sys_err != sql.ErrNoRows {
		// 查询数据库出错
		//return 0, errno.ErrServer.WithSysError(sys_err)
		return 0, status.Errorf(codes.Internal, "内部错误[Get] %s", sys_err.Error())
	}

	if sys_err == sql.ErrNoRows {
		// 用户不存在
	} else {
		//用户已存在

		return 0, status.Errorf(codes.NotFound, "用户已存在")
		//return 0, errno.ErrUserExists
	}

	// 插入数据
	// 生成加密密码与查询到的密码比较
	password := utils.Md5V(createUser.Password)

	fmt.Printf("手机号：%s  Md5V密码：%s\n", createUser.Mobile, password)

	insertSql := "INSERT user (user_mobile, user_password, nickname, add_time) VALUES (?, ?, ?, ?)"
	var sqlResult sql.Result = nil
	sqlResult, sys_err = global.DBMgr.DB.Exec(insertSql, createUser.Mobile, password, createUser.Nickname, time.Now().Format("2006-01-02 15:04:05"))
	if sys_err != nil {
		//return 0, errno.ErrServer.WithSysError(sys_err)
		return 0, status.Errorf(codes.Internal, "内部错误[Exec] %s", sys_err.Error())
	}

	var userID int64
	userID, sys_err = sqlResult.LastInsertId()
	if sys_err != nil {
		//return 0, errno.ErrServer.WithSysError(sys_err)
		return 0, status.Errorf(codes.Internal, "内部错误[LastInsertId] %s", sys_err.Error())
	}
	lastID = uint64(userID)

	return lastID, nil
}

func LoginUser(userLoginRequest *pb.UserLoginRequest) (user *pb.UserBaseInfoResponse, err errno.Error) {

	defer func() {
		if p := recover(); p != nil {
			user = nil
			errmsg := p.(string)
			err = errno.ErrServer.WithSysError(errors.New(errmsg))

			debug.PrintStack() //堆栈的异常调用信息
		}
	}()

	log.Println("[DAO层] [LoginUser] 用户登录 Begin")

	user = new(pb.UserBaseInfoResponse)
	//
	//originPassword := login.Password // 记录一下原始密码
	//生成加密密码与查询到的密码比较
	password := utils.Md5V(userLoginRequest.Password)

	fmt.Printf("手机号：%s  Md5V密码：%s\n", userLoginRequest.Mobile, password)

	sqlStr := "select id, nickname, role from user where user_mobile = ? and user_password = ?"
	sys_err := global.DBMgr.DB.Get(user, sqlStr, userLoginRequest.Mobile, password)
	if sys_err != nil && sys_err != sql.ErrNoRows {
		// 查询数据库出错
		fmt.Printf("查询失败：%s\n", sys_err.Error())
		return nil, errno.ErrServer.WithSysError(sys_err)
	}

	if sys_err == sql.ErrNoRows {
		// 用户不存在
		return nil, errno.ErrNameOrPasswordIncorrect
	}

	log.Println("[DAO层] [LoginUser] 用户登录 End", user)

	return user, nil
}

/*
// 事务操作
func updateTransaction() (err error) {
	tx, err := main_db.Begin()
	if err != nil {
		fmt.Printf("transaction begin failed, err:%v\n", err)
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			fmt.Printf("transaction rollback")
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			fmt.Printf("transaction commit")
			return
		}
	}()

	sqlStr1 := "UPDATE user SET age = ? WHERE id = ? "
	reuslt1, err := tx.Exec(sqlStr1, 18, 1)
	if err != nil {
		fmt.Printf("sql exec failed, err:%v\n", err)
		return err
	}
	rows1, err := reuslt1.RowsAffected()
	if err != nil {
		fmt.Printf("affected rows is 0")
		return
	}
	sqlStr2 := "UPDATE user SET age = ? WHERE id = ? "
	reuslt2, err := tx.Exec(sqlStr2, 19, 5)
	if err != nil {
		fmt.Printf("sql exec failed, err:%v\n", err)
		return err
	}
	rows2, err := reuslt2.RowsAffected()
	if err != nil {
		fmt.Printf("affected rows is 0\n")
		return
	}

	if rows1 > 0 && rows2 > 0 {
		fmt.Printf("update data success\n")
	}
	return
}
*/
