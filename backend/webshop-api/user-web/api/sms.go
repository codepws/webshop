package api

import (
	"fmt"
	"math/rand"
	"webshop-api/user-web/common"
	"webshop-api/user-web/common/global"
	"webshop-api/user-web/models/forms"

	"context"
	"strings"
	"time"

	//"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	//"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func GenerateSmsCode(witdh int) string {
	//生成width长度的短信验证码

	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < witdh; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

func SendSms(ctx *gin.Context) {
	zap.L().Debug("发送短信验证请求")

	sendSmsForm := forms.SendSmsForm{}
	if err := ctx.ShouldBind(&sendSmsForm); err != nil {
		HandleRequestValidatorError(ctx, err)
		return
	}

	//生成6位短信验证码
	smsCode := GenerateSmsCode(6)
	/*
		client, err := dysmsapi.NewClientWithAccessKey("cn-beijing", global.ServerConfig.AliSmsInfo.ApiKey, global.ServerConfig.AliSmsInfo.ApiSecrect)
		if err != nil {
			panic(err)
		}
		request := requests.NewCommonRequest()
		request.Method = "POST"
		request.Scheme = "https" // https | http
		request.Domain = "dysmsapi.aliyuncs.com"
		request.Version = "2017-05-25"
		request.ApiName = "SendSms"
		request.QueryParams["RegionId"] = "cn-beijing"
		request.QueryParams["PhoneNumbers"] = sendSmsForm.Mobile            //手机号
		request.QueryParams["SignName"] = global.AppConfig.Name             //阿里云验证过的项目名 自己设置
		request.QueryParams["TemplateCode"] = "SMS_181850725"               //阿里云的短信模板号 自己设置
		request.QueryParams["TemplateParam"] = "{\"code\":" + smsCode + "}" //短信模板中的验证码内容 自己生成   之前试过直接返回，但是失败，加上code成功。
		response, err := client.ProcessCommonRequest(request)
		fmt.Print(client.DoAction(request, response))
		if err != nil {
			fmt.Print(err.Error())
		}
	*/

	redis_addr := fmt.Sprintf("%s:%d", global.AppConfig.Caches[0].Host, global.AppConfig.Caches[0].Port)
	//将短信验证码保存起来 - redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_addr,
		Password: global.AppConfig.Caches[0].Password,
	})

	str, syserr := rdb.Ping(context.Background()).Result()
	if syserr != nil {
		fmt.Println("错误：", syserr)
	}
	fmt.Println(str)

	syserr = rdb.Set(context.Background(), sendSmsForm.Mobile, smsCode, time.Duration(global.AppConfig.AliSmsConfig.Expire)*time.Second).Err()
	if syserr != nil {
		zap.L().Debug(fmt.Sprintf("[Register] 【短信缓存】设置错误：%s", syserr.Error()))

	}

	zap.L().Debug(fmt.Sprintf("短信验证码缓存地址: %s", redis_addr))
	zap.L().Debug(fmt.Sprintf("手机号：%s, 类型：%d -> 验证码: %s", sendSmsForm.Mobile, sendSmsForm.Type, smsCode))

	//返回响应：短信验证码
	//common.ResponseSuccessWithMsg(ctx, "发送成功", nil)
	common.ResponseSuccess(ctx, gin.H{
		"expired_at": time.Now().Add(time.Second * time.Duration(global.AppConfig.AliSmsConfig.Expire)).Unix(),
	})

}
