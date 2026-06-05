package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"xingyunpan-v2/pkg/logger"
)

const runnerHeartbeatKeyPrefix = "queue_runner:heartbeat"

type RunnerHeartbeat struct {
	Mode      string   `json:"mode"`
	Process   string   `json:"process"`
	Host      string   `json:"host"`
	PID       int      `json:"pid"`
	Queues    []string `json:"queues"`
	StartedAt string   `json:"started_at"`
	UpdatedAt string   `json:"updated_at"`
}

func StartRunnerHeartbeat(ctx context.Context, redisClient *goredis.Client, mode, process string, interval time.Duration) {
	if redisClient == nil {
		logger.Warn("queue runner heartbeat disabled: redis client is nil", zap.String("mode", mode), zap.String("process", process))
		return
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}

	host, _ := os.Hostname()
	if host == "" {
		host = "unknown"
	}
	startedAt := time.Now().Format(time.RFC3339)
	key := fmt.Sprintf("%s:%s:%s:%d", runnerHeartbeatKeyPrefix, mode, host, os.Getpid())

	write := func() {
		now := time.Now().Format(time.RFC3339)
		payload := RunnerHeartbeat{
			Mode:      mode,
			Process:   process,
			Host:      host,
			PID:       os.Getpid(),
			Queues:    DescribeImplementedQueues(),
			StartedAt: startedAt,
			UpdatedAt: now,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			logger.Warn("marshal queue runner heartbeat failed", zap.Error(err))
			return
		}
		if err := redisClient.Set(ctx, key, string(data), interval*3).Err(); err != nil && ctx.Err() == nil {
			logger.Warn("write queue runner heartbeat failed", zap.String("key", key), zap.Error(err))
		}
	}

	write()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				_ = redisClient.Del(context.Background(), key).Err()
				return
			case <-ticker.C:
				write()
			}
		}
	}()
}

func RunnerHeartbeatKeyPattern() string {
	return runnerHeartbeatKeyPrefix + ":*"
}
