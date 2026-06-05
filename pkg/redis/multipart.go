// 路径: pkg/redis/multipart.go
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// MultipartRedis 分片上传 Redis 操作
type MultipartRedis struct {
	client *redis.Client
}

type PresignedURLCacheEntry struct {
	UploadID  string         `json:"upload_id"`
	URLs      map[int]string `json:"urls"`
	ExpiresAt int64          `json:"expires_at"`
}

// NewMultipartRedis 创建分片上传 Redis 实例
func NewMultipartRedis(client *redis.Client) *MultipartRedis {
	return &MultipartRedis{client: client}
}

// SetUploadMetadata 设置上传任务元数据
func (r *MultipartRedis) SetUploadMetadata(ctx context.Context, uploadID string, metadata map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("upload:%s", uploadID)
	return r.client.HSet(ctx, key, metadata).Err()
}

// GetUploadMetadata 获取上传任务元数据
func (r *MultipartRedis) GetUploadMetadata(ctx context.Context, uploadID string) (map[string]string, error) {
	key := fmt.Sprintf("upload:%s", uploadID)
	return r.client.HGetAll(ctx, key).Result()
}

// SetExpiration 设置过期时间
func (r *MultipartRedis) SetExpiration(ctx context.Context, uploadID string, expiration time.Duration) error {
	key := fmt.Sprintf("upload:%s", uploadID)
	return r.client.Expire(ctx, key, expiration).Err()
}

// AddCompletedChunk 添加已完成的分片
func (r *MultipartRedis) AddCompletedChunk(ctx context.Context, uploadID string, chunkNumber int) error {
	key := fmt.Sprintf("upload:%s:chunks", uploadID)
	return r.client.SAdd(ctx, key, chunkNumber).Err()
}

// GetCompletedChunks 获取已完成的分片列表
func (r *MultipartRedis) GetCompletedChunks(ctx context.Context, uploadID string) ([]string, error) {
	key := fmt.Sprintf("upload:%s:chunks", uploadID)
	return r.client.SMembers(ctx, key).Result()
}

// GetCompletedChunkCount 获取已完成的分片数量
func (r *MultipartRedis) GetCompletedChunkCount(ctx context.Context, uploadID string) (int64, error) {
	key := fmt.Sprintf("upload:%s:chunks", uploadID)
	return r.client.SCard(ctx, key).Result()
}

// IsChunkCompleted 检查分片是否已完成
func (r *MultipartRedis) IsChunkCompleted(ctx context.Context, uploadID string, chunkNumber int) (bool, error) {
	key := fmt.Sprintf("upload:%s:chunks", uploadID)
	return r.client.SIsMember(ctx, key, chunkNumber).Result()
}

// DeleteUploadData 删除上传任务数据
func (r *MultipartRedis) DeleteUploadData(ctx context.Context, uploadID string) error {
	metaKey := fmt.Sprintf("upload:%s", uploadID)
	chunksKey := fmt.Sprintf("upload:%s:chunks", uploadID)
	presignedURLKey := fmt.Sprintf("upload:%s:presigned_urls", uploadID)

	pipe := r.client.Pipeline()
	pipe.Del(ctx, metaKey)
	pipe.Del(ctx, chunksKey)
	pipe.Del(ctx, presignedURLKey)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *MultipartRedis) GetPresignedURLCache(ctx context.Context, uploadID string) (*PresignedURLCacheEntry, error) {
	key := fmt.Sprintf("upload:%s:presigned_urls", uploadID)
	raw, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var entry PresignedURLCacheEntry
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *MultipartRedis) SetPresignedURLCache(ctx context.Context, uploadID string, entry *PresignedURLCacheEntry, expiration time.Duration) error {
	if entry == nil {
		return nil
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("upload:%s:presigned_urls", uploadID)
	return r.client.Set(ctx, key, payload, expiration).Err()
}

func (r *MultipartRedis) DeletePresignedURLCache(ctx context.Context, uploadID string) error {
	key := fmt.Sprintf("upload:%s:presigned_urls", uploadID)
	return r.client.Del(ctx, key).Err()
}

func (r *MultipartRedis) ClearPresignedURLCaches(ctx context.Context) error {
	var cursor uint64

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "upload:*:presigned_urls", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
