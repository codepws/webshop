package routers

import (
	"net/http"
	"webshop-api/user-web/api"

	"github.com/gin-gonic/gin"
)

type RouterOption func(*gin.Engine)

var routerOptions = []RouterOption{}

// 设计模式–函数式选项模式
// 注册app的路由配置
func include(opts ...RouterOption) {
	routerOptions = append(routerOptions, opts...)
}

//初始化
func SetupRouter() *gin.Engine {

	////////////////////////////////////////
	// 加载多个APP的路由配置
	//include(login.Routers, channel.Routers, comment.Routers)

	////////////////////////////////////////
	router := gin.Default()
	//r := gin.New()
	//r.Use(logger.GinLogger(), logger.GinRecovery(true))

	//1.首位多余元素会被删除(../ or //);
	//2.然后路由会对新的路径进行不区分大小写的查找;
	//3.如果能正常找到对应的handler，路由就会重定向到正确的handler上并返回301或者307.(比如: 用户访问/FOO 和 /..//Foo可能会被重定向到/foo这个路由上)
	router.RedirectFixedPath = true

	//for _, opt := range routerOptions {
	//	opt(r)
	//}

	//配置跨域
	//router.Use(middlewares.Cors())

	//Web服务健康检查
	initHealthCheck(router)

	//
	apiGroup := router.Group("/u/v1")
	initUserRouter(apiGroup)
	initBaseRouter(apiGroup)

	router.GET("/test", api.UnaryEcho)

	return router
}

// Web服务健康检查
func initHealthCheck(router *gin.Engine) {

	router.GET("/actuator/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, nil)
	})

}

func initUserRouter(router *gin.RouterGroup) {
	userGroup := router.Group("user")
	{

		userGroup.POST("mobile_login", api.UserLogin)
		userGroup.POST("register", api.UserRegister)

		//userGroup.GET("list", middlewares.JWTAuthMiddleware(), api.GetUserList) //middlewares.IsAdminAuth()  api.GetUserList
		//userGroup.POST("pwd_login", api.PasswordLogin)
	}
}

func initBaseRouter(router *gin.RouterGroup) {
	baseGroup := router.Group("base")
	{
		baseGroup.GET("captcha", api.GetCaptcha) //获取图形验证码
		baseGroup.POST("send_sms", api.SendSms)  //发送短信验证码
	}

}
