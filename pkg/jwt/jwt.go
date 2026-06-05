// 路径: pkg/jwt/jwt.go
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// TokenPair Token 对（Access Token 和 Refresh Token）
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access Token 过期时间（秒）
}

// GenerateToken 生成 Access Token
func GenerateToken(userID uint, username string, secret string, expireHours int) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("JWT 密钥不能为空")
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(expireHours) * time.Hour)

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("生成 Token 失败: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成 Refresh Token
func GenerateRefreshToken(userID uint, username string, secret string, expireHours int) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("JWT 密钥不能为空")
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(expireHours) * time.Hour)

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("生成 Refresh Token 失败: %w", err)
	}

	return tokenString, nil
}

// GenerateTokenPair 生成 Token 对（Access Token 和 Refresh Token）
func GenerateTokenPair(userID uint, username string, secret string, accessExpireHours, refreshExpireHours int) (*TokenPair, error) {
	accessToken, err := GenerateToken(userID, username, secret, accessExpireHours)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken(userID, username, secret, refreshExpireHours)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessExpireHours * 3600),
	}, nil
}

// ParseToken 解析和验证 Token
func ParseToken(tokenString string, secret string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("Token 不能为空")
	}

	if secret == "" {
		return nil, fmt.Errorf("JWT 密钥不能为空")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名算法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析 Token 失败: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的 Token")
}

// RefreshAccessToken 使用 Refresh Token 刷新 Access Token
func RefreshAccessToken(refreshToken string, secret string, accessExpireHours int) (string, error) {
	// 解析 Refresh Token
	claims, err := ParseToken(refreshToken, secret)
	if err != nil {
		return "", fmt.Errorf("Refresh Token 无效: %w", err)
	}

	// 生成新的 Access Token
	newAccessToken, err := GenerateToken(claims.UserID, claims.Username, secret, accessExpireHours)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}
