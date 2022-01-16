package forms

// 用户注册
type RegisterForm struct {
	Mobile   string `form:"mobile" json:"mobile" binding:"required,isMobilePhone"`    //手机号码格式有规范可寻， 自定义validator
	Password string `form:"password" json:"password" binding:"required,min=3,max=20"` //密码
	Code     string `form:"code" json:"code" binding:"required,min=6,max=6"`          //短信验证码
}

// 用户登录：手机号 + 密码
type LoginForm struct {
	Mobile    string `form:"mobile" json:"mobile" binding:"required,isMobilePhone"`    //手机号码格式有规范可寻， 自定义validator
	Password  string `form:"password" json:"password" binding:"required,min=3,max=20"` //密码
	Captcha   string `form:"captcha" json:"captcha" binding:"required,len=6"`          //验证码
	CaptchaId string `form:"captcha_id" json:"captcha_id" binding:"required"`          //验证码ID
}
