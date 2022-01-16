package api

import (
	"net/http"
	"webshop-api/user-web/common"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"go.uber.org/zap"
)

//内存缓存验证码
var captchaStore = base64Captcha.DefaultMemStore

//生成图形验证码
func GetCaptcha(ctx *gin.Context) {

	zap.L().Debug("获取图形验证请求")

	driver := base64Captcha.NewDriverDigit(80, 240, 6, 0.8, 100)
	cp := base64Captcha.NewCaptcha(driver, captchaStore)
	id, b64s, err := cp.Generate()
	if err != nil {
		zap.S().Errorf("生成图形验证码错误: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "生成图形验证码错误",
		})
		return
	}

	//返回响应：图形验证码
	common.ResponseSuccess(ctx, gin.H{
		"captchaId": id,
		"picPath":   b64s,
	})

}
