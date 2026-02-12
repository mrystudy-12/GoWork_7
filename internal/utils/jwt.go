package utils

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var jwtKey = []byte("my_secret_key")

type Claims struct {
	ID                   int64  `json:"ID,omitempty"`
	Username             string `json:"Username,omitempty"`
	Role                 string `json:"Role,omitempty"`
	jwt.RegisteredClaims `json:"Jwt.RegisteredClaims"`
}

func GenerateToken(id int64, username, role string) (string, error) {
	expirationTime := time.Now().Add(30 * time.Minute)
	claims := Claims{
		ID:       id,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "myapp",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 1. 【核心安全优化】校验签名算法是否匹配，防止算法替换攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		// 2. 【错误处理优化】利用 v5 内置的错误判断返回更清晰的提示
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token 已过期")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, errors.New("token 签名无效")
		}
		return nil, err
	}

	// 3. 【逻辑优化】类型转换并返回。只要 err 为 nil，token.Valid 就为 true
	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}
