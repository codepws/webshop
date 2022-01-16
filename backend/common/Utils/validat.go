package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

//
// 验证邮箱: "^([//w-//.]+)@((//[[0-9]%7b1,3%7d//.[0-9]%7B1,3%7D//.[0-9]%7B1,3%7D//.)%7C(([//w-]+//.)+))([a-zA-Z]%7B2,4%7D%7C[0-9]%7B1,3%7D)(//]?)$"
// 验证输入密码条件(字符与数据同时出现)："[A-Za-z0-9]"
// 验证输入密码长度 (6-18位)："^//d{6,18}$"
// 验证输入邮政编号："^//d{6}$"
// 验证输入一年的12个月："^(0?[[1-9]|1[0-2])$"
// 验证输入一个月的31天："^((0?[1-9])|((1|2)[0-9])|30|31)$"
// 验证日期时间："^((((1[6-9]|[2-9]//d)//d{2})-(0?[13578]|1[02])-(0?[1-9]|[12]//d|3[01]))|(((1[6-9]|[2-9]//d)//d{2})-(0?[13456789]|1[012])-(0?[1-9]|[12]//d|30))|(((1[6-9]|[2-9]//d)//d{2})-0?2-(0?[1-9]|1//d|2[0-8]))|(((1[6-9]|[2-9]//d)(0[48]|[2468][048]|[13579][26])|((16|[2468][048]|[3579][26])00))-0?2-29-)) (20|21|22|23|[0-1]?//d):[0-5]?//d:[0-5]?//d$"
// 验证数字输入："^[0-9]*$"
// 验证非零的正整数："^//+?[1-9][0-9]*$"
// 验证大写字母："^[A-Z]+$"
// 验证小写字母："^[a-z]+$"
// 验证输入字母："^[A-Za-z]+$"
// 验证输入汉字："^[/u4e00-/u9fa5],{0,}$"
// 验证输入字符串："^.{8,}$"
//
//
//
//

/**
* 验证电话号码
* @return 如果是符合格式的字符串, 返回 <b>true </b>,否则为 <b>false </b>
 */
func IsTelephone(fl validator.FieldLevel) bool {
	ismatch, _ := regexp.MatchString("^(//d%7b3,4%7d-)/?//d{6,8}$", fl.Field().Interface().(string))
	return ismatch
}

/**
* 验证输入手机号码
* @param
* @return 如果是符合格式的字符串, 返回 <b>true </b>,否则为 <b>false </b>
 */
func isMobilePhone(fl validator.FieldLevel) bool {
	ismatch, _ := regexp.MatchString("^(((1[3,5,7,8][0-9]{1})|145|147|170|176|178|177)+\\d{8})$", fl.Field().Interface().(string))
	return ismatch
}

/**
* 验证输入身份证号
* @param
* @return 如果是符合格式的字符串, 返回 <b>true </b>,否则为 <b>false </b>
 */
func IsIDcard(fl validator.FieldLevel) bool {
	ismatch, _ := regexp.MatchString("^(\\d{15}$|^\\d{18}$|^\\d{17}(\\d|X|x))$", fl.Field().Interface().(string))
	return ismatch
}

/**
* 验证输入两位小数
* @param
* @return 如果是符合格式的字符串, 返回 <b>true </b>,否则为 <b>false </b>
 */
func IsDecimal(fl validator.FieldLevel) bool {
	ismatch, _ := regexp.MatchString("^[0-9]+(.[0-9]{2})?$", fl.Field().Interface().(string))
	return ismatch
}

func aaaaaaaaaaa(fl validator.FieldLevel) bool {
	ismatch, _ := regexp.MatchString("", fl.Field().Interface().(string))
	return ismatch
}
