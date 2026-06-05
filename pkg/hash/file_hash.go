// 路径: pkg/hash/file_hash.go
package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// CalculateFileHash 计算文件的 SHA256 哈希值
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	return CalculateHash(file)
}

// CalculateHash 计算 io.Reader 的 SHA256 哈希值（支持流式计算）
func CalculateHash(reader io.Reader) (string, error) {
	hash := sha256.New()
	
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("计算哈希失败: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CalculateStringHash 计算字符串的 SHA256 哈希值
func CalculateStringHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
