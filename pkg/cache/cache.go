// 路径: pkg/cache/cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheService 缓存服务
type CacheService struct {
	redis *redis.Client
}

type DirectoryStatsCacheEntry struct {
	ChildCount  int64 `json:"child_count"`
	FileCount   int64 `json:"file_count"`
	FolderCount int64 `json:"folder_count"`
	TotalSize   int64 `json:"total_size"`
}

type UserAuthStatusCacheEntry struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Enabled  bool   `json:"enabled"`
}

// NewCacheService 创建缓存服务实例
func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{redis: redis}
}

// CacheUserProfile 缓存用户信息（5 分钟 TTL）
func (c *CacheService) CacheUserProfile(ctx context.Context, userID uint, profile interface{}) error {
	if c == nil || c.redis == nil {
		return nil
	}
	key := fmt.Sprintf("user:profile:%d", userID)
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("序列化用户信息失败: %w", err)
	}
	return c.redis.Set(ctx, key, data, 5*time.Minute).Err()
}

// GetUserProfile 获取缓存的用户信息
func (c *CacheService) GetUserProfile(ctx context.Context, userID uint, dest interface{}) error {
	if c == nil || c.redis == nil {
		return redis.Nil
	}
	key := fmt.Sprintf("user:profile:%d", userID)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// InvalidateUserProfile 使用户信息缓存失效
func (c *CacheService) InvalidateUserProfile(ctx context.Context, userID uint) error {
	if c == nil || c.redis == nil {
		return nil
	}
	key := fmt.Sprintf("user:profile:%d", userID)
	return c.redis.Del(ctx, key).Err()
}

func (c *CacheService) CacheUserAuthStatus(ctx context.Context, userID uint, status UserAuthStatusCacheEntry) error {
	if c == nil || c.redis == nil {
		return nil
	}
	return c.Set(ctx, c.getUserAuthStatusKey(userID), status, 5*time.Minute)
}

func (c *CacheService) GetUserAuthStatus(ctx context.Context, userID uint, dest *UserAuthStatusCacheEntry) error {
	if c == nil || c.redis == nil {
		return redis.Nil
	}
	return c.Get(ctx, c.getUserAuthStatusKey(userID), dest)
}

func (c *CacheService) InvalidateUserSession(ctx context.Context, userID uint) error {
	if c == nil || c.redis == nil {
		return nil
	}
	return c.redis.Del(ctx, c.getUserAuthStatusKey(userID)).Err()
}

func (c *CacheService) InvalidateUserProfileAndSession(ctx context.Context, userID uint) error {
	if err := c.InvalidateUserProfile(ctx, userID); err != nil {
		return err
	}
	return c.InvalidateUserSession(ctx, userID)
}

func (c *CacheService) getUserAuthStatusKey(userID uint) string {
	return fmt.Sprintf("user:auth-status:%d", userID)
}

// CacheFileList 缓存文件列表（1 分钟 TTL）
func (c *CacheService) CacheFileList(ctx context.Context, userID uint, folderID *uint, files interface{}) error {
	return c.CacheFileListWithTTL(ctx, userID, folderID, 0, files, time.Minute)
}

func (c *CacheService) CacheFileListWithTTL(ctx context.Context, userID uint, folderID *uint, pageSize int, files interface{}, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	key := c.getFileListKey(userID, folderID, pageSize)
	data, err := json.Marshal(files)
	if err != nil {
		return fmt.Errorf("序列化文件列表失败: %w", err)
	}
	return c.redis.Set(ctx, key, data, ttl).Err()
}

// GetFileList 获取缓存的文件列表
func (c *CacheService) GetFileList(ctx context.Context, userID uint, folderID *uint, dest interface{}) error {
	return c.GetFileListWithPageSize(ctx, userID, folderID, 0, dest)
}

func (c *CacheService) GetFileListWithPageSize(ctx context.Context, userID uint, folderID *uint, pageSize int, dest interface{}) error {
	key := c.getFileListKey(userID, folderID, pageSize)
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// InvalidateFileList 使文件列表缓存失效
func (c *CacheService) InvalidateFileList(ctx context.Context, userID uint, folderID *uint) error {
	pattern := ""
	if folderID == nil {
		pattern = fmt.Sprintf("files:list:%d:root:*", userID)
	} else {
		pattern = fmt.Sprintf("files:list:%d:%d:*", userID, *folderID)
	}
	iter := c.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.redis.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// InvalidateUserAllFileLists 使用户所有文件列表缓存失效
func (c *CacheService) InvalidateUserAllFileLists(ctx context.Context, userID uint) error {
	// 使用模式匹配删除所有该用户的文件列表缓存
	pattern := fmt.Sprintf("files:list:%d:*", userID)
	iter := c.redis.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := c.redis.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	// 同时删除根目录缓存
	rootKey := fmt.Sprintf("files:list:%d:root", userID)
	return c.redis.Del(ctx, rootKey).Err()
}

// getFileListKey 生成文件列表缓存键
func (c *CacheService) getFileListKey(userID uint, folderID *uint, pageSize int) string {
	sizePart := "default"
	if pageSize > 0 {
		sizePart = fmt.Sprintf("size:%d", pageSize)
	}
	if folderID == nil {
		return fmt.Sprintf("files:list:%d:root:%s", userID, sizePart)
	}
	return fmt.Sprintf("files:list:%d:%d:%s", userID, *folderID, sizePart)
}

// Get 通用缓存获取方法
func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	if c == nil || c.redis == nil {
		return redis.Nil
	}
	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set 通用缓存设置方法
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c == nil || c.redis == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}
	return c.redis.Set(ctx, key, data, ttl).Err()
}

func (c *CacheService) CacheDirectoryStats(ctx context.Context, userID, folderID uint, entry DirectoryStatsCacheEntry, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return c.Set(ctx, c.getDirectoryStatsKey(userID, folderID), entry, ttl)
}

func (c *CacheService) GetDirectoryStats(ctx context.Context, userID, folderID uint, dest interface{}) error {
	return c.Get(ctx, c.getDirectoryStatsKey(userID, folderID), dest)
}

func (c *CacheService) InvalidateDirectoryStats(ctx context.Context, userID, folderID uint) error {
	if c == nil || c.redis == nil {
		return nil
	}
	return c.redis.Del(ctx, c.getDirectoryStatsKey(userID, folderID)).Err()
}

func (c *CacheService) InvalidateDirectoryStatsMany(ctx context.Context, userID uint, folderIDs ...uint) error {
	if c == nil || c.redis == nil || len(folderIDs) == 0 {
		return nil
	}

	keys := make([]string, 0, len(folderIDs))
	seen := make(map[uint]struct{}, len(folderIDs))
	for _, folderID := range folderIDs {
		if _, ok := seen[folderID]; ok {
			continue
		}
		seen[folderID] = struct{}{}
		keys = append(keys, c.getDirectoryStatsKey(userID, folderID))
	}

	return c.redis.Del(ctx, keys...).Err()
}

func (c *CacheService) getDirectoryStatsKey(userID, folderID uint) string {
	return fmt.Sprintf("files:dir-stats:%d:%d", userID, folderID)
}
