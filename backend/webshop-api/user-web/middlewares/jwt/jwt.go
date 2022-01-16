package jwt

import (
	"fmt"
	"time"
	"webshop-api/user-web/common/errno"
	"webshop-api/user-web/common/global"
	"webshop-api/user-web/models"

	"github.com/dgrijalva/jwt-go"
)

//Aeccse Token 过期时间
const APP_TOKEN_ACCESS_EXPIRE_DURATION = time.Minute * 10

//Refresh Token 过期时间
const APP_TOKEN_REFRESH_EXPIRE_DURATION = time.Hour * 24 * 7

var APP_TOKEN_SECRET_KEY = []byte("夏天夏天悄悄过去")

//func keyFunc(_ *jwt.Token) (i interface{}, err error) {
//	return APP_TOKEN_ACCESS_EXPIRE_DURATION, nil
//}

// WrapToken 生成token，并封装JWT Token
func WrapToken(userSession *models.UserSession) (aToken string, rToken string, err errno.Error) {

	if global.AppConfig.JWTConfig.SigningKey == "" {
		return "", "", errno.ErrJWTKeyEmpty
	}

	//系统错误
	var syserr error

	tokenSecretKey := []byte(global.AppConfig.JWTConfig.SigningKey)

	// 创建一个我们自己的声明
	userClaims := models.UserClaims{
		UserSession: userSession, // 自定义字段
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(global.AppConfig.JWTConfig.Expire)).Unix(), // 过期时间
			Issuer:    global.AppConfig.Name,                                                                 // 签发人(项目名称)
		},
	}

	// 加密并获得完整的编码后的字符串token
	// 使用指定的secret签名并获得完整的编码后的字符串token
	//Aeccse Token
	aToken, syserr = jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims).SignedString(tokenSecretKey)
	if syserr != nil {

		return "", "", errno.ErrTokenSign.WithSysError(syserr)
	}
	fmt.Println(aToken)

	//Refresh Token 不需要存任何自定义数据
	rToken, syserr = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(APP_TOKEN_REFRESH_EXPIRE_DURATION).Unix(),
		Issuer:    global.AppConfig.Name,
	}).SignedString(tokenSecretKey)
	if syserr != nil {
		return "", "", errno.ErrTokenSign.WithSysError(syserr)
	}

	return aToken, rToken, nil
}

// 解析token
func UnwrapToken(tokenStr string) (*models.UserClaims, errno.Error) {

	if global.AppConfig.JWTConfig.SigningKey == "" {
		return nil, errno.ErrJWTKeyEmpty
	}

	token, syserr := jwt.ParseWithClaims(tokenStr, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.AppConfig.JWTConfig.SigningKey), nil
	})
	if syserr != nil {

		if ve, ok := syserr.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				//return nil, errors.New("That's not even a token")
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				//return nil, errors.New("Token is expired")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				//return nil, errors.New("Token not active yet")
			} else {
				//return nil, errors.New("Couldn't handle this token:")
			}
		}

		return nil, errno.ErrTokenValidation.WithSysError(syserr)
	}

	if token != nil {

		if claims, ok := token.Claims.(*models.UserClaims); ok && token.Valid {
			return claims, nil
		}

		//if claims, ok := token.Claims.(*models.CustomClaims); ok && token.Valid {
		//	return claims, nil
		//}
		return nil, errno.ErrTokenValidation

	} else {
		return nil, errno.ErrTokenValidation
	}

	/*
		var claims = new(UserClaims)
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return APP_TOKEN_SECRET_KEY, nil
		})

		if token != nil { // 校验token
			if token.Valid {
				return claims, nil
			} else {
				//return nil, errors.New("invalid token")
			}
		}
	*/

	//fmt.Println("jwt.ParseWithClaims error:", err) // token is expired by 1h16m29s
	//return nil, err
}
