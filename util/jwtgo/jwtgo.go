package jwtgo

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("kephi520.") // 用于签名的密钥

// 声明一个结构体，用于存储 JWT 的声明信息
type Claims struct {
	UserPhone string `json:"userphone"`
	jwt.StandardClaims
}

// 生成 JWT
func GenerateJWT(userPhone string) (string, error) {
	expirationTime := time.Now().Add(31 * 24 * time.Hour) // 设置过期时间为31天后
	claims := &Claims{
		UserPhone: userPhone,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// 验证 JWT
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
