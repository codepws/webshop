// errno/code.go

package errno

/*
错误码规则
错误码需在 code.go 文件中定义。

错误码需为 > 0 的数，反之表示正确。

错误码为 5 位数
1				01				01
服务级错误码	模块级错误码	具体错误码

服务级别错误码：1 位数进行表示，比如 1 为系统级错误；2 为普通错误，通常是由用户非法操作引起。
模块级错误码：2 位数进行表示，比如 01 为用户模块；02 为订单模块。
具体错误码：2 位数进行表示，比如 01 为手机号不合法；02 为验证码输入错误。
*/

var (
	// OK
	Success = NewError(0, "Success") //Success

	// 系统错误, 前缀为 100   服务级错误码
	ErrServer    = NewError(10001, "内部服务器错误") //InternalServerError 服务异常，请联系管理员
	ErrParam     = NewError(10002, "参数有误")    // ErrBind 请求参数错误
	ErrSignParam = NewError(10003, "签名参数有误")
	// 签名JWT时发生错误
	ErrTokenSign       = NewError(10004, "签名JWT时发生错误")
	ErrTokenValidation = NewError(10005, "校验JWT时发生错误")
	ErrEncrypt         = NewError(10006, "加密用户密码时发生错误")

	// 模块级错误码 - 用户模块
	//ErrUserPhone   = NewError(20101, "用户手机号不合法")
	//ErrUserCaptcha = NewError(20102, "用户验证码有误")

	// 数据库错误, 前缀为 201
	ErrDatabase = NewError(20101, "数据库错误")               //&Errno{Code: 20100, Message: "数据库错误"}
	ErrFill     = NewError(20102, "从数据库填充 struct 时发生错误") //&Errno{Code: 20101, Message: "从数据库填充 struct 时发生错误"}

	// 认证错误, 前缀是 202
	ErrValidation   = NewError(20201, "验证失败")    //&Errno{Code: 20201, Message: "验证失败"}
	ErrTokenInvalid = NewError(20202, "Token无效") //&Errno{Code: 20202, Message: "jwt 是无效的"}

	// 用户错误, 前缀为 203
	ErrUserNotFound            = NewError(20301, "用户不存在")  //&Errno{Code: 20301, Message: "用户没找到"}
	ErrUserExists              = NewError(20302, "用户已经存在") //&Errno{Code: 20301, Message: "用户没找到"}
	ErrorPasswordWrong         = NewError(20303, "用户密码错误")
	ErrNameOrPasswordIncorrect = NewError(20304, "用户名或密码错误") //&Errno{Code: 20302, Message: "密码错误"}
)
