// 路径: pkg/crypto/password.go
package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 加密密码
// cost 参数控制加密强度，默认使用 bcrypt.DefaultCost (10)
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("密码不能为空")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword 验证密码是否匹配
// hashedPassword: 存储在数据库中的加密密码
// password: 用户输入的明文密码
func VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" || password == "" {
		return fmt.Errorf("密码不能为空")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("密码不正确")
		}
		return fmt.Errorf("密码验证失败: %w", err)
	}

	return nil
}

// HashPasswordWithCost 使用指定的 cost 加密密码
// cost 范围: 4-31，值越大越安全但越慢
// 推荐值: 10-14
func HashPasswordWithCost(password string, cost int) (string, error) {
	if password == "" {
		return "", fmt.Errorf("密码不能为空")
	}

	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return "", fmt.Errorf("cost 必须在 %d 到 %d 之间", bcrypt.MinCost, bcrypt.MaxCost)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}

	return string(hashedBytes), nil
}
