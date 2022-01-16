package common

import (
	"net/http"
	"webshop-api/user-web/common/errno"

	"github.com/gin-gonic/gin"
)

func ResponseError(ctx *gin.Context, statusCode int, c errno.Error) {

	ctx.JSON(statusCode, c)
}

func ResponseErrorWithMsg(ctx *gin.Context, statusCode int, errcode int, errmsg string, data interface{}) {

	errno := errno.NewError(errcode, errmsg)
	errno.WithData(data)

	ctx.JSON(statusCode, errno)
}

// 成功响应，不带msg形参，默认success
func ResponseSuccess(ctx *gin.Context, data interface{}) {

	ok := errno.Success.WithData(data)
	ctx.JSON(http.StatusOK, ok)
}

// 成功响应，带msg形参
func ResponseSuccessWithMsg(ctx *gin.Context, msg string, data interface{}) {
	success := errno.NewError(0, msg).WithData(data)
	ctx.JSON(http.StatusOK, success)
}

// Http响应
func Response(ctx *gin.Context, httpStatus int, code int, msg string, data interface{}) {
	errno := errno.NewError(code, msg)
	errno.WithData(data)

	ctx.JSON(httpStatus, errno)
}

// 模板响应
func ResponseHtml(c *gin.Context, path string, data interface{}) {
	c.HTML(http.StatusOK, path, data)
}
