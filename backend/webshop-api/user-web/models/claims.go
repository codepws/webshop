package models

import "github.com/dgrijalva/jwt-go"

type UserSession struct {
	UserID   uint64 `json:"userid"`
	NickName string `json:"nickName"`
	UserRole int    `json:"userrole"`
}

//用户鉴权 CustomClaims
type UserClaims struct {
	*UserSession
	jwt.StandardClaims
}
