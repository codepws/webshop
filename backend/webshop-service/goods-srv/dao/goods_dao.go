package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"webshop-service/goods-srv/common/global"
	"webshop-service/goods-srv/model"

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

//
func GetAllCategorysList() (categorylist []*model.CategoryInfo, err error) {

	//我们在调用recover的延迟函数中以最合理的方式响应该异常：
	//1. 打印堆栈的异常调用信息和关键的业务信息，以便这些问题保留可见；
	//2. 将异常转换为错误，以便调用者让程序恢复到健康状态并继续安全运行。
	defer func() {
		if p := recover(); p != nil {
			errmsg := p.(string)
			err = errors.New(errmsg)

			debug.PrintStack() //堆栈的异常调用信息
		}
	}()

	var sys_err error = nil
	//启动事务

	//
	sqlStr := "select id, name, parent_category_id,level, is_tab from category where is_deleted = 0" //非删除的所有商品类别
	sys_err = global.DBMgr.DB.Select(&categorylist, sqlStr)
	if sys_err != nil && sys_err != sql.ErrNoRows {
		// 查询数据库出错
		//return 0, errno.ErrServer.WithSysError(sys_err)
		return nil, status.Errorf(codes.Internal, "内部错误[DB.Select] %s", sys_err.Error())
	}
	//
	return categorylist, nil
}

func GetGoodsListByIds(goodsIds []uint32) (goodslist []*model.GoodsInfo, err error) {

	defer func() {
		if p := recover(); p != nil {
			errmsg := p.(string)
			err = errors.New(errmsg)

			debug.PrintStack() //堆栈的异常调用信息
		}
	}()

	var sys_err error = nil
	//启动事务
	count := len(goodsIds)
	goodsIdsString := make([]string, count)
	for i := 0; i < count; i++ {
		goodsIdsString[i] = strconv.FormatUint(uint64(goodsIds[i]), 10)
	}
	ids := strings.Join(goodsIdsString, ",")
	sqlStr := fmt.Sprintf("select * from goods where id IN(%s) and is_deleted = 0", ids) //IN是比较等不等	性能高
	//sqlStr := fmt.Sprintf("select * from goods where FIND_IN_SET(id,'%s') and is_deleted = 0", ids) //FIND_IN_SET函数用来比较是不是包含	性能不如 IN
	sys_err = global.DBMgr.DB.Select(&goodslist, sqlStr)
	if sys_err != nil && sys_err != sql.ErrNoRows {
		// 查询数据库出错
		//return 0, errno.ErrServer.WithSysError(sys_err)
		return nil, status.Errorf(codes.Internal, "内部错误[DB.Select] %s", sys_err.Error())
	}
	//
	return goodslist, nil
}

func GetGoodsDetailById(goodsId uint32) (goods *model.GoodsInfo, err error) {

	defer func() {
		if p := recover(); p != nil {
			errmsg := p.(string)
			err = errors.New(errmsg)

			debug.PrintStack() //堆栈的异常调用信息
		}
	}()

	var sys_err error = nil

	//
	goods = new(model.GoodsInfo)
	//
	sqlStr := "select * from goods where id=? and is_deleted=0"
	sys_err = global.DBMgr.DB.Get(goods, sqlStr, goodsId)
	if sys_err != nil && sys_err != sql.ErrNoRows {
		// 查询数据库出错
		//return 0, errno.ErrServer.WithSysError(sys_err)
		return nil, status.Errorf(codes.Internal, "内部错误[DB.Get] %s", sys_err.Error())
	}
	//
	return goods, nil
}
