package sqldb

import (
	"common/settings"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

//var main_db *sqlx.DB

type DBMgr struct {
	DB *sqlx.DB
}

func (dbmgr *DBMgr) Init(cfg *settings.DBs) (err error) {

	log.Println("初始化数据库")

	// "user:password@tcp(host:port)/dbname"
	//登录数据库
	//其中连接参数可以有如下几种形式：
	//user@unix(/path/to/socket)/dbname?charset=utf8
	//user:password@tcp(localhost:5555)/dbname?charset=utf8		//通常使用这种方式
	//user:password@/dbname
	//user:password@tcp([de:ad:be:ef::ca:fe]:80)/dbname
	main_dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		cfg.MasterDB.User, cfg.MasterDB.Password, cfg.MasterDB.Host, cfg.MasterDB.Port, cfg.MasterDB.Database)
	//商品数据库
	//shop_dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
	//	cfg.LoginDB.User, cfg.LoginDB.Password, cfg.LoginDB.Host, cfg.ShopDB.Port, cfg.LoginDB.Database)
	//其他数据库

	dbmgr.DB, err = sqlx.Connect(cfg.MasterDB.Type, main_dsn)
	if err != nil {
		return err
	}
	//设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	//连接池的实现关键在于SetMaxOpenConns和SetMaxIdleConns，其中：
	dbmgr.DB.SetMaxOpenConns(cfg.MasterDB.MaxOpenConns) //用于设置最大打开的连接数，默认值为0表示不限制。
	dbmgr.DB.SetMaxIdleConns(cfg.MasterDB.MaxIdleConns) //用于设置闲置的连接数。

	return nil
}

// Close 关闭MySQL连接
func (dbmgr *DBMgr) Close() {
	log.Println("关闭数据库")

	if dbmgr.DB != nil {
		_ = dbmgr.DB.Close()
	}

}
