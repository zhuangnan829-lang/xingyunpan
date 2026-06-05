package service

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/queue"
)

type QueueRunnerHeartbeatPayload struct {
	Mode      string   `json:"mode"`
	Process   string   `json:"process"`
	Host      string   `json:"host"`
	PID       int      `json:"pid"`
	Queues    []string `json:"queues"`
	StartedAt string   `json:"started_at"`
	UpdatedAt string   `json:"updated_at"`
}

type QueueRuntimeStatusPayload struct {
	EmbeddedRunnerEnabled bool                          `json:"embedded_runner_enabled"`
	WorkerEnabled         bool                          `json:"worker_enabled"`
	HeartbeatAvailable    bool                          `json:"heartbeat_available"`
	IndependentWorkerSeen bool                          `json:"independent_worker_seen"`
	EmbeddedRunnerSeen    bool                          `json:"embedded_runner_seen"`
	RunnerCount           int                           `json:"runner_count"`
	Runners               []QueueRunnerHeartbeatPayload `json:"runners"`
	Message               string                        `json:"message"`
}

type QueueRuntimeService interface {
	GetStatus() (*QueueRuntimeStatusPayload, error)
}

type queueRuntimeService struct {
	redis *goredis.Client
}

func NewQueueRuntimeService(redisClient *goredis.Client) QueueRuntimeService {
	return &queueRuntimeService{redis: redisClient}
}

func (s *queueRuntimeService) GetStatus() (*QueueRuntimeStatusPayload, error) {
	payload := &QueueRuntimeStatusPayload{
		EmbeddedRunnerEnabled: config.Config != nil && config.Config.Queue.IsEmbeddedRunnerEnabled(),
		WorkerEnabled:         config.Config == nil || config.Config.Worker.IsEnabled(),
		Runners:               []QueueRunnerHeartbeatPayload{},
	}

	if s.redis == nil {
		payload.Message = "Redis 未初始化，无法检测 runner 心跳；请结合配置和进程列表判断。"
		return payload, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	keys, err := s.redis.Keys(ctx, queue.RunnerHeartbeatKeyPattern()).Result()
	if err != nil {
		payload.Message = "读取 runner 心跳失败：" + err.Error()
		return payload, nil
	}

	payload.HeartbeatAvailable = true
	for _, key := range keys {
		raw, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var item QueueRunnerHeartbeatPayload
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			continue
		}

		payload.Runners = append(payload.Runners, item)
		switch item.Mode {
		case "worker":
			payload.IndependentWorkerSeen = true
		case "server_embedded":
			payload.EmbeddedRunnerSeen = true
		}
	}

	payload.RunnerCount = len(payload.Runners)
	if payload.RunnerCount == 0 {
		payload.Message = "未检测到 runner 心跳。"
	} else if payload.IndependentWorkerSeen && payload.EmbeddedRunnerSeen {
		payload.Message = "同时检测到 server 内置 runner 与独立 worker；worker_num 会按 runner 进程叠加。"
	} else if payload.IndependentWorkerSeen {
		payload.Message = "检测到独立 worker 心跳。"
	} else if payload.EmbeddedRunnerSeen {
		payload.Message = "检测到 server 内置 runner 心跳。"
	}

	return payload, nil
}
