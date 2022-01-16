package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"webshop-api/user-web/common"
	"webshop-api/user-web/common/errno"
	"webshop-api/user-web/common/global"
	"webshop-api/user-web/middlewares/jwt"
	"webshop-api/user-web/models"
	"webshop-api/user-web/models/forms"
	"webshop-api/user-web/proto"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func HandleRequestValidatorError(c *gin.Context, err error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		//非验证错误
		common.ResponseError(c, http.StatusBadRequest, errno.ErrParam)
	} else {
		//
		common.ResponseErrorWithMsg(c, http.StatusBadRequest, -1, errs.Error(), nil)
	}

}

func GetUserList(ctx *gin.Context) {
	//拨号连接用户grpc服务器 跨域的问题 - 后端解决 也可以前端来解决
	claims, _ := ctx.Get("KeyCtxContextUserClaims")
	currentUser := claims.(*models.UserClaims)

	zap.S().Infof("访问用户: %d", currentUser.UserID)
	//生成grpc的client并调用接口

}

//用户注册API
func UserRegister(c *gin.Context) {

	log.Println("[Register] 【用户注册】")

	//Step1: 请求参数验证
	registerForm := forms.RegisterForm{}
	if err := c.ShouldBind(&registerForm); err != nil {
		HandleRequestValidatorError(c, err)
		return
	}

	//Step2: 短信验证码验证
	redisAddr := fmt.Sprintf("%s:%d", global.AppConfig.Caches[0].Host, global.AppConfig.Caches[0].Port)

	zap.L().Debug(fmt.Sprintf("[Register] 【短信缓存】Redis地址：%s  passowrd:%s", redisAddr, global.AppConfig.Caches[0].Password))

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: global.AppConfig.Caches[0].Password,
	})

	str, syserr := rdb.Ping(context.Background()).Result()
	if syserr != nil {
		zap.L().Debug(fmt.Sprintf("Redis连接错误：%s", syserr))
	} else {
		zap.L().Debug(fmt.Sprintf("Redis连接成功：%s", str))
	}

	smsCode, syserr := rdb.Get(context.Background(), registerForm.Mobile).Result()
	if syserr == redis.Nil {
		zap.L().Debug(fmt.Sprintf("[Register] 【短信缓存】验证码不存在：%s", syserr.Error())) //"key2 does not exist"
		common.ResponseError(c, http.StatusBadRequest, errno.ErrVerificationCodeExpired)
		return
	} else if syserr != nil {
		zap.L().Error(fmt.Sprintf("[Register] 【短信缓存】获取错误：%s", syserr.Error()))

		common.ResponseError(c, http.StatusBadRequest, errno.ErrVerificationCodeInvalid)
		return
	}

	zap.L().Debug(fmt.Sprintf("[Register] 【短信验证码验证】用户手机：%s 用户验证码：%s - 缓存验证码:%s", registerForm.Mobile, registerForm.Code, smsCode))
	if smsCode == "" {
		common.ResponseError(c, http.StatusBadRequest, errno.ErrVerificationCodeExpired)
		return
	} else if smsCode != registerForm.Code {
		common.ResponseError(c, http.StatusBadRequest, errno.ErrVerificationCodeInvalid)
		return
	}

	//Step3: 请求 user-srv 服务
	user, err := global.UserSrvGrpcClient.RegisterUser(context.Background(), &proto.UserRegisterRequest{
		Nickname: registerForm.Mobile,
		Password: registerForm.Password,
		Mobile:   registerForm.Mobile,
	})
	if err != nil {
		zap.S().Errorf("[Register] 查询 【新建用户】 失败: %s", err.Error())
		common.ResponseError(c, http.StatusInternalServerError, errno.ErrUserExists)
		return
	}

	var userId uint64 = user.Id
	var nickName string = user.Nickname

	userSession := models.UserSession{UserID: userId, NickName: nickName}
	aToken, rToken, err11 := jwt.WrapToken(&userSession)
	if err != nil {
		zap.L().Error(fmt.Sprintf("[Register] 【封装JWT】 生成错误：%s", err11.ToString()))

		common.ResponseError(c, http.StatusInternalServerError, err11)
		return
	}

	//
	syserr = rdb.Del(context.Background(), registerForm.Mobile).Err() // intCmd不为空: del 13800138000: 1
	if syserr != nil {
		zap.L().Warn(fmt.Sprintf("[Register] 【短信缓存】 删除错误：%s", syserr.Error()))
	}

	common.ResponseSuccess(c, gin.H{
		"id":         userSession.UserID,
		"nick_name":  userSession.NickName,
		"aToken":     aToken,
		"rToken":     rToken,
		"expired_at": time.Now().Add(time.Second * time.Duration(global.AppConfig.JWTConfig.Expire)).Unix(),
	})

}

// 用户登录API
func UserLogin(c *gin.Context) {

	log.Println("[Login] 【用户登录】")

	//Step1: 请求参数验证
	loginForm := forms.LoginForm{}
	if err := c.ShouldBind(&loginForm); err != nil {
		HandleRequestValidatorError(c, err)
		return
	}

	//Step2: 图片验证码验证
	if !captchaStore.Verify(loginForm.CaptchaId, loginForm.Captcha, false) {
		c.JSON(http.StatusBadRequest, gin.H{
			"captcha": "验证码错误",
		})
		return
	}

	//Step3: 请求 user-srv 服务
	user, err := global.UserSrvGrpcClient.LoginUser(context.Background(), &proto.UserLoginRequest{
		Mobile:   loginForm.Mobile,
		Password: loginForm.Password,
	})
	if err != nil {
		zap.S().Errorf("[UserLogin] 查询 【用户登录】 失败: %s", err.Error())
		common.ResponseError(c, http.StatusInternalServerError, errno.ErrUserExists)
		return
	}

	var userId uint64 = user.Id
	var nickName string = user.Nickname

	userSession := models.UserSession{UserID: userId, NickName: nickName}
	aToken, rToken, err11 := jwt.WrapToken(&userSession)
	if err != nil {
		zap.L().Error(fmt.Sprintf("[UserLogin] 【封装JWT】 生成错误：%s", err11.ToString()))

		common.ResponseError(c, http.StatusInternalServerError, err11)
		return
	}

	common.ResponseSuccess(c, gin.H{
		"id":         userSession.UserID,
		"nick_name":  userSession.NickName,
		"token":      aToken,
		"rToken":     rToken,
		"expired_at": time.Now().Add(time.Second * time.Duration(global.AppConfig.JWTConfig.Expire)).Unix(),
	})

}

//
func UnaryEcho(ctx *gin.Context) {

	log.Println("[UnaryEcho] 【普通一元方法】")

	//Step1: 请求参数验证
	name := ctx.DefaultQuery("name", "unknown")

	//Step3: 请求 user-srv 服务 EchoResponse
	echo, err := global.UserSrvGrpcClient.UnaryEcho(context.Background(), &proto.EchoRequest{
		Name: name,
	})
	if err != nil {
		zap.S().Errorf("[UnaryEcho] 调用 【普通一元方法】 失败: %s", err.Error())
		common.ResponseError(ctx, http.StatusInternalServerError, errno.ErrUserExists)
		return
	}

	common.ResponseSuccess(ctx, gin.H{
		"message": echo.Message,
	})

}
