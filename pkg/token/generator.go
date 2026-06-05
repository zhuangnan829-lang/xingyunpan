// 路径: pkg/token/generator.go
package token

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateShareToken 生成加密安全的分享令牌
// 使用 crypto/rand 生成 32 字节随机数，Base64 URL 编码
// 返回的令牌长度约为 43 字符（32 字节 * 4/3）
func GenerateShareToken() (string, error) {
	// 生成 32 字节随机数
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}

	// Base64 URL 编码（URL 安全，不包含 +/ 字符）
	token := base64.URLEncoding.EncodeToString(randomBytes)

	return token, nil
}

// GenerateShareTokenWithLength 生成指定字节长度的分享令牌
// length: 随机字节数（推荐 16-64）
// 返回的令牌长度约为 length * 4/3 字符
func GenerateShareTokenWithLength(length int) (string, error) {
	if length < 16 {
		return "", fmt.Errorf("令牌长度至少为 16 字节")
	}

	if length > 256 {
		return "", fmt.Errorf("令牌长度不能超过 256 字节")
	}

	// 生成指定长度的随机数
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}

	// Base64 URL 编码
	token := base64.URLEncoding.EncodeToString(randomBytes)

	return token, nil
}

// ValidateShareToken 验证分享令牌格式
// 检查令牌是否为有效的 Base64 URL 编码字符串
func ValidateShareToken(token string) error {
	if token == "" {
		return fmt.Errorf("令牌不能为空")
	}

	// 尝试解码
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return fmt.Errorf("令牌格式无效: %w", err)
	}

	// 检查解码后的长度（至少 16 字节）
	if len(decoded) < 16 {
		return fmt.Errorf("令牌长度不足")
	}

	return nil
}
