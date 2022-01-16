package model

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// 用户信息
type User struct {
	UserID    uint64     `json:"user_id,string" db:"user_id"`
	VIP       uint       `json:"vip" db:"vip"`
	UserName  string     `json:"username" db:"username"`
	Email     string     `form:"email" json:"email" binding:"omitempty,email,lte=5"`
	Gender    uint8      `form:"gender" json:"gender" binding:"omitempty,oneof==0 1 2"`        //female male
	Addresses []*Address `form:"addresses" json:"addresses" binding:"omitempty,dive,required"` // a person can have a home and cottage...
}

// 登录请求请求参数 绑定模型
type SignInRequest struct {
	UserName string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// Address houses a users address information
type Address struct {
	Street string `validate:"required"`
	City   string `validate:"required"`
	Planet string `validate:"required"`
	Phone  string `validate:"required"`
}

// 用户注册请求参数
type SignUpRequest struct {
	UserName        string     `form:"username" json:"user_name" binding:"required,gte=6,lte=12"`
	Password        string     `form:"password" json:"password" binding:"required,gte=6,lte=12"`
	ConfirmPassword string     `form:"confirm_password" json:"confirm_password" binding:"required,gte=6,lte=12,eqfield=Password"`
	Email           string     `form:"email" json:"email" binding:"omitempty,email,lte=64"`
	Gender          uint8      `form:"gender" json:"gender" binding:"omitempty,oneof=0 1 2"`         //female male
	Addresses       []*Address `form:"addresses" json:"addresses" binding:"omitempty,dive,required"` // a person can have a home and cottage...
}

//tag中出现的dive的使用，dive一般用在slice、array、map、嵌套的struct验证中，
//作为分隔符表示进入里面一层的验证规则

/*
// 用户注册请求参数
type SignUpRequest struct {
	UserName        string `form:"user_name" json:"user_name" binding:"required"`
	Password        string `form:"password" json:"password" binding:"required"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password" binding:"required"`
	Email           string `form:"email" json:"email" binding:""`
}
*/

/*
// 方式一：自定义核验错误提示
func (r *SignUpRequest) UnmarshalJSON(data []byte) (err error) {

	required := struct {
		UserName        string `json:"username" db:"username"`
		Password        string `json:"password" db:"password"`
		ConfirmPassword string `json:"confirm_password" db:"confirm_password"`
		Email           string `json:"email" db:"email"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	} else if len(required.UserName) == 0 {
		err = errors.New("缺少必填字段username")
	} else if len(required.Password) == 0 {
		err = errors.New("缺少必填字段password")
	} else if required.Password != required.ConfirmPassword {
		err = errors.New("两次密码不一致")
	} else {
		r.UserName = required.UserName
		r.Password = required.Password
		r.ConfirmPassword = required.ConfirmPassword
		r.Email = required.Email
	}
	return
}
*/

// 方式二：绑定模型获取验证错误的方法
func (signup *SignUpRequest) GetError(err validator.ValidationErrors) string {

	var jsonMap = make(map[string]string, len(err))
	for idx, item := range err {
		fmt.Printf("%d'th: Tag=%s   ActualTag=%s    Namespace=%s        StructNamespace=%s        Field=%s        StructField=%s        Param=%s        Kind=%v    Type=%v    \n",
			idx, item.Tag(), item.ActualTag(), item.Namespace(), item.StructNamespace(), item.Field(), item.StructField(), item.Param(), item.Kind(), item.Type())
		fmt.Printf("Value=%v    Error=%s\n", item.Value(), item.Error())

		key := item.StructField()
		value := item.Tag()
		jsonMap[key] = value
	}

	//mapInstances := []map[string]interface{}{}
	//instance_1 := map[string]interface{}{"name": "John", "age": 10}
	//instance_2 := map[string]interface{}{"name": "Alex", "age": 12}
	//mapInstances = append(mapInstances, instance_1, instance_2)

	//map转Json
	jsonStr, errjson := json.Marshal(jsonMap)

	if errjson != nil {
		//fmt.Println("MapToJsonDemo err: ", err)

		return "参数错误"
	}
	//fmt.Println(string(jsonStr))

	return string(jsonStr)

	/*
		// 这里的 "LoginRequest.Mobile" 索引对应的是模型的名称和字段
		if val, exist := err["SignUpRequest.UserName"]; exist {
			if val.Field == "UserName" {
				switch val.Tag {
				case "required":
					return "请输入用户名"
				}
			}
		}
		if val, exist := err["SignUpRequest.Password"]; exist {
			if val.Field == "Password" {
				switch val.Tag {
				case "required":
					return "请输入密码"
				}
			}
		}
	*/

	//return "参数错误"
}

type UserInfo struct {
	Id       uint64 `json:"id,string" db:"id"`
	Name     string `json:"name" db:"name"`
	Password string `form:"password" json:"password"`
	AddTime  string `form:"add_time" json:"add_time" binding:"omitempty"`
	Avatar   string `json:"avatar" binding:"omitempty"` //female male
}

// 登录请求请求参数 绑定模型
type LoginRequest struct {
	Mobile    string `form:"mobile" json:"mobile" binding:"required,len=11,isMobile"`
	Password  string `form:"password" json:"password" binding:"required,min=6,max=12"`
	ReturnUrl string `form:"ReturnUrl" json:"ReturnUrl" binding:"omitempty"`
}

func (login *LoginRequest) GetError(err validator.ValidationErrors) string {

	for _, item := range err {

		fmt.Printf("%d'th: Tag=%s   ActualTag=%s    Namespace=%s        StructNamespace=%s        Field=%s        StructField=%s        Param=%s        Kind=%v    Type=%v    \n",
			0, item.Tag(), item.ActualTag(), item.Namespace(), item.StructNamespace(), item.Field(), item.StructField(), item.Param(), item.Kind(), item.Type())
		fmt.Printf("Value=%v    Error=%s\n", item.Value(), item.Error())

		key := item.StructField()
		value := item.Tag()

		switch {
		case key == "Mobile":
			switch {
			case value == "required":
				return "请输入手机号！"
			case value == "len":
				return "手机号长度不正确！"
			case value == "isMobile":
				return "手机号格式不正确！"
			default:
				return "手机号无效！"
			}
		case key == "Password":
			switch {
			case value == "required":
				return "请输入密码！"
			case value == "min", value == "max":
				return "密码长度为6-12位！"
			default:
				return "密码无效！"
			}
		default:
			return "参数错误！"
		}
	}

	return err.Error()
}

// 用户注册请求参数
type CreateUserParam struct {
	Mobile   string `form:"mobile" json:"mobile" binding:"required,len=11,isMobile"`
	Password string `form:"password" json:"password" binding:"required,min=6,max=12"`
	Nickname string `form:"nickname" json:"nickname" binding:"required,max=16"`
}

func (register *CreateUserParam) GetError(err validator.ValidationErrors) string {

	for _, item := range err {

		key := item.StructField()
		value := item.Tag()
		switch {
		case key == "Mobile":
			switch {
			case value == "required":
				return "请输入手机号！"
			case value == "len":
				return "手机号长度不正确！"
			case value == "isMobile":
				return "手机号格式不正确！"
			default:
				return "手机号无效！"
			}
		case key == "Password":
			switch {
			case value == "required":
				return "请输入密码！"
			case value == "min", value == "max":
				return "密码长度为6-12位！"
			default:
				return "密码无效！"
			}
		default:
			return "参数错误！"
		}
	}

	return err.Error()
}
