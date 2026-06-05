package service

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/go-redis/redis/v8"

	"xingyunpan-v2/internal/config"
)

func TestQueueRuntimeServiceReportsWorkerHeartbeat(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer redisServer.Close()

	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	oldConfig := config.Config
	embedded := false
	worker := true
	config.Config = &config.AppConfig{
		Queue:  config.QueueConfig{EmbeddedRunnerEnabled: &embedded},
		Worker: config.WorkerConfig{Enabled: &worker},
	}
	defer func() { config.Config = oldConfig }()

	ctx := context.Background()
	err = redisClient.Set(ctx, "queue_runner:heartbeat:worker:test-host:1234", `{
		"mode":"worker",
		"process":"cmd/worker",
		"host":"test-host",
		"pid":1234,
		"queues":["metadata","io"],
		"started_at":"2026-06-05T10:00:00Z",
		"updated_at":"2026-06-05T10:00:05Z"
	}`, time.Minute).Err()
	if err != nil {
		t.Fatalf("write heartbeat: %v", err)
	}

	status, err := NewQueueRuntimeService(redisClient).GetStatus()
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	if status.EmbeddedRunnerEnabled {
		t.Fatalf("expected embedded runner disabled")
	}
	if !status.WorkerEnabled {
		t.Fatalf("expected worker enabled")
	}
	if !status.HeartbeatAvailable {
		t.Fatalf("expected heartbeat available")
	}
	if !status.IndependentWorkerSeen {
		t.Fatalf("expected independent worker heartbeat")
	}
	if status.EmbeddedRunnerSeen {
		t.Fatalf("did not expect embedded runner heartbeat")
	}
	if status.RunnerCount != 1 {
		t.Fatalf("expected 1 runner, got %d", status.RunnerCount)
	}
}

func TestQueueRuntimeServiceReportsCombinedRunnerMode(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer redisServer.Close()

	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	oldConfig := config.Config
	embedded := true
	config.Config = &config.AppConfig{
		Queue: config.QueueConfig{EmbeddedRunnerEnabled: &embedded},
	}
	defer func() { config.Config = oldConfig }()

	ctx := context.Background()
	for key, payload := range map[string]string{
		"queue_runner:heartbeat:worker:test-host:1234":          `{"mode":"worker","process":"cmd/worker","host":"test-host","pid":1234,"queues":["metadata"],"started_at":"2026-06-05T10:00:00Z","updated_at":"2026-06-05T10:00:05Z"}`,
		"queue_runner:heartbeat:server_embedded:test-host:5678": `{"mode":"server_embedded","process":"cmd/server","host":"test-host","pid":5678,"queues":["metadata"],"started_at":"2026-06-05T10:00:00Z","updated_at":"2026-06-05T10:00:05Z"}`,
	} {
		if err := redisClient.Set(ctx, key, payload, time.Minute).Err(); err != nil {
			t.Fatalf("write heartbeat %s: %v", key, err)
		}
	}

	status, err := NewQueueRuntimeService(redisClient).GetStatus()
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	if !status.EmbeddedRunnerEnabled {
		t.Fatalf("expected embedded runner enabled")
	}
	if !status.IndependentWorkerSeen || !status.EmbeddedRunnerSeen {
		t.Fatalf("expected both runner modes, got worker=%v embedded=%v", status.IndependentWorkerSeen, status.EmbeddedRunnerSeen)
	}
	if status.RunnerCount != 2 {
		t.Fatalf("expected 2 runners, got %d", status.RunnerCount)
	}
}
