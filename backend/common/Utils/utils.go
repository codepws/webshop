package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

func PageStart(page int, limit int) int {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	start := (page - 1) * limit
	return start
}

//字串截取
func SubString(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

var md5code = "RaW#XhH2aVgo!IyL"

//用户注册MD5加密
func Md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str + md5code))
	return hex.EncodeToString(h.Sum(nil))
}

//生成随记字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

//视频文件名生成函数
func GetVideoName(uid string) string {
	//用户ID+精确到毫秒的时间戳
	h := md5.New()
	h.Write([]byte(uid + strconv.FormatInt(time.Now().UnixNano(), 10)))
	return hex.EncodeToString(h.Sum(nil))
}

func DataValidator(data interface{}) error {
	validate := validator.New()
	var sys_err = validate.Struct(data)
	if sys_err != nil {
		return sys_err
	}
	return nil
}
